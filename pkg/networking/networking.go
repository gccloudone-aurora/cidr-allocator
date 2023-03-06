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

package networking

import (
	"fmt"
	"math"
	"net"

	"github.com/c-robinson/iplib"
	corev1 "k8s.io/api/core/v1"
)

// SmallestCIDRForHosts calculates the smallest number of network bits required to
// satisfy the required number of hosts supplied
func SmallestCIDRForHosts(requiredHosts int) int {
	return 32 - int(math.Ceil(math.Log2(float64(requiredHosts+2))))
}

// UsableHosts will calculate the total number of usable hosts provided a number of 1s (or network-assigned bits) from a network mask
// desired is the total number of network assigned bits for the desired inner network/subnet
// max (optional) is the total size of an outer network that is not the default maximum (default: 32)
func NetHosts(desired float64, max ...float64) float64 {
	m := float64(32)
	if len(max) > 0 && max[0] > 0 && max[0] <= 32 {
		m = float64(max[0])
	}

	return math.Pow(2, float64(m-desired))
}

// UsableHosts is an alias for removing 2 unusable hosts (subnet addr, broadcast) from the total network hosts from `NetHosts()`
func UsableHosts(desired float64, max ...float64) float64 {
	return NetHosts(desired, max...) - 2
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
		offset := i * int(NetHosts(float64(netBits)))

		nextSub := iplib.IncrementIP4By(poolNet.IP, uint32(offset))
		subnets[i] = fmt.Sprintf("%s/%d", nextSub, netBits)
	}

	return subnets, nil
}

// StringNetIntersect determines whether the supplied networks (in CIDR format)
// are overlapping or otherwise have an intersection between them
func StringNetIntersect(a, b string) (bool, error) {
	_, aNet, err := iplib.ParseCIDR(a)
	if err != nil {
		return false, err
	}
	_, bNet, err := iplib.ParseCIDR(b)
	if err != nil {
		return false, err
	}

	return a == b || aNet.Contains(bNet.IP()) || bNet.Contains(aNet.IP()), nil
}

// StringNetIsAllocated uses a variety of conditions to ensure that there is no
// conflicting allocation that would present problems for subnet.
// returns (true,nil) when the subnet provided is not allocated by or overlapping with any nodes.
func StringNetIsAllocated(subnet string, nodes *corev1.NodeList) (bool, error) {
	for _, n := range nodes.Items {
		if n.Spec.PodCIDR == "" {
			continue
		}

		networkOverlapExists, err := StringNetIntersect(subnet, n.Spec.PodCIDR)
		if err != nil {
			return false, err
		}

		if networkOverlapExists || subnet == n.Spec.PodCIDR {
			return true, nil
		}
	}

	return false, nil
}
