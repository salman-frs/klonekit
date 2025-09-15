package scm

import "klonekit/pkg/blueprint"

// ScmProvider defines the interface for source control management operations.
// This interface is provider-agnostic and can be implemented by any SCM provider
// such as GitLab, GitHub, Bitbucket, etc.
type ScmProvider interface {
	// CreateRepo creates a repository based on the blueprint specification.
	// It handles repository creation, initialization, and pushing scaffolded files.
	CreateRepo(spec *blueprint.Spec) error
}