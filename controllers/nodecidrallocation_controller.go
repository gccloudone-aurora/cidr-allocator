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

package controllers

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"statcan.gc.ca/cidr-allocator/api/v1alpha1"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *NodeCIDRAllocationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("fetching NodeCIDRAllocation resource")
	nodeCIDRAllocation := v1alpha1.NodeCIDRAllocation{}
	if err := r.Client.Get(ctx, req.NamespacedName, &nodeCIDRAllocation); err != nil {
		log.Error(
			err,
			"unable to get NodeCIDRAllocation resource",
		)
		return ctrl.Result{}, nil
	}

	log.Info("fetching matching Node resources")
	matchingNodes := corev1.NodeList{}
	listOptions := client.ListOptions{
		LabelSelector: client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(nodeCIDRAllocation.Spec.NodeSelector)},
	}
	if err := r.Client.List(ctx, &matchingNodes, &listOptions, client.InNamespace(corev1.NamespaceAll)); err != nil {
		log.Error(
			err,
			"unable to get list of Node resources",
		)
		return ctrl.Result{}, nil
	}

	if len(matchingNodes.Items) == 0 {
		log.Info("no matching nodes exist. skipping")
		return ctrl.Result{}, nil
	}

	// implement NodeCIDRAllocation resource finalizer to handle cleanup
	if nodeCIDRAllocation.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&nodeCIDRAllocation, finalizerName) {
			controllerutil.AddFinalizer(&nodeCIDRAllocation, finalizerName)
			if err := r.Update(ctx, &nodeCIDRAllocation); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info(
			"NodeCIDRAllocation resource is queued for deletion",
			"resourceName", nodeCIDRAllocation.GetName(),
		)
		if controllerutil.ContainsFinalizer(&nodeCIDRAllocation, finalizerName) {
			readyForRemoval := true
			for _, node := range matchingNodes.Items {
				if node.Spec.PodCIDR != "" {
					log.Error(
						errors.New("a node allocation still exists"),
						"there is an existing Node allocation that is still tied to this resource. waiting until all nodes watched by this NodeCIDRAllocation resource are removed",
						"nodeName", node.GetName(),
						"NodeCIDRAllocation", nodeCIDRAllocation.GetName(),
					)
					readyForRemoval = false
				}
			}

			if readyForRemoval {
				controllerutil.RemoveFinalizer(&nodeCIDRAllocation, finalizerName)
				if err := r.Update(ctx, &nodeCIDRAllocation); err != nil {
					log.Error(
						err,
						"unable to remove finalizer from NodeCIDRAllocation resource",
						"NodeCIDRAllocation", nodeCIDRAllocation.GetName(),
					)
					return ctrl.Result{}, err
				}

				log.Info(
					"NodeCIDRAllocation resource has been deleted after resolving finalizer",
					"resourceName", nodeCIDRAllocation.GetName(),
				)
				return ctrl.Result{}, nil
			}
		}
	}

	freeSubnets := make(map[int][]string)
	for _, node := range matchingNodes.Items {
		maxPods := node.Status.Allocatable.Pods().Value()
		requiredMaskCIDR := SmallestCIDRForHosts(int(maxPods))

		log.V(2).Info("determined Node resource PodCIDR requirements",
			"nodeName", node.GetName(),
			"maxPods", maxPods,
			"requiredMaskCIDR", requiredMaskCIDR,
		)

		for _, pool := range nodeCIDRAllocation.Spec.AddressPools {
			subnets, err := SubnetsFromPool(pool, requiredMaskCIDR)
			if err != nil {
				log.Error(
					err,
					"unable to break down address pool into subnets",
					"pool", pool,
					"maskCIDR", requiredMaskCIDR,
				)
				return ctrl.Result{}, err
			}

			for _, subnet := range subnets {
				networkAllocated, err := StringNetIsAllocated(subnet, &matchingNodes)
				if err != nil {
					log.Error(
						err,
						"unable to determine whether subnet is already allocated",
						"subnet", subnet,
					)
					return ctrl.Result{}, err
				}

				if !networkAllocated && !StringInSlice(subnet, freeSubnets[requiredMaskCIDR]) {
					freeSubnets[requiredMaskCIDR] = append(freeSubnets[requiredMaskCIDR], subnet)
				}
			}
		}

		if node.Spec.PodCIDR != "" {
			log.V(2).Info("node already contains CIDR allocation. skipping",
				"nodeName", node.GetName(),
				"podCIDR", node.Spec.PodCIDR,
			)
			continue
		}

		if len(freeSubnets[requiredMaskCIDR]) == 0 {
			log.Info("unable to allocate podCIDR for node. no available address space. you may want to add some additional pools",
				"nodeName", node.GetName(),
				"requiredSubnetCIDR", requiredMaskCIDR,
			)

			r.Recorder.Eventf(
				&nodeCIDRAllocation,
				corev1.EventTypeWarning,
				"failure",
				"no available address space to assign PodCIDR to node (%s)", node.GetName(),
			)
			return ctrl.Result{}, nil
		}

		node.Spec.PodCIDR, freeSubnets[requiredMaskCIDR] = freeSubnets[requiredMaskCIDR][0], freeSubnets[requiredMaskCIDR][1:]
		if err := r.Update(ctx, &node); err != nil {
			log.Error(err, "unable to set pod CIDR for Node resource",
				"nodeName", node.GetName(),
				"podCIDR", node.Spec.PodCIDR,
				"freeAvailable", freeSubnets,
			)

			return ctrl.Result{}, nil
		}

		log.Info("assigned podCIDR to Node resource", "nodeName", node.GetName(), "podCIDR", node.Spec.PodCIDR, "remainingFreeSubnets", len(freeSubnets[requiredMaskCIDR]))
		r.Recorder.Eventf(
			&node,
			corev1.EventTypeNormal,
			"update",
			"PodCIDR Allocation has been applied to Node resource (%s)", node.GetName(),
		)
	}

	return ctrl.Result{}, nil
}

// triggerNodeCIDRAllocationReconcileFromNodeChange is a mapping function which takes a Node object
// and returns a list of reconciliation requests for all NodeCIDRAllocation resources that have a matching NodeSelector
func (r *NodeCIDRAllocationReconciler) triggerNodeCIDRAllocationReconcileFromNodeChange(o client.Object) []reconcile.Request {
	allNodeCIDRAllocations := &v1alpha1.NodeCIDRAllocationList{}
	usedByNodeCIDRAllocation := &v1alpha1.NodeCIDRAllocationList{}

	// get all the available NodeCIDRAllocations on the cluster
	if err := r.Client.List(context.TODO(), allNodeCIDRAllocations, &client.ListOptions{
		Namespace: corev1.NamespaceAll,
	}); err != nil {
		return []reconcile.Request{}
	}

	// find CIDR allocations that have a NodeSelector that points to the node that triggered this reconciliation
	for _, item := range allNodeCIDRAllocations.Items {
		if ObjectContainsLabel(o, item.Spec.NodeSelector) {
			usedByNodeCIDRAllocation.Items = append(usedByNodeCIDRAllocation.Items, item)
		}
	}

	requests := make([]reconcile.Request, len(usedByNodeCIDRAllocation.Items))
	for i, item := range usedByNodeCIDRAllocation.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
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
			&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(r.triggerNodeCIDRAllocationReconcileFromNodeChange),
			builder.WithPredicates(predicate.Or(
				predicate.LabelChangedPredicate{},
			)),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 0}).
		Complete(r)
}
