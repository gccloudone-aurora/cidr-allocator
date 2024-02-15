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

import "testing"

func TestAccumulatedHosts(t *testing.T) {
	cidrs := []string{"10.0.0.0/26", "10.0.0.64/27"}

	// Case 1: Valid CIDRs are provided and specify /26 and /27 as CIDR block size
	// expected: 64 + 32 = 96
	var got uint32 = accumulatedHosts(cidrs)
	var want uint32 = 96

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// Case 2: There is an invalid CIDR among the specified network CIDRs provided
	// expected: should calculate all addresses and ignore invalid CIDRs
	cidrs = append(cidrs, "10.0.0.128/36")
	got = accumulatedHosts(cidrs)
	want = 96

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// Case 3: No CIDRs are passed
	// expected: should return 0 accumulated addresses
	got = accumulatedHosts([]string{})
	want = 0

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}
}

func TestCalculateRemainingHosts(t *testing.T) {
	// Case 1: 64 available hosts and 16 allocated hosts
	// expected: should return (48, 75) where 48 is the number of remaining addresses and 75 is the percentage representation of remaining addresses
	got, gotP := calculateRemainingHosts(64, 16)
	var want, wantP float64 = 48, 75

	if got != want || gotP != wantP {
		t.Errorf("got (%.0f, %.0f), wanted (%.0f, %.0f)", got, gotP, want, wantP)
	}
}
