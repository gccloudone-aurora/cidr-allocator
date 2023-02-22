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

package util

import "github.com/c-robinson/iplib"

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

// Keys takes a map and returns a list of keys from that map
func Keys[T comparable](m map[T]T) []T {
	keys := make([]T, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// RemoveByVal rebuilds the supplied list of items with the specified item removed
func RemoveByVal[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}
