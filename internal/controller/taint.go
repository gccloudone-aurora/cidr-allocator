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

package controller

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	nodeTaintKey = "node.networking.statcan.gc.ca/network-unavailable"
)

type NodeTainter interface {
	Handle(*corev1.Node)
}

type NodeTaintClient struct{}

var _ NodeTainter = &NodeTaintClient{}

func (n *NodeTaintClient) Handle(node *corev1.Node) {
	if node.Spec.PodCIDR == "" {
		n.addNodeTaint(node)
	} else {
		n.removeNodeTaintIfExists(node)
	}
}

func (n *NodeTaintClient) addNodeTaint(node *corev1.Node) {
	node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
		Key:    nodeTaintKey,
		Value:  "true",
		Effect: corev1.TaintEffectNoSchedule,
	})
}

func (n *NodeTaintClient) removeNodeTaintIfExists(node *corev1.Node) {
	taints := []corev1.Taint{}
	for _, taint := range node.Spec.Taints {
		if taint.Key != nodeTaintKey {
			taints = append(taints, taint)
		}
	}

	node.Spec.Taints = taints
}
