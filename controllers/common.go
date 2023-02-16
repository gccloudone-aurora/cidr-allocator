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
	"fmt"
	"math"
	"net"

	"github.com/c-robinson/iplib"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"statcan.gc.ca/cidr-allocator/util"
)

// SmallestCIDRForHosts calculates the smallest number of network bits required to
// satisfy the required number of hosts supplied
func SmallestCIDRForHosts(requiredHosts int) int {
	return 32 - int(math.Ceil(math.Log2(float64(requiredHosts+2))))
}

// SubnetsFromPool breaks down the supplied pool into all possible subnets of the supplied size (given by netBits)
func SubnetsFromPool(pool string, netBits int) ([]string, error) {
	_, poolNet, err := net.ParseCIDR(pool)
	if err != nil {
		return []string{}, err
	}
	poolNetBits, _ := poolNet.Mask.Size()
	numSubnets := int(math.Pow(2, float64(netBits-poolNetBits)))

	subnets := make([]string, numSubnets)
	for i := 0; i < numSubnets; i++ {
		offset := i * (int(math.Pow(2, float64(32-netBits))))

		nextSub := iplib.IncrementIP4By(poolNet.IP, uint32(offset))
		subnets[i] = fmt.Sprintf("%s/%d", nextSub, netBits)
	}

	return subnets, nil
}

// ObjectContainsLabels is a utility function that parses a provided Kubernetes API client Objects'
// labels searching for a matching label
func ObjectContainsLabels(o client.Object, labels map[string]string) bool {
	for lk, lv := range labels {
		objectLabels := o.GetLabels()
		if StringInSlice(lk, util.Keys(objectLabels)) && objectLabels[lk] == lv {
			continue
		}
		return false
	}

	return true
}

// StringNetContainsNet compares string represented CIDR networks for composition.
// Returns true if `a` contains `b`
func StringNetContainsNet(a, b string) (bool, error) {
	_, aNet, err := iplib.ParseCIDR(a)
	if err != nil {
		return true, err
	}
	_, bNet, err := iplib.ParseCIDR(b)
	if err != nil {
		return true, err
	}

	return aNet.ContainsNet(bNet), nil
}

// StringInSlice looks for exact matches of the supplied string query in the supplied string slice
// Returns true if 's' is in [arr]
func StringInSlice(s string, arr []string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

// StringNetIsAllocated uses a variety of conditions to ensure that there is no
// conflicting allocation that would present problems for subnet.
// returns (true,nil) when the subnet provided is not allocated by or overlapping with any nodes.
func StringNetIsAllocated(subnet string, nodes *corev1.NodeList) (bool, error) {
	for _, n := range nodes.Items {
		if n.Spec.PodCIDR == "" {
			continue
		}

		subnetContainsNodeNet, err := StringNetContainsNet(subnet, n.Spec.PodCIDR)
		if err != nil {
			return false, err
		}
		networkOverlapExists, err := util.StringNetIntersect(subnet, n.Spec.PodCIDR)
		if err != nil {
			return false, err
		}

		if subnetContainsNodeNet || networkOverlapExists {
			return true, nil
		}
	}

	return false, nil
}
