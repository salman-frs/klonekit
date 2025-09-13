package blueprint

// Blueprint is the root object that holds the entire configuration for a KloneKit execution.
// It's populated by parsing the user's klonekit.yaml file.
type Blueprint struct {
	APIVersion string   `yaml:"apiVersion" validate:"required"`
	Kind       string   `yaml:"kind" validate:"required,eq=Blueprint"`
	Metadata   Metadata `yaml:"metadata" validate:"required"`
	Spec       Spec     `yaml:"spec" validate:"required"`
}

// Metadata contains project-level metadata.
type Metadata struct {
	Name        string            `yaml:"name" validate:"required"`
	Description string            `yaml:"description"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

// Spec contains the detailed specifications for the orchestration.
type Spec struct {
	SCM       SCMProvider            `yaml:"scm" validate:"required"`
	Cloud     CloudProvider          `yaml:"cloud" validate:"required"`
	Scaffold  Scaffold               `yaml:"scaffold" validate:"required"`
	Variables map[string]interface{} `yaml:"variables,omitempty"`
}

// SCMProvider configuration for the Source Control Management provider.
type SCMProvider struct {
	Provider string        `yaml:"provider" validate:"required,oneof=gitlab"`
	URL      string        `yaml:"url" validate:"required,url"`
	Token    string        `yaml:"token" validate:"required"`
	Project  ProjectConfig `yaml:"project" validate:"required"`
}

// ProjectConfig defines the SCM project configuration.
type ProjectConfig struct {
	Name        string `yaml:"name" validate:"required"`
	Namespace   string `yaml:"namespace" validate:"required"`
	Description string `yaml:"description"`
	Visibility  string `yaml:"visibility" validate:"oneof=private public internal"`
}

// CloudProvider configuration for the Cloud provider.
type CloudProvider struct {
	Provider string `yaml:"provider" validate:"required,oneof=aws"`
	Region   string `yaml:"region" validate:"required"`
}

// Scaffold configuration for the file scaffolding process.
type Scaffold struct {
	Source      string `yaml:"source" validate:"required"`
	Destination string `yaml:"destination" validate:"required"`
}
