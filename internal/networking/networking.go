/*
MIT License

Copyright (c) His Majesty the King in Right of Canada, as represented by the Minister responsible for Statistics Canada, 2024

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

// SmallestMaskForNumHosts calculates the smallest number of network bits required to
// satisfy the required number of hosts supplied
func SmallestMaskForNumHosts(requiredHosts uint32) uint8 {
	return uint8(32 - int(math.Ceil(math.Log2(float64(requiredHosts+2)))))
}

// NumHostsForMask will calculate the total number of addresses provided a number of 1s (network mask) from a network address
// ones is the total number of network assigned bits in the network mask
func NumHostsForMask(ones uint8) (uint32, error) {
	m := uint8(32)
	if ones > 32 || ones < 1 {
		return 0, fmt.Errorf("invalid desired network bits (ones) specified. %d. must be 1 <= ones <= 32", ones)
	}

	return uint32(math.Ceil(math.Pow(2, float64(m-ones)))), nil
}

// NumUsableHostsForMask is an alias for removing 2 unusable/reserved host addresses (subnet addr, broadcast) from the total network hosts from `NetHosts()`
func NumUsableHostsForMask(ones uint8) (uint32, error) {
	total, err := NumHostsForMask(ones)
	if err != nil {
		return 0, err
	}

	return total - 2, nil
}

// SubnetsFromPool breaks down the supplied pool into all possible subnets of the supplied size (given by ones)
func SubnetsFromPool(pool string, ones uint8) ([]string, error) {
	_, poolNet, err := net.ParseCIDR(pool)
	if err != nil {
		return []string{}, err
	}
	poolNetOnes, _ := poolNet.Mask.Size()
	numSubnets := uint32(math.Pow(2, float64(ones-uint8(poolNetOnes))))

	subnets := make([]string, numSubnets)
	for i := uint32(0); i < numSubnets; i++ {
		netSize, err := NumHostsForMask(ones)
		if err != nil {
			return []string{}, err
		}
		offset := i * uint32(netSize)

		nextSub := iplib.IncrementIP4By(poolNet.IP, uint32(offset))
		subnets[i] = fmt.Sprintf("%s/%d", nextSub, ones)
	}

	return subnets, nil
}

// NetworksOverlap determines whether the supplied networks (in CIDR format)
// are overlapping or otherwise have an intersection between them
func NetworksOverlap(a, b string) (bool, error) {
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

// NetworkAllocated uses a variety of conditions to ensure that there is no
// conflicting allocation that would present problems for subnet.
// returns (true,nil) when the subnet provided is not allocated by or overlapping with any nodes.
func NetworkAllocated(subnet string, nodes *corev1.NodeList) (bool, error) {
	for _, n := range nodes.Items {
		if n.Spec.PodCIDR == "" {
			continue
		}

		networksOverlap, err := NetworksOverlap(subnet, n.Spec.PodCIDR)
		if err != nil {
			return false, err
		}

		if networksOverlap {
			return true, nil
		}
	}

	return false, nil
}
