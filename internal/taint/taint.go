/*
MIT License

Copyright (c) His Majesty the King in Right of Canada, as represented by the
Minister responsible for Statistics Canada, 2024

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package taint

import (
	corev1 "k8s.io/api/core/v1"
)

// NodeTainter is an interface that describes how to handle Node taints
type NodeTainter interface {
	AddNodeTaint(*corev1.Node, corev1.Taint)
	RemoveNodeTaint(*corev1.Node, string)
	HasTaint(*corev1.Node, string) bool
}

// NodeTaintClient provides a set of related functionality for managing Node taints for the CIDR-Allocator.
//
// This implementation is opinionated and designed for use with the CIDR-Allocator
type NodeTaintClient struct{}

// A blank assignment to ensure conformity to the NodeTainer interface
var _ NodeTainter = &NodeTaintClient{}

// New creates a new NodeTaintClient and provides a reference to it
func New() *NodeTaintClient {
	return &NodeTaintClient{}
}

// taintExists checks the existing taints on the provided Node for one that has the key provided. If one exists, it returns true
func (n *NodeTaintClient) HasTaint(node *corev1.Node, key string) bool {
	for _, t := range node.Spec.Taints {
		if t.Key == key {
			return true
		}
	}

	return false
}

func (n *NodeTaintClient) AddNodeTaint(node *corev1.Node, taint corev1.Taint) {
	node.Spec.Taints = append(node.Spec.Taints, taint)
}

func (n *NodeTaintClient) RemoveNodeTaint(node *corev1.Node, key string) {
	taints := []corev1.Taint{}
	for _, taint := range node.Spec.Taints {
		if taint.Key != key {
			taints = append(taints, taint)
		}
	}

	node.Spec.Taints = taints
}
