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

package util_test

import (
	"strings"
	"testing"

	"statcan.gc.ca/cidr-allocator/pkg/util"
)

func TestKeys(t *testing.T) {
	// Case 1: An empty map is provided
	// expected: return an empty list
	m := make(map[string]string, 0)
	got := util.Keys(m)
	want := []string{}

	if len(got) != len(want) {
		t.Errorf("got %v, wanted %v", got, want)
	}

	// Case 2: A map of keys with empty values
	// expected: returns a list of the keys (should ignore values)
	m2 := map[string]*string{"keyA": nil, "keyB": nil, "keyC": nil}
	got = util.Keys(m2)
	want = []string{"keyA", "keyB", "keyC"}

	for _, g := range got {
		contain := false
		for _, w := range want {
			if g == w {
				contain = true
			}
		}
		if !contain {
			t.Errorf("got [%s], wanted [%s]", strings.Join(got, ","), strings.Join(want, ","))
		}
	}

	// Case 3: A map of keys and values
	// expected: returns a list of keys from the map
	m3 := map[string]int{"keyA": 0, "keyB": 1}
	got = util.Keys(m3)
	want = []string{"keyA", "keyB"}

	for _, g := range got {
		contain := false
		for _, w := range want {
			if g == w {
				contain = true
			}
		}
		if !contain {
			t.Errorf("got [%s], wanted [%s]", strings.Join(got, ","), strings.Join(want, ","))
		}
	}
}

func TestRemoveByVal(t *testing.T) {
	// Case 1: Item does exist in the provided list
	// expected: should return a list with the item removed
	l := []int{1, 2, 3}
	n := 2
	got := util.RemoveByVal(l, n)
	want := []int{1, 3}

	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %v, wanted %v", got, want)
	}

	// Case 2: Item does not exist in the provided list
	// expected: should return the original list
	n = 4
	got = util.RemoveByVal(l, n)
	want = l

	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("got %v, wanted %v", got, want)
	}
}
