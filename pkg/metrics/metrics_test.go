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

package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"statcan.gc.ca/cidr-allocator/api/v1alpha1"
	"statcan.gc.ca/cidr-allocator/pkg/metrics"
)

func TestGet(t *testing.T) {
	got := metrics.Get()

	if len(got) <= 0 {
		t.Errorf("did not get any metrics collectors (length: %d)", len(got))
	}
}

func TestUpdate(t *testing.T) {
	allocations := &v1alpha1.NodeCIDRAllocationList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []v1alpha1.NodeCIDRAllocation{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1alpha1.NodeCIDRAllocationSpec{
					AddressPools: []string{"10.0.0.0/24"},
					NodeSelector: map[string]string{},
				},
				Status: v1alpha1.NodeCIDRAllocationStatus{},
			},
		},
	}
	nodes := &corev1.NodeList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []corev1.Node{
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testNodeA",
				},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.0/26", // (64-2 * 4)
				},
				Status: corev1.NodeStatus{},
			},
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testNodeB",
				},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.64/26",
				},
				Status: corev1.NodeStatus{},
			},
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testNodeC",
				},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.128/26",
				},
				Status: corev1.NodeStatus{},
			},
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testNodeD",
				},
				Spec: corev1.NodeSpec{
					PodCIDR: "10.0.0.192/26",
				},
				Status: corev1.NodeStatus{},
			},
		},
	}

	// Case 1: Address pools of 10.0.0.0/24 for 4 nodes each occupying a /26 (64 IPs)
	// expected: should result in 0 available hosts and expected allocations equal to actual allocations.
	// This result is given by: 2^8 - 4(2^(32-26)) = 0
	metrics.Update(allocations, nodes)

	for _, c := range metrics.Get() {
		if metrics.GetMetricValue(c) == -1 {
			t.Error("got metrics that contained non-supported metrics types")
		}
	}

	expectedVal := metrics.GetMetricValue(metrics.ExpectedAllocations())
	actualVal := metrics.GetMetricValue(metrics.ActualAllocations())

	if expectedVal != actualVal {
		t.Errorf("got %.0f, wanted %.0f", actualVal, expectedVal)
	}

	actualVal = metrics.GetMetricValue(metrics.AvailableHosts())
	actualValPercent := metrics.GetMetricValue(metrics.AvailableHostsPercent())
	expectedVal = 0

	if expectedVal != actualVal || expectedVal != actualValPercent {
		t.Errorf("got (%.0f, %.0f%%), wanted (%.0f, %.0f%%)", actualVal, actualValPercent, expectedVal, expectedVal)
	}

	// Case 2: Address pools of 10.0.0.0/24 for 4 nodes where 1 does not have a PodCIDR set
	// expected: should result in expected allocations not equal to actual allocations
	// and the difference being 1. Additionally, the available hosts should be 64 and as a percent should be 25.
	nodes.Items[0].Spec.PodCIDR = ""
	metrics.Update(allocations, nodes)

	actualVal = metrics.GetMetricValue(metrics.ActualAllocations())
	expectedVal = metrics.GetMetricValue(metrics.ExpectedAllocations())

	if expectedVal == actualVal {
		t.Errorf("got %.0f, wanted %.0f", actualVal, expectedVal)
	}

	if actualVal != expectedVal-1 {
		t.Errorf("got %.0f, wanted %.0f", actualVal, expectedVal-1)
	}

	actualVal = metrics.GetMetricValue(metrics.AvailableHosts())
	actualValPercent = metrics.GetMetricValue(metrics.AvailableHostsPercent())
	expectedVal = 64
	expectedValPercent := float64(25)

	if expectedVal != actualVal {
		t.Errorf("got %.0f, wanted %.0f", actualVal, expectedVal)
	}

	if actualValPercent != expectedValPercent {
		t.Errorf("got %.0f, wanted %.0f", actualValPercent, expectedValPercent)
	}
}

func TestGetMetricValue(t *testing.T) {
	m1 := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "testGaugeA",
	})
	m2 := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "testCounterA",
	})
	m3 := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "testHistA",
	})

	// Case 1: Metric is a Gauge
	// expected: the metrics value for the gauge metric m1
	m1.Set(10)
	var got float64 = metrics.GetMetricValue(m1)
	var want float64 = 10

	if got != want {
		t.Errorf("got %.0f, wanted %.0f", got, want)
	}

	// Case 2: Metric is a Counter
	// expected: the metrics value for the counter metric m2
	m2.Inc()
	got = metrics.GetMetricValue(m2)
	want = 1

	if got != want {
		t.Errorf("got %.0f, wanted %.0f", got, want)
	}

	// Case 3: Metric is something other than a Counter or Gauge (not supported)
	// expected: -1 to indicate that the metric value was not of the expected type
	m3.Observe(20)
	got = metrics.GetMetricValue(m3)
	want = -1

	if got != want {
		t.Errorf("got %.0f, wanted %.0f", got, want)
	}
}
