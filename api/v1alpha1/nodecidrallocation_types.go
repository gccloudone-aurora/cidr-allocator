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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HealthStatus string

const (
	HealthStatusHealthy     HealthStatus = "Healthy"
	HealthStatusProgressing HealthStatus = "Progressing"
	HealthStatusUnhealthy   HealthStatus = "Unhealthy"
)

// NodeCIDRAllocationSpec defines the desired state of NodeCIDRAllocation
// This CRD defines an allocation of Node Pod ranges to be assigned to nodes in the cluster
type NodeCIDRAllocationSpec struct {
	// AddressPools represents a list of basic address pools in the form of a list of
	// network CIDRs that can be allocated to nodes running in the cluster.
	// These pools exist as a base subnet for the allocation of dynamically sized and positioned podCIDRs which will be
	// applied to Nodes that match the provided node selector
	// +required
	// +patchStrategy=merge
	// +kubebuilder:validation:MinItems=1
	AddressPools []string `json:"addressPools,omitempty" protobuf:"bytes,7,opt,name=addressPools" patchStrategy:"merge"`

	// NodeSelector represents a Kubernetes node selector to filter nodes from
	// the cluster for which to apply Pod CIDRs onto.
	// NOTE: Nodes that are selected through the node selector MUST specify a maximum number of pods in order to help identify
	//       the correct size for the NodeCIDRAllocation Controller to allocate to it. If none is specified a subnet WILL NOT be allocated for the Node.
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
}

// NodeCIDRAllocationStatus defines the observed state of NodeCIDRAllocation
// Nodes matching the supplied .Spec.NodeSelector are tracked by watching *corev1.Node resources in the cluster
// Actual state in the cluster is calculated at runtime using information from the matching Node resources
// The Status for NodeCIDRAllocation will be used for reporting purposes ONLY and may not always be up-to-date with the actual state of the cluster
type NodeCIDRAllocationStatus struct {
	// Health represents the current health of the NodeCIDRAllocation resource
	// Health status can be one of:
	//    v1alpha1.HealthStatusHealthy       - Represents a NodeCIDRAllocation resource that has performed all allocations and none have failed or are in a failing state
	//    v1alpha1.HealthStatusProgressing   - Represents a NodeCIDRAllocation resource that is progressing or otherwise does not have a determined health state
	//    v1alpha1.HealthStatusUnhealthy     - Represents a NodeCIDRAllocation resource that is currently tracking failed node allocations or failure to calculate the correct state of the cluster
	// +optional
	Health HealthStatus `json:"health,omitempty"`

	// ExpectedAllocations tracks the total number of Nodes being tracked for CIDR allocations using this NodeCIDRAllocation resource
	// +optional
	ExpectedAllocations int32 `json:"expected,omitempty"`

	// CompletedAllocations tracks the total number of Nodes being tracked that have successfully completed a CIDR allocation using this NodeCIDRAllocation resource
	// +optional
	CompletedAllocations int32 `json:"completed,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// This is a CRD that defines a NodeCIDRAllocation resource which allows for the allocation of node pod ranges
// to be assigned to nodes in a cluster. This is implemented using a list of network CIDRs as blocks of available address space that can be allocated
// to nodes using a node selector to filter the nodes upon which to apply the Pod CIDRs.
// +kubebuilder:printcolumn:name="Created",type="date",JSONPath=".metadata.creationTimestamp",description="NodeCIDRAllocation creation timestamp"
// +kubebuilder:printcolumn:name="Pools",type="string",JSONPath=".spec.addressPools",description="NodeCIDRAllocation Address Pools"
// +kubebuilder:printcolumn:name="NodeSelector",type="string",JSONPath=".spec.nodeSelector",description="NodeCIDRAllocation NodeSelector value"
// +kubebuilder:printcolumn:name="Health",type="string",JSONPath=".status.health",description="Current NodeCIDRAllocation resource Health"
// +kubebuilder:printcolumn:name="Expected",type="integer",JSONPath=".status.expected",description="Expected number of Node allocations"
// +kubebuilder:printcolumn:name="Completed",type="integer",JSONPath=".status.completed",description="Completed Node allocations"
type NodeCIDRAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeCIDRAllocationSpec   `json:"spec,omitempty"`
	Status NodeCIDRAllocationStatus `json:"status,omitempty"`
}

func (n *NodeCIDRAllocation) HealthStatus() HealthStatus {
	return n.Status.Health
}

func (n *NodeCIDRAllocation) ExpectedAllocations() int32 {
	return n.Status.ExpectedAllocations
}

func (n *NodeCIDRAllocation) CompletedAllocations() int32 {
	return n.Status.CompletedAllocations
}

func (n *NodeCIDRAllocation) SetHealthStatus(newStatus HealthStatus) {
	if newStatus == HealthStatusHealthy || newStatus == HealthStatusProgressing || newStatus == HealthStatusUnhealthy {
		n.Status.Health = newStatus
	}
}

func (n *NodeCIDRAllocation) SetExpectedAllocations(expected int32) {
	n.Status.ExpectedAllocations = expected
}

func (n *NodeCIDRAllocation) SetCompletedAllocations(completed int32) {
	n.Status.CompletedAllocations = completed
}

//+kubebuilder:object:root=true

// NodeCIDRAllocationList contains a list of NodeCIDRAllocation
type NodeCIDRAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeCIDRAllocation `json:"items"`
}

// init Registers the NodeCIDRAllocation CRD with the provided manager Scheme
func init() {
	SchemeBuilder.Register(&NodeCIDRAllocation{}, &NodeCIDRAllocationList{})
}
