package scm

import (
	"fmt"
	"log/slog"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitlab "github.com/xanzy/go-gitlab"

	"klonekit/pkg/blueprint"
)

// ScmProvider defines the interface for source control management operations.
type ScmProvider interface {
	CreateRepo(spec *blueprint.Spec) error
}

// GitLabProvider implements the ScmProvider interface for GitLab.
type GitLabProvider struct {
	client *gitlab.Client
	token  string
}

// NewGitLabProvider creates a new GitLabProvider with authentication.
func NewGitLabProvider() (*GitLabProvider, error) {
	token := os.Getenv("GITLAB_PRIVATE_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITLAB_PRIVATE_TOKEN environment variable is required")
	}

	// For now, use gitlab.com as the default URL
	// In production, this should be configurable from the blueprint
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &GitLabProvider{
		client: client,
		token:  token,
	}, nil
}

// CreateRepo creates a GitLab repository and pushes the scaffolded files to it.
func (g *GitLabProvider) CreateRepo(spec *blueprint.Spec) error {
	slog.Info("Creating GitLab repository", "name", spec.SCM.Project.Name, "namespace", spec.SCM.Project.Namespace)

	// Check if repository already exists
	repoPath := fmt.Sprintf("%s/%s", spec.SCM.Project.Namespace, spec.SCM.Project.Name)
	existingProject, _, err := g.client.Projects.GetProject(repoPath, nil)
	if err == nil && existingProject != nil {
		slog.Warn("Repository already exists, skipping creation", "path", repoPath)
		return nil
	}

	// Set default visibility to private if not specified
	visibility := spec.SCM.Project.Visibility
	if visibility == "" {
		visibility = "private"
	}

	// Convert visibility string to GitLab visibility level
	var visibilityLevel gitlab.VisibilityValue
	switch visibility {
	case "private":
		visibilityLevel = gitlab.PrivateVisibility
	case "public":
		visibilityLevel = gitlab.PublicVisibility
	case "internal":
		visibilityLevel = gitlab.InternalVisibility
	default:
		visibilityLevel = gitlab.PrivateVisibility
	}

	// Create the project
	createOpts := &gitlab.CreateProjectOptions{
		Name:                     &spec.SCM.Project.Name,
		Path:                     &spec.SCM.Project.Name,
		Description:              &spec.SCM.Project.Description,
		Visibility:               &visibilityLevel,
		InitializeWithReadme:     gitlab.Bool(false),
		IssuesEnabled:            gitlab.Bool(true),
		MergeRequestsEnabled:     gitlab.Bool(true),
		WikiEnabled:              gitlab.Bool(true),
		SnippetsEnabled:          gitlab.Bool(true),
		AutoDevopsEnabled:        gitlab.Bool(false),
		SharedRunnersEnabled:     gitlab.Bool(true),
		ContainerRegistryEnabled: gitlab.Bool(true),
		PackagesEnabled:          gitlab.Bool(true),
	}

	project, _, err := g.client.Projects.CreateProject(createOpts)
	if err != nil {
		return fmt.Errorf("failed to create GitLab project: %w", err)
	}

	slog.Info("GitLab repository created successfully", "id", project.ID, "url", project.HTTPURLToRepo)

	// Initialize git repository and push files
	if err := g.initializeAndPushRepo(spec, project.HTTPURLToRepo); err != nil {
		return fmt.Errorf("failed to initialize and push repository: %w", err)
	}

	return nil
}

// initializeAndPushRepo initializes a git repository in the scaffolded directory and pushes to GitLab.
func (g *GitLabProvider) initializeAndPushRepo(spec *blueprint.Spec, repoURL string) error {
	scaffoldDir := spec.Scaffold.Destination

	// Check if the scaffold directory exists
	if _, err := os.Stat(scaffoldDir); os.IsNotExist(err) {
		return fmt.Errorf("scaffold directory does not exist: %s", scaffoldDir)
	}

	slog.Info("Initializing git repository", "directory", scaffoldDir)

	// Initialize git repository
	repo, err := git.PlainInit(scaffoldDir, false)
	if err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all files
	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// Create initial commit
	commit, err := worktree.Commit("Initial commit - scaffolded from KloneKit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "KloneKit",
			Email: "noreply@klonekit.dev",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	slog.Info("Created initial commit", "hash", commit)

	// Add remote origin
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote origin: %w", err)
	}

	// Push to remote
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: "oauth2", // GitLab uses oauth2 as username for token auth
			Password: g.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push to remote repository: %w", err)
	}

	slog.Info("Successfully pushed repository to GitLab", "url", repoURL)
	return nil
}
