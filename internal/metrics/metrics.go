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

package metrics

import (
	"net"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	corev1 "k8s.io/api/core/v1"

	"statcan.gc.ca/cidr-allocator/api/v1alpha1"
	"statcan.gc.ca/cidr-allocator/internal/helper"
	statcan_net "statcan.gc.ca/cidr-allocator/internal/networking"
)

var (
	metricsExpectedAllocations = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cnp_cidr_allocator_expected_nodecidr_allocations",
		Help: "describes the total number of expected Node PodCIDR allocations that should be applied to this cluster by all CRs. This is equivallent to the number of watched Node resources.",
	})
	metricsActualAllocations = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cnp_cidr_allocator_actual_nodecidr_allocations",
		Help: "the total number of completed/present Node PodCIDR allocations for Nodes being watched",
	})
	metricsAvailableHosts = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cnp_cidr_allocator_available_hosts",
		Help: "the total number of configured host addresses remaining based on cumulative count for all configured address pools across ALL NodeCIDRAllocation CRs",
	})
	metricsAvailableHostsPercent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cnp_cidr_allocator_available_hosts_percent",
		Help: "the ratio of host address space remaining compared to the total number of addresses available for all configured address pools across ALL NodeCIDRAllocation CRs",
	})
)

// Get returns a list of all associated metrics collectors
// Note: values returned from this function will need to be coerced into the appropriate types if attempting to make changes with resulting collectors
// this is out-of-scope for this function, however. This function is designed to be used ONLY as an interface for supplying dynamic set of metrics
// to the controllers metrics registry directly.
func Get() []prometheus.Collector {
	return []prometheus.Collector{
		metricsExpectedAllocations,
		metricsActualAllocations,
		metricsAvailableHosts,
		metricsAvailableHostsPercent,
	}
}

func ExpectedAllocations() prometheus.Gauge {
	return metricsExpectedAllocations
}

func ActualAllocations() prometheus.Gauge {
	return metricsActualAllocations
}

func AvailableHosts() prometheus.Gauge {
	return metricsAvailableHosts
}

func AvailableHostsPercent() prometheus.Gauge {
	return metricsAvailableHostsPercent
}

// Update performs an update to ALL available metrics captured for the operator. These are not to be accessed or supplied via the `Get()` function,
// but rather from the local package variables. Metrics will be exposed via `Get()` outside of the package
func Update(nodeCIDRAllocations *v1alpha1.NodeCIDRAllocationList, allNodes *corev1.NodeList) {
	var notAllocated uint64
	nodeAllocationsCumulative := map[string]struct{}{}
	addressPoolsCumulative := map[string]struct{}{} // like helper.ObjectContainsLabel(...), we are using a map-like DS to avoid duplicates
	for _, n := range nodeCIDRAllocations.Items {
		for _, p := range n.Spec.AddressPools {
			addressPoolsCumulative[p] = struct{}{}
		}
	}
	totalAvailableHosts := accumulatedHosts(helper.Keys(addressPoolsCumulative))

	for _, n := range allNodes.Items {
		if n.Spec.PodCIDR == "" {
			notAllocated += 1
			continue
		}
		nodeAllocationsCumulative[n.Spec.PodCIDR] = struct{}{}
	}
	totalAllocatedHosts := accumulatedHosts(helper.Keys(nodeAllocationsCumulative))

	metricsExpectedAllocations.Set(float64(len(allNodes.Items)))
	metricsActualAllocations.Set(float64(len(allNodes.Items) - int(notAllocated)))

	remainingCount, remainingPercent := calculateRemainingHosts(totalAvailableHosts, totalAllocatedHosts)
	metricsAvailableHosts.Set(remainingCount)
	metricsAvailableHostsPercent.Set(remainingPercent)
}

// accumulatedHosts will calculate the total number of hosts accumulated for all networkCIDRs that are passed
func accumulatedHosts(networkCIDRs []string) uint32 {
	var totalSupportedHosts uint32
	for _, n := range networkCIDRs {
		_, ipNet, err := net.ParseCIDR(n)
		if err != nil {
			continue
		}
		networkOnes, _ := ipNet.Mask.Size()
		networkSize, err := statcan_net.NumHostsForMask(uint8(networkOnes))
		if err != nil {
			continue
		}

		totalSupportedHosts += networkSize
	}

	return totalSupportedHosts
}

// calculateRemainingHosts is a helper function to calculate remaining hosts given total available hosts and the number of hosts that are already allocated.
// this function returns both the total count of available hosts and a ratio (as a percent) of hosts left that are allocable.
func calculateRemainingHosts(totalAvailableHosts, totalAllocatedHosts uint32) (float64, float64) {
	return float64(totalAvailableHosts - totalAllocatedHosts), (1.0 - (float64(totalAllocatedHosts) / float64(totalAvailableHosts))) * 100
}

// GetMetricValue is a helper function from the metrics package to
// pull metric values as floats from prometheus metrics collectors.
//
// Currently, it supports only Counter and Gauge metrics.
func GetMetricValue(col prometheus.Collector) float64 {
	c := make(chan prometheus.Metric, 1)
	col.Collect(c)
	m := dto.Metric{}
	err := (<-c).Write(&m)
	if err != nil {
		return -1
	}

	if _, ok := col.(prometheus.Gauge); ok {
		return *m.Gauge.Value
	}

	if _, ok := col.(prometheus.Counter); ok {
		return *m.Counter.Value
	}

	return -1
}
