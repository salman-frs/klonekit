package runtime

import (
	"testing"
)

func TestGetDockerSocketPaths(t *testing.T) {
	paths := getDockerSocketPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one socket path, got none")
	}

	// Check that all paths are non-empty
	for i, path := range paths {
		if path == "" {
			t.Errorf("Socket path at index %d is empty", i)
		}
	}
}

func TestNewDockerRuntime_RequiresDockerDaemon(t *testing.T) {
	// This test will fail if Docker daemon is not running, but that's expected
	// We're testing the error handling path
	_, err := NewDockerRuntime()

	// We expect either success (if Docker is running) or a specific error format
	if err != nil {
		// Verify error message contains expected context
		errorMsg := err.Error()
		if errorMsg == "" {
			t.Error("Error message should not be empty")
		}

		// Should contain either "failed to create Docker client" or "failed to connect to Docker daemon"
		hasCreateError := len(errorMsg) > 0 && (errorMsg[:20] == "failed to create Doc" || errorMsg[:20] == "failed to connect to")
		if !hasCreateError {
			t.Errorf("Unexpected error format: %s", errorMsg)
		}
	}
}