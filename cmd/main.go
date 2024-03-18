/*
Copyright 2024 Statistics Canada.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"crypto/tls"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	networkingstatcangccav1alpha1 "statcan.gc.ca/cidr-allocator/api/v1alpha1"
	"statcan.gc.ca/cidr-allocator/internal/controller"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// lookupEnvOrDefault is a short helper function which provides a way to lookup environment variables and return a default if nothing is set
func lookupEnvOrDefault(key string, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}

var (
	// leaderElectionID represents the lock/lease identity used in leader-election
	leaderElectionID string
	// metricsAddr represents the address that the metric endpoint binds to
	metricsAddr string
	// probeAddr Is the address that the health probe endpoint binds to
	probeAddr string
	// enableLeaderElection specifies whether or not leader election should be used for the controller manager
	enableLeaderElection bool
	// debugLogging specifies whether to enable Development (Debug) logging for ZAP. Otherwise, Zap Production logging will be used
	debugLogging bool
	// secureMetrics specifies whether the metrics endpoint is served over https
	secureMetrics bool
	// enableHTTP2 specifies that HTTP/2 will be enabled for the metrics and webhook servers (if exists)
	enableHTTP2 bool
)

func init() {
	// Configure CLI arguments
	flag.StringVar(
		&metricsAddr,
		"metrics-bind-address",
		lookupEnvOrDefault("METRICS_BIND_ADDR", ":9003"),
		"The address the metric endpoint binds to.",
	)
	flag.StringVar(
		&probeAddr,
		"health-probe-bind-address",
		lookupEnvOrDefault("HEALTH_PROBE_BIND_ADDR", ":8081"),
		"The address the probe endpoint binds to.",
	)
	flag.StringVar(
		&leaderElectionID,
		"leader-election-id",
		lookupEnvOrDefault("LEADER_ELECTION_ID", "cidr-allocator-leader.networking.statcan.gc.ca"),
		"The identity to use for leader-election",
	)
	flag.BoolVar(
		&enableLeaderElection,
		"leader-elect",
		false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.",
	)
	flag.BoolVar(
		&debugLogging,
		"debug",
		false,
		"Enable development logging",
	)
	flag.BoolVar(
		&secureMetrics,
		"metrics-secure",
		false,
		"If set, the metrics endpoint is served over https",
	)
	flag.BoolVar(
		&enableHTTP2,
		"enable-http2",
		false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers (if exists)",
	)

	opts := zap.Options{
		Development: debugLogging,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Configure ZAP Logger
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(networkingstatcangccav1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancelation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.NodeCIDRAllocationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("NodeCIDRAllocationController"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeCIDRAllocation")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	managerContext := ctrl.SetupSignalHandler()

	if err := mgr.GetFieldIndexer().IndexField(managerContext, &corev1.Node{}, "spec.podCIDR", func(o client.Object) []string {
		node := o.(*corev1.Node)
		return []string{node.Spec.PodCIDR}
	}); err != nil {
		setupLog.Error(err, "unable to setup field indexer for Node PodCIDRs")
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(managerContext); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
