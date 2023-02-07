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

package v1alpha1

import (
	"fmt"
	"net"
	"os"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"statcan.gc.ca/cidr-allocator/util"
)

// log is for logging in this package.
var NodeCIDRAllocationlog = logf.Log.WithName("NodeCIDRAllocation-resource")

// Configures Manager to manage a new Validating webhook.
// if ENV ENABLE_WEBHOOKS is 'false', this setup will return without creating a webhook
func (r *NodeCIDRAllocation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if os.Getenv("ENABLE_WEBHOOKS") == "false" {
		return nil
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-networking-statcan-gc-ca-v1alpha1-NodeCIDRAllocation,mutating=false,failurePolicy=fail,sideEffects=None,groups=networking.statcan.gc.ca,resources=NodeCIDRAllocations,verbs=create;update,versions=v1alpha1,name=vNodeCIDRAllocation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NodeCIDRAllocation{}

// ValidateCreate implements webhook.Validator for creation validation
func (r *NodeCIDRAllocation) ValidateCreate() error {
	NodeCIDRAllocationlog.Info(
		"validate create",
		"name", r.Name,
	)

	return r.ValidateNodeCIDRAllocation()
}

// ValidateUpdate implements webhook.Validator for update validation
func (r *NodeCIDRAllocation) ValidateUpdate(old runtime.Object) error {
	NodeCIDRAllocationlog.Info(
		"validate update",
		"name", r.Name,
	)

	return r.ValidateNodeCIDRAllocation()
}

// ValidateDelete implements webhook.Validator for deletion validation (this is not enabled and this implementation will not do anything)
func (r *NodeCIDRAllocation) ValidateDelete() error {
	NodeCIDRAllocationlog.Info(
		"validate delete",
		"name", r.Name,
	)
	return nil
}

func (r *NodeCIDRAllocation) ValidateNodeCIDRAllocation() error {
	var errs field.ErrorList
	if err := r.ValidateName(); err != nil {
		errs = append(errs, err)
	}

	if err := r.ValidateSpec(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "networking.statcan.gc.ca", Kind: "NodeCIDRAllocation"},
		r.Name, errs)
}

func (r *NodeCIDRAllocation) ValidateName() *field.Error {
	if len(r.ObjectMeta.Name) <= 0 || len(r.ObjectMeta.Name) > 63 {
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "NodeCIDRAllocation name cannot be empty and cannot exceed 63 characters in length")
	}

	return nil
}

func (r *NodeCIDRAllocation) ValidateSpec() *field.Error {
	if err := r.ValidateNodeSelector(r.Spec.NodeSelector, field.NewPath("spec").Child("nodeSelector")); err != nil {
		return err
	}

	if err := r.ValidateAddressPools(r.Spec.AddressPools, field.NewPath("spec").Child("addressPools")); err != nil {
		return err
	}

	return nil
}

func (r *NodeCIDRAllocation) ValidateNodeSelector(nodeSelector map[string]string, fldPath *field.Path) *field.Error {
	if len(nodeSelector) != 1 {
		return field.Invalid(fldPath, nodeSelector, "A single NodeSelector MUST be specified for this resource")
	}

	if len(strings.TrimSpace(util.Keys(nodeSelector)[0])) == 0 {
		return field.Invalid(fldPath, nodeSelector, "NodeSelector field must be specified and non-empty")
	}

	if len(strings.TrimSpace(nodeSelector[util.Keys(nodeSelector)[0]])) == 0 {
		return field.Invalid(fldPath, nodeSelector, "NodeSelector value must be specified and non-empty")
	}

	return nil
}

func (r *NodeCIDRAllocation) ValidateAddressPools(addressPools []string, fldPath *field.Path) *field.Error {
	if len(addressPools) == 0 {
		return field.Invalid(fldPath, addressPools, "AddressPools must contain at least one entry")
	}

	for i, pool := range addressPools {
		if _, _, err := net.ParseCIDR(pool); err != nil {
			return field.Invalid(fldPath.Child(fmt.Sprintf("%d", i)), pool, "pool is not in valid CIDR format and cannot be parsed")
		}

		for _, other := range util.RemoveByVal(addressPools, pool) {
			networkOverlapExists, err := util.StringNetIntersect(pool, other)
			if err != nil {
				NodeCIDRAllocationlog.Error(
					err,
					"could not determine network overlap for pool validation",
				)
				return field.InternalError(fldPath.Child(fmt.Sprintf("%d", i)), err)
			}
			if networkOverlapExists {
				return field.Invalid(fldPath.Child(fmt.Sprintf("%d", i)), pool, "pool intersects with another pool in the specified addressPools")
			}
		}
	}

	return nil
}
