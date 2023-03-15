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

package networking_test

import (
	"strings"
	"testing"

	"github.com/c-robinson/iplib"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"statcan.gc.ca/cidr-allocator/pkg/networking"
)

func TestSmallestMaskForNumHosts(t *testing.T) {
	// Case 1: Standard 254 required hosts
	// expected: should return 24 since 254+2(reserved) = 256 = 2^8 => 32-8 = 24
	got, err := networking.SmallestMaskForNumHosts(254)
	var want uint8 = 24

	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}
}

func TestNumHostsForMask(t *testing.T) {
	// Case 1: 26 ones in host network (between 1 and 32 inclusive) (valid input)
	// expected: 64
	got, _ := networking.NumHostsForMask(26)
	var want uint32 = 64

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// Case 2: ones are greater than 32
	// expected: should error
	_, err := networking.NumHostsForMask(33)
	if err == nil {
		t.Error("function was expected to return with an error")
	}

	// Case 3: ones are less than 1
	// expected: should error
	_, err = networking.NumHostsForMask(0)
	if err == nil {
		t.Error("function was expected to return with an error")
	}
}

func TestNumUsableHostsForMask(t *testing.T) {
	// Case 1: 26 ones in host network (valid)
	// expected: 62 (64 - 2 for reserved)
	got, _ := networking.NumUsableHostsForMask(26)
	var want uint32 = 62

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// Case 2: ones are greater than allowed 32
	// expected: should error
	_, err := networking.NumUsableHostsForMask(33)
	if err == nil {
		t.Error("function was expected to return with an error")
	}
}

func TestSubnetsFromPool(t *testing.T) {
	// Case 1: Incorrectly formatted pool CIDR
	// expected: should error
	_, err := networking.SubnetsFromPool("10.0.0/24", 26)
	if err == nil {
		t.Error("function was expected to return with an error")
	}

	// Case 2: Incorrect CIDR block size for pool
	// expected: should error
	_, err = networking.SubnetsFromPool("10.0.0.0/39", 26)
	if err == nil {
		t.Error("function was expected to return with an error")
	}

	// Case 3: Desired size is larger than pool
	// expected: should produce empty list of subnets
	// explanation: If the pool size cannot fit the desired sized subnet, this does not indicate that there is any invalid input.
	// The function is expected to the produce an empty list indicating that there are no suitable subnets that can be created within the provided address pool given the desired subnet mask.
	// In this case, we want to return the empty list so that the controller can continue to pass in address pools in order to find subnets that can be made and therefore used elsewhere in the program.
	got, err := networking.SubnetsFromPool("10.0.0.0/24", 23)
	want := []string{}
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if len(got) != len(want) {
		t.Errorf("got [%s], wanted []string{}", strings.Join(got, ","))
	}

	// Case 4: A /24 pool provided to be broken up into 4 /26
	// expected: 4 unique pools of size /26
	got, err = networking.SubnetsFromPool("10.0.0.0/24", 26)
	want = []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26"}
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		t.Errorf("got [%s], wanted [%s]", strings.Join(got, ","), strings.Join(want, ","))
	}

	// Case 5: A /27 pool provided with desired size of /28
	// expected: 2 unique pools of size /28
	got, err = networking.SubnetsFromPool("10.0.0.0/27", 28)
	want = []string{"10.0.0.0/28", "10.0.0.16/28"}
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got [%s], wanted [%s]", strings.Join(got, ","), strings.Join(want, ","))
	}
}

func TestNetworksOverlap(t *testing.T) {
	// Case 1: Network A is invalid
	// expected: should produce an error
	_, err := networking.NetworksOverlap("10.0.0/20", "10.0.0.0/24")
	if err == nil {
		t.Error("function was expected to return with an error")
	}

	// Case 2: Network B is invalid
	// expected: should produce an error
	_, err = networking.NetworksOverlap("10.0.0.0/24", "10.0.0.0/200")
	if err == nil {
		t.Error("function was expected to return with an error")
	}

	// Case 3: Network A is the same as Network B
	// expected: true
	got, err := networking.NetworksOverlap("10.0.0.0/24", "10.0.0.0/24")
	want := true
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 4: Network A is a superset of Network B
	// expected: true
	got, err = networking.NetworksOverlap("10.0.0.0/24", "10.0.0.128/25")
	want = true
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 5: Network B is a superset of Network A
	// expected: true
	got, err = networking.NetworksOverlap("10.0.0.128/25", "10.0.0.0/24")
	want = true
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 6: Network A is a non-overlapping unique network to Network B
	// expected: false
	got, err = networking.NetworksOverlap("10.0.1.0/24", "10.0.0.128/25")
	want = false
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestNetworkAllocated(t *testing.T) {
	subnets := []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26"}
	nodeList := corev1.NodeList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.Node{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.NodeSpec{
					PodCIDR: "",
				},
				Status: corev1.NodeStatus{},
			},
		},
	}

	// Case 1: All nodes specified do not have a PodCIDR in their spec
	// expected: false
	got, err := networking.NetworkAllocated(subnets[0], &nodeList)
	want := false
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 2: No nodes are specified
	// expected: false
	got, err = networking.NetworkAllocated(subnets[0], &corev1.NodeList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    []corev1.Node{},
	})
	want = false
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	_, _, e := iplib.ParseCIDR("10.0.0.259/24")
	t.Logf("error: %s", e.Error())

	// Case 3: Provided subnet is invalid
	// expected: error should be caught at NetworkAllocated and passed-through
	_, err = networking.NetworkAllocated("10.0.0.259/24", &corev1.NodeList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.Node{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.0/26",
				},
				Status: corev1.NodeStatus{},
			},
		},
	})
	if err == nil {
		t.Error("function was expected to return with an error")
	}

	// Case 4: Provided subnet does intersect with some existing allocation
	// expected: true
	got, err = networking.NetworkAllocated(subnets[1], &corev1.NodeList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.Node{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.64/28",
				},
				Status: corev1.NodeStatus{},
			},
		},
	})
	want = true
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	// Case 5: Provided subnet does not intersect with any Node, but at least one Node has an allocation to its PodCIDR
	// expected: false
	got, err = networking.NetworkAllocated(subnets[2], &corev1.NodeList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.Node{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.64/28",
				},
				Status: corev1.NodeStatus{},
			},
		},
	})
	want = false
	if err != nil {
		t.Errorf("function was not expected to error. got %e", err)
	} else if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}
