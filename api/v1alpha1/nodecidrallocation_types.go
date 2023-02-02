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
type NodeCIDRAllocationStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// This is a CRD that defines a NodeCIDRAllocation resource which allows for the allocation of node pod ranges
// to be assigned to nodes in a cluster. This is implemented using a list of network CIDRs as blocks of available address space that can be allocated
// to nodes using a node selector to filter the nodes upon which to apply the Pod CIDRs.
// +kubebuilder:printcolumn:name="Created",type="date",JSONPath=".metadata.creationTimestamp",description="NodeCIDRAllocation creation timestamp"
// +kubebuilder:printcolumn:name="Pools",type="string",JSONPath=".spec.addressPools",description="NodeCIDRAllocation Address Pools"
// +kubebuilder:printcolumn:name="NodeSelector",type="string",JSONPath=".spec.nodeSelector",description="NodeCIDRAllocation NodeSelector value"
type NodeCIDRAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeCIDRAllocationSpec   `json:"spec,omitempty"`
	Status NodeCIDRAllocationStatus `json:"status,omitempty"`
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
