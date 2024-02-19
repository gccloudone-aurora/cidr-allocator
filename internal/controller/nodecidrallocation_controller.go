/*
MIT License

Copyright (c) His Majesty the King in Right of Canada, as represented by the Minister responsible for Statistics Canada, 2024

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"statcan.gc.ca/cidr-allocator/api/v1alpha1"
	"statcan.gc.ca/cidr-allocator/internal/helper"
	statcan_metrics "statcan.gc.ca/cidr-allocator/internal/metrics"
	statcan_net "statcan.gc.ca/cidr-allocator/internal/networking"
)

const (
	finalizerName = "nodecidrallocation.networking.statcan.gc.ca/finalizer"
)

// NodeCIDRAllocationReconciler reconciles a NodeCIDRAllocation object
type NodeCIDRAllocationReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=networking.statcan.gc.ca,resources=nodecidrallocations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.statcan.gc.ca,resources=nodecidrallocations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.statcan.gc.ca,resources=nodecidrallocations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;patch;update;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *NodeCIDRAllocationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rl := log.FromContext(ctx)

	nodeCIDRAllocation := v1alpha1.NodeCIDRAllocation{}
	if err := r.Client.Get(ctx, req.NamespacedName, &nodeCIDRAllocation); err != nil {
		if apierrors.IsNotFound(err) {
			rl.V(1).Info(
				"Request object not found for NodeCIDRAllocation, it could have been deleted after reconcile request.",
				"name", req.Name,
				"namespace", req.Namespace,
			)

			// return and don't requeue
			return ctrl.Result{}, nil
		}

		rl.Error(
			err,
			"could not read NodeCIDRAllocation resource from API server",
			"name", req.Name,
			"namespace", req.Namespace,
		)

		// something else went wrong - return and requeue
		return ctrl.Result{}, err
	}

	matchingNodes := corev1.NodeList{}
	listOptions := client.ListOptions{
		LabelSelector: client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(nodeCIDRAllocation.Spec.NodeSelector)},
	}
	if err := r.Client.List(ctx, &matchingNodes, &listOptions, client.InNamespace(corev1.NamespaceAll)); err != nil {
		rl.Error(
			err,
			"unable to list Node resources from API server",
		)

		// could not list node resources from apiserver - return and requeue
		return ctrl.Result{}, err
	}

	// implement NodeCIDRAllocation resource finalizer to handle cleanup
	if nodeCIDRAllocation.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&nodeCIDRAllocation, finalizerName) {
			controllerutil.AddFinalizer(&nodeCIDRAllocation, finalizerName)
			if err := r.Update(ctx, &nodeCIDRAllocation); err != nil {
				if apierrors.IsNotFound(err) {
					// The resource no longer exists - return and do not requeue
					return ctrl.Result{}, nil
				}

				rl.Error(
					err,
					"unable to add Finalizer to NodeCIDRAllocation resource",
					"name", nodeCIDRAllocation.GetName(),
				)

				// something went wrong and could not add finalizer - return and requeue
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&nodeCIDRAllocation, finalizerName) {
			// check if there are any Nodes that are watched by this NodeCIDRAllocation that would be left orphaned
			if r.anyPodCIDRAllocated(&matchingNodes) {
				rl.V(1).Info(
					"there are existing Node allocations that are still tied to this resource. waiting until all nodes watched by this NodeCIDRAllocation resource are removed or no longer managed by this resource",
					"NodeCIDRAllocation", nodeCIDRAllocation.GetName(),
					"Selector", nodeCIDRAllocation.Spec.NodeSelector,
				)

				r.Recorder.Eventf(
					&nodeCIDRAllocation,
					corev1.EventTypeWarning,
					EventReasonOrphanedNodes,
					"Deletion of NodeCIDRAllocation resource (%s) would leave Nodes orphaned", nodeCIDRAllocation.GetName(),
				)

				// NodeCIDRAllocation is being deleted, but is not ready - return and do not requeue
				return ctrl.Result{}, nil
			}

			controllerutil.RemoveFinalizer(&nodeCIDRAllocation, finalizerName)
			if err := r.Update(ctx, &nodeCIDRAllocation); err != nil {
				if apierrors.IsNotFound(err) {
					// A previous reconcilliation likely removed the finalizer and completed the deletion of the resource - return and do not requeue
					return ctrl.Result{}, nil
				}

				rl.Error(
					err,
					"unable to remove finalizer from NodeCIDRAllocation resource",
					"NodeCIDRAllocation", nodeCIDRAllocation.GetName(),
				)

				// failed to remove finalizer - return and requeue
				return ctrl.Result{}, err
			}

			rl.Info(
				"nodeCIDRAllocation was removed",
				"name", nodeCIDRAllocation.GetName(),
			)

			r.Recorder.Eventf(
				&nodeCIDRAllocation,
				corev1.EventTypeNormal,
				EventReasonDeleted,
				"NodeCIDRAllocation resource was deleted: %s", nodeCIDRAllocation.GetName(),
			)

			// NodeCIDRAllocation has been successfully removed - return and do not requeue
			return ctrl.Result{}, nil
		}
	}

	if len(matchingNodes.Items) == 0 {
		rl.V(1).Info("no matching nodes exist. skipping")
		return ctrl.Result{}, nil
	}

	// retrieve a list of all Nodes in the cluster.
	// this is necessary since we need to ensure that we do not collide with any Node in the cluster regardless of whether it is managed by CIDR-Allocator or not.
	allClusterNodes := corev1.NodeList{}
	if err := r.Client.List(ctx, &allClusterNodes); err != nil {
		rl.Error(
			err,
			"unable to list Node resources from Kubernetes API server.",
		)

		// could not list Nodes in the cluster - return and requeue
		return ctrl.Result{}, err
	}

	rl.Info(
		"reconciling matching Nodes with NodeCIDRAllocation ...",
		"nodeCIDRAllocation", nodeCIDRAllocation.GetName(),
	)

	freeSubnets := make(map[uint8][]string)
	for _, node := range matchingNodes.Items {
		if node.Spec.PodCIDR != "" {
			rl.V(1).Info("node already contains CIDR allocation. skipping",
				"name", node.GetName(),
				"podCIDR", node.Spec.PodCIDR,
			)

			// Node does not need to be allocated a PodCIDR - move on to processing the next Node
			continue
		}

		maxPods := node.Status.Allocatable.Pods().Value()
		requiredCIDRMask := statcan_net.SmallestMaskForNumHosts(uint32(maxPods))

		rl.V(1).Info("determined Node resource PodCIDR requirements",
			"name", node.GetName(),
			"maxPods", maxPods,
			"requiredMaskCIDR", requiredCIDRMask,
		)

		for _, pool := range nodeCIDRAllocation.Spec.AddressPools {
			subnets, err := statcan_net.SubnetsFromPool(pool, requiredCIDRMask)
			if err != nil {
				rl.Error(
					err,
					"unable to break down address pool into subnets",
					"pool", pool,
					"maskCIDR", requiredCIDRMask,
				)

				return ctrl.Result{}, r.finalizeReconcile(ctx, &nodeCIDRAllocation, &matchingNodes, err)
			}

			for _, subnet := range subnets {
				networkAllocated, err := statcan_net.NetworkAllocated(subnet, &allClusterNodes)
				if err != nil {
					rl.Error(
						err,
						"unable to determine whether subnet is already allocated",
						"subnet", subnet,
					)
					return ctrl.Result{}, r.finalizeReconcile(ctx, &nodeCIDRAllocation, &matchingNodes, err)
				}

				if !networkAllocated && !helper.StringInSlice(subnet, freeSubnets[requiredCIDRMask]) {
					freeSubnets[requiredCIDRMask] = append(freeSubnets[requiredCIDRMask], subnet)
				}
			}
		}

		if len(freeSubnets[requiredCIDRMask]) == 0 {
			rl.Info("unable to allocate podCIDR for node. no available address space. you may want to add some additional pools",
				"name", node.GetName(),
				"requiredSubnetCIDR", requiredCIDRMask,
			)

			r.Recorder.Eventf(
				&nodeCIDRAllocation,
				corev1.EventTypeWarning,
				EventReasonNoAddressSpace,
				"There are no available subnets for the requested size (/%s). Could not assign PodCIDR to Node (%s)", requiredCIDRMask, node.GetName(),
			)

			// no available subnet to assign to Node - return and do not requeue
			return ctrl.Result{}, r.finalizeReconcile(ctx, &nodeCIDRAllocation, &matchingNodes, nil)
		}

		// Assign the first free subnet of the requested size PodCIDR to the Node
		node.Spec.PodCIDR, freeSubnets[requiredCIDRMask] = freeSubnets[requiredCIDRMask][0], freeSubnets[requiredCIDRMask][1:]
		if err := r.Update(ctx, &node); err != nil {
			if apierrors.IsNotFound(err) {
				// Node no longer found. It may have been deleted after reconcilliation request - return and do not requeue
				return ctrl.Result{}, r.finalizeReconcile(ctx, &nodeCIDRAllocation, &matchingNodes, nil)
			}
			rl.Error(err, "unable to set pod CIDR for Node resource",
				"name", node.GetName(),
				"podCIDR", node.Spec.PodCIDR,
				"freeAvailable", freeSubnets,
			)

			return ctrl.Result{}, r.finalizeReconcile(ctx, &nodeCIDRAllocation, &matchingNodes, err)
		}

		rl.Info(
			"assigned PodCIDR to Node resource",
			"name", node.GetName(),
			"podCIDR", node.Spec.PodCIDR,
			"remainingFreeSubnets", len(freeSubnets[requiredCIDRMask]),
		)
	}

	r.Recorder.Eventf(
		&nodeCIDRAllocation,
		corev1.EventTypeNormal,
		EventReasonAllocated,
		"PodCIDR Allocation has been applied to Matching Nodes { NodeSelector: %v, MatchingNodesCount: %d }", nodeCIDRAllocation.Spec.NodeSelector, len(matchingNodes.Items),
	)

	// Allocation successful for all matching Nodes - update current status + metrics + return and do not requeue
	return ctrl.Result{}, r.finalizeReconcile(ctx, &nodeCIDRAllocation, &matchingNodes, nil)
}

// anyPodCIDRAllocated checks the PodCIDR field in the Node spec for the provided nodes and returns true if **any** that field is allocated, otherwise, false.
func (r *NodeCIDRAllocationReconciler) anyPodCIDRAllocated(nodes *corev1.NodeList) bool {
	for _, node := range nodes.Items {
		if node.Spec.PodCIDR != "" {
			return true
		}
	}

	return false
}

// finalizeReconcile performs any final tasks/functions before the reconcile will be considered complete.
// this function will pass-through any errors so that information is not lost, but we can use it to adjust status and metric information
func (r *NodeCIDRAllocationReconciler) finalizeReconcile(ctx context.Context, nodeCIDRAllocation *v1alpha1.NodeCIDRAllocation, nodes *corev1.NodeList, err error) error {
	r.updateNodeCIDRAllocationStatus(ctx, nodeCIDRAllocation, nodes, err)
	r.updatePrometheusMetrics(ctx)

	// passthrough for err (if non-nil) to the Reconcile Result
	return err
}

// updatePrometheusMetrics will capture metrics for cluster-wide usage of the NodeCIDRAllocator.
// metrics are aggregate and considers all nodes and all NodeCIDRAllocation resources in its processes
func (r *NodeCIDRAllocationReconciler) updatePrometheusMetrics(ctx context.Context) {
	log := log.FromContext(ctx)
	allNodeCIDRAllocations := v1alpha1.NodeCIDRAllocationList{}
	if fErr := r.Client.List(ctx, &allNodeCIDRAllocations); fErr != nil {
		log.Error(
			fErr,
			"unable to get NodeCIDRAllocationList resource. cannot update metrics",
		)
		return
	}

	allNodes := corev1.NodeList{}
	if fErr := r.Client.List(ctx, &allNodes); fErr != nil {
		log.Error(
			fErr,
			"unable to get NodeList resource. cannot update metrics",
		)
		return
	}

	// calculate and update metrics
	statcan_metrics.Update(&allNodeCIDRAllocations, &allNodes)
}

// updateNodeCIDRAllocationStatus will calculate the current state of Cluster Node allocations for all matching Nodes from the provided NodeCIDRAllocation
// This function will additionally update Health of the NodeCIDRAllocation resource according to it's perceived state. The perceived state is then stored in
// the associated NodeCIDRAllocation's Status.
func (r *NodeCIDRAllocationReconciler) updateNodeCIDRAllocationStatus(ctx context.Context, nodeCIDRAllocation *v1alpha1.NodeCIDRAllocation, nodes *corev1.NodeList, err error) {
	log := log.FromContext(ctx)

	nodeCIDRAllocation.SetExpectedAllocations(int32(len(nodes.Items)))
	nodeCIDRAllocation.SetCompletedAllocations(0)
	nodeCIDRAllocation.SetHealthStatus(v1alpha1.HealthStatusHealthy)
	for _, node := range nodes.Items {
		if node.Spec.PodCIDR != "" {
			nodeCIDRAllocation.SetCompletedAllocations(nodeCIDRAllocation.CompletedAllocations() + 1)
		}
	}

	if nodeCIDRAllocation.ExpectedAllocations() != nodeCIDRAllocation.CompletedAllocations() {
		nodeCIDRAllocation.SetHealthStatus(v1alpha1.HealthStatusProgressing)
	}

	if err != nil {
		nodeCIDRAllocation.SetHealthStatus(v1alpha1.HealthStatusUnhealthy)
	}

	if err := r.Status().Update(ctx, nodeCIDRAllocation); err != nil {
		log.Error(
			err,
			"unable to update resource status for NodeCIDRAllocation",
		)
	}
}

// triggerNodeCIDRAllocationReconcileFromNodeChange is a mapping function which takes a Node object
// and returns a list of reconciliation requests for all NodeCIDRAllocation resources that have a matching NodeSelector
func (r *NodeCIDRAllocationReconciler) triggerNodeCIDRAllocationReconcileFromNodeChange(ctx context.Context, o client.Object) []reconcile.Request {
	allNodeCIDRAllocations := &v1alpha1.NodeCIDRAllocationList{}
	usedByNodeCIDRAllocation := map[*v1alpha1.NodeCIDRAllocation]struct{}{} // implements a set-like structure to ensure that we only process a single reconcile for each unique match

	// get all the available NodeCIDRAllocations on the cluster
	if err := r.Client.List(context.TODO(), allNodeCIDRAllocations, &client.ListOptions{
		Namespace: corev1.NamespaceAll,
	}); err != nil {
		return []reconcile.Request{}
	}

	// find CIDR allocations that have a NodeSelector that points to the node that triggered this reconciliation
	for _, item := range allNodeCIDRAllocations.Items {
		if helper.ObjectContainsLabels(o, item.Spec.NodeSelector) {
			usedByNodeCIDRAllocation[&item] = struct{}{}
		}
	}

	// create reconcile requests for all matching NodeCIDRAllocation resources for the Node object
	requests := make([]reconcile.Request, len(usedByNodeCIDRAllocation))
	i := 0
	for used := range usedByNodeCIDRAllocation {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      used.GetName(),
				Namespace: used.GetNamespace(),
			},
		}
		i++
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeCIDRAllocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&v1alpha1.NodeCIDRAllocation{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(r.triggerNodeCIDRAllocationReconcileFromNodeChange),
			builder.WithPredicates(predicate.Or(
				predicate.LabelChangedPredicate{},
			)),
		).
		Complete(r)
}

// init registers custom metrics
func init() {
	metrics.Registry.MustRegister(statcan_metrics.Get()...)
}
