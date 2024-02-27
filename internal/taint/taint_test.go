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

package taint_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"statcan.gc.ca/cidr-allocator/internal/taint"
)

const nodeTaintKey = "abc"

func TestNodeTaintClientHasTaint(t *testing.T) {
	node := &corev1.Node{
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key: nodeTaintKey,
				},
			},
		},
	}

	ntc := taint.New()

	want := true
	got := ntc.HasTaint(node, nodeTaintKey)

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	want = false
	got = ntc.HasTaint(node, "xyz")

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestNodeTaintClientAddNodeTaint(t *testing.T) {
	node := &corev1.Node{
		Spec: corev1.NodeSpec{},
	}

	ntc := taint.New()
	ntc.AddNodeTaint(node, corev1.Taint{Key: nodeTaintKey, Value: "true"})

	want := true
	got := len(node.Spec.Taints) == 1 && node.Spec.Taints[0].Key == nodeTaintKey

	if got != want {
		t.Errorf("got %v, wanted %v", node.Spec.Taints, []corev1.Taint{{Key: nodeTaintKey, Value: "true"}})
	}
}

func TestNodeTaintClientRemoveNodeTaint(t *testing.T) {
	node := &corev1.Node{
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{{Key: nodeTaintKey}},
		},
	}

	ntc := taint.New()
	ntc.RemoveNodeTaint(node, nodeTaintKey)

	want := true
	got := len(node.Spec.Taints) == 0

	if got != want {
		t.Errorf("got %v, wanted %v", node.Spec.Taints, []corev1.Taint{})
	}
}
