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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"statcan.gc.ca/cidr-allocator/pkg/util"
)

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
