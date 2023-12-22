/*
MIT License

Copyright (c) His Majesty the King in Right of Canada, as represented by the Minister responsible for Statistics Canada, 2023

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	networkingv1alpha1 "statcan.gc.ca/cidr-allocator/api/v1alpha1"
	"statcan.gc.ca/cidr-allocator/controllers"
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
	// developmentLogging specifies whether to enable Development (Debug) logging for ZAP. Otherwise, Zap Production logging will be used
	developmentLogging bool
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
		lookupEnvOrDefault("LEADER_ELECTION_ID",
			fmt.Sprintf("%s-cidr-allocator-leader.statcan.gc.ca", uuid.NewUUID())),
		"The identity to use for leader-election",
	)
	flag.BoolVar(
                &enableLeaderElection,
		"leader-elect",
		false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.",
	)
	flag.BoolVar(
		&developmentLogging,
		"dev-mode",
		false,
		"Enable development mode (Logging)",
	)

	opts := zap.Options{
		Development: developmentLogging,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Configure ZAP Logger
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(networkingv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       leaderElectionID,
	})
	if err != nil {
		setupLog.Error(
			err,
			"unable to start NodeCIDRAllocation controller",
		)
		os.Exit(1)
	}

	if err = (&controllers.NodeCIDRAllocationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("NodeCIDRAllocation-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(
			err,
			"unable to create controller", "controller", "NodeCIDRAllocation",
		)
		os.Exit(1)
	}
	if err = (&networkingv1alpha1.NodeCIDRAllocation{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(
			err,
			"unable to create webhook", "webhook", "NodeCIDRAllocation",
		)
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(
			err,
			"unable to set up health check",
		)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(
			err,
			"unable to set up ready check",
		)
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(
			err,
			"problem running manager",
		)
		os.Exit(1)
	}
}
