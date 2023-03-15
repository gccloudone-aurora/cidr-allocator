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

package controllers_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"statcan.gc.ca/cidr-allocator/controllers"
)

func TestObjectContainsLabel(t *testing.T) {
	obj := &corev1.Node{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "testNode",
			Labels: map[string]string{
				"kubernetes.io/role": "agent",
				"kubernetes.io/os":   "linux",
			},
		},
	}

	// Case 1: All labels exists on object and has the correct value (expected: true)
	nodeSelectorLabels := map[string]string{
		"kubernetes.io/role": "agent",
	}

	got := controllers.ObjectContainsLabels(obj, nodeSelectorLabels)
	want := true

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 2: Some label does not exist in the object (expected: false)
	nodeSelectorLabels = map[string]string{
		"kubernetes.io/role": "agent",
		"kubernetes.io/bad":  "abc",
	}
	got = controllers.ObjectContainsLabels(obj, nodeSelectorLabels)
	want = false

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 3: All labels exists, but at least one has non-matching value (expected: false)
	nodeSelectorLabels = map[string]string{
		"kubernetes.io/role": "agent",
		"kubernetes.io/os":   "windows",
	}
	got = controllers.ObjectContainsLabels(obj, nodeSelectorLabels)
	want = false

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 4: All labels exist exactly as described on the object (expected: true)
	nodeSelectorLabels = map[string]string{
		"kubernetes.io/role": "agent",
		"kubernetes.io/os":   "linux",
	}
	got = controllers.ObjectContainsLabels(obj, nodeSelectorLabels)
	want = true

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestStringInSlice(t *testing.T) {
	s := "birthday"
	sl := []string{"happy", "birthday", "to", "you!"}

	// Case 1: The string does exist in the string slice
	got := controllers.StringInSlice(s, sl)
	want := true

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 2: The string does NOT exist in the string slice
	s = "Shenanigans"
	got = controllers.StringInSlice(s, sl)
	want = false

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}
