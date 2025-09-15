package provisioner

import "klonekit/pkg/blueprint"

// Provisioner defines the interface for infrastructure provisioning operations.
// This interface is provider-agnostic and can be implemented by any provisioning tool
// such as Terraform with different cloud providers (AWS, Azure, GCP, Vercel, etc.).
type Provisioner interface {
	// Provision executes the infrastructure provisioning based on the blueprint specification.
	// The autoApprove parameter controls whether to automatically apply changes or just validate.
	Provision(spec *blueprint.Spec, autoApprove bool) error
}