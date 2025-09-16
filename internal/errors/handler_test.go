package errors

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewErrorHandler(t *testing.T) {
	handler, err := NewErrorHandler()
	if err != nil {
		t.Fatalf("NewErrorHandler() failed: %v", err)
	}

	if handler == nil {
		t.Fatal("NewErrorHandler() returned nil handler")
	}

	if handler.logger == nil {
		t.Error("ErrorHandler.logger is nil")
	}

	if handler.console == nil {
		t.Error("ErrorHandler.console is nil")
	}
}

func TestErrorHandler_Handle_KloneKitError(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	// Use custom log directory for testing
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "logs")
	os.Setenv("KLONEKIT_LOG_DIR", logDir)

	handler, err := NewErrorHandler()
	if err != nil {
		t.Fatalf("NewErrorHandler() failed: %v", err)
	}

	// Test handling of KloneKitError
	testErr := NewBlueprintError(
		"Test context",
		"Test cause",
		"Test suggestion",
		errors.New("original error"),
	)

	handler.Handle(testErr)

	// Verify log file was created and contains expected content
	logFile := filepath.Join(logDir, "klonekit.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestErrorHandler_Handle_GenericError(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	// Use custom log directory for testing
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "logs")
	os.Setenv("KLONEKIT_LOG_DIR", logDir)

	handler, err := NewErrorHandler()
	if err != nil {
		t.Fatalf("NewErrorHandler() failed: %v", err)
	}

	// Test handling of generic error
	testErr := errors.New("generic test error")

	handler.Handle(testErr)

	// Verify log file was created
	logFile := filepath.Join(logDir, "klonekit.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestErrorHandler_Handle_NilError(t *testing.T) {
	handler, err := NewErrorHandler()
	if err != nil {
		t.Fatalf("NewErrorHandler() failed: %v", err)
	}

	// Handle nil error should not panic
	handler.Handle(nil)
}

func TestGetErrorTypeName(t *testing.T) {
	tests := []struct {
		errorType error
		expected  string
	}{
		{ErrBlueprintNotFound, "blueprint_not_found"},
		{ErrBlueprintParseFailed, "blueprint_parse_failed"},
		{ErrScaffoldFailed, "scaffold_failed"},
		{ErrSCMFailed, "scm_failed"},
		{ErrProvisionFailed, "provision_failed"},
		{ErrRuntimeFailed, "runtime_failed"},
		{ErrConfigInvalid, "config_invalid"},
		{ErrNetworkFailed, "network_failed"},
		{ErrFileSystemFailed, "filesystem_failed"},
		{errors.New("unknown"), "unknown"},
	}

	for _, test := range tests {
		result := getErrorTypeName(test.errorType)
		if result != test.expected {
			t.Errorf("getErrorTypeName(%v) = %q, want %q", test.errorType, result, test.expected)
		}
	}
}

func TestGetDefaultHandler(t *testing.T) {
	// Reset singleton before test
	resetDefaultHandler()
	defer resetDefaultHandler()

	// Test that GetDefaultHandler returns the same instance on multiple calls
	handler1, err1 := GetDefaultHandler()
	if err1 != nil {
		t.Fatalf("GetDefaultHandler() first call failed: %v", err1)
	}

	handler2, err2 := GetDefaultHandler()
	if err2 != nil {
		t.Fatalf("GetDefaultHandler() second call failed: %v", err2)
	}

	if handler1 != handler2 {
		t.Error("GetDefaultHandler() should return the same instance on multiple calls")
	}
}

func TestHandleError(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
		// Reset singleton after test
		resetDefaultHandler()
	}()

	// Reset singleton before test
	resetDefaultHandler()

	// Use custom log directory for testing
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "logs")
	os.Setenv("KLONEKIT_LOG_DIR", logDir)

	testErr := errors.New("test error for HandleError")

	// Should not panic
	HandleError(testErr)

	// Verify log file was created in custom directory
	logFile := filepath.Join(logDir, "klonekit.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created by HandleError")
	}
}

func TestKloneKitError_Error(t *testing.T) {
	originalErr := errors.New("original error message")
	kloneKitErr := NewBlueprintError("context", "cause", "suggestion", originalErr)

	if kloneKitErr.Error() != originalErr.Error() {
		t.Errorf("KloneKitError.Error() = %q, want %q", kloneKitErr.Error(), originalErr.Error())
	}
}

func TestKloneKitError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error message")
	kloneKitErr := NewBlueprintError("context", "cause", "suggestion", originalErr)

	if kloneKitErr.Unwrap() != originalErr {
		t.Error("KloneKitError.Unwrap() should return the original error")
	}
}

func TestErrorConstructors(t *testing.T) {
	originalErr := errors.New("test error")

	tests := []struct {
		name        string
		constructor func(string, string, string, error) *KloneKitError
		expectedType error
	}{
		{"NewBlueprintError", NewBlueprintError, ErrBlueprintNotFound},
		{"NewParseError", NewParseError, ErrBlueprintParseFailed},
		{"NewScaffoldError", NewScaffoldError, ErrScaffoldFailed},
		{"NewSCMError", NewSCMError, ErrSCMFailed},
		{"NewProvisionError", NewProvisionError, ErrProvisionFailed},
		{"NewRuntimeError", NewRuntimeError, ErrRuntimeFailed},
		{"NewConfigError", NewConfigError, ErrConfigInvalid},
		{"NewNetworkError", NewNetworkError, ErrNetworkFailed},
		{"NewFileSystemError", NewFileSystemError, ErrFileSystemFailed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.constructor("context", "cause", "suggestion", originalErr)

			if err.Type != test.expectedType {
				t.Errorf("%s created error with type %v, want %v", test.name, err.Type, test.expectedType)
			}

			if err.Context != "context" {
				t.Errorf("%s created error with context %q, want %q", test.name, err.Context, "context")
			}

			if err.Cause != "cause" {
				t.Errorf("%s created error with cause %q, want %q", test.name, err.Cause, "cause")
			}

			if err.Suggestion != "suggestion" {
				t.Errorf("%s created error with suggestion %q, want %q", test.name, err.Suggestion, "suggestion")
			}

			if err.OriginalErr != originalErr {
				t.Errorf("%s created error with originalErr %v, want %v", test.name, err.OriginalErr, originalErr)
			}
		})
	}
}

// Test OS-standard log directory functionality

func TestGetOSStandardLogDir(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	t.Run("environment variable override", func(t *testing.T) {
		testDir := "/custom/log/dir"
		os.Setenv("KLONEKIT_LOG_DIR", testDir)

		result, err := getOSStandardLogDir()
		if err != nil {
			t.Fatalf("getOSStandardLogDir() failed: %v", err)
		}

		if result != testDir {
			t.Errorf("getOSStandardLogDir() = %q, want %q", result, testDir)
		}
	})

	t.Run("platform-specific directories", func(t *testing.T) {
		os.Unsetenv("KLONEKIT_LOG_DIR")

		result, err := getOSStandardLogDir()
		if err != nil {
			t.Fatalf("getOSStandardLogDir() failed: %v", err)
		}

		homeDir, _ := os.UserHomeDir()
		var expectedPath string

		switch runtime.GOOS {
		case "darwin":
			expectedPath = filepath.Join(homeDir, "Library", "Logs", "KloneKit")
		case "linux", "freebsd", "openbsd", "netbsd":
			expectedPath = filepath.Join(homeDir, ".local", "share", "klonekit", "logs")
		case "windows":
			appDataDir := os.Getenv("APPDATA")
			if appDataDir == "" {
				expectedPath = filepath.Join(homeDir, "AppData", "Roaming", "KloneKit", "logs")
			} else {
				expectedPath = filepath.Join(appDataDir, "KloneKit", "logs")
			}
		default:
			expectedPath = filepath.Join(homeDir, ".klonekit", "logs")
		}

		if result != expectedPath {
			t.Errorf("getOSStandardLogDir() = %q, want %q", result, expectedPath)
		}
	})
}

func TestCreateLogDirectoryWithFallback(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	t.Run("successful standard directory creation", func(t *testing.T) {
		tempDir := t.TempDir()
		logDir := filepath.Join(tempDir, "logs")
		os.Setenv("KLONEKIT_LOG_DIR", logDir)

		result, fallbackUsed, err := createLogDirectoryWithFallback()
		if err != nil {
			t.Fatalf("createLogDirectoryWithFallback() failed: %v", err)
		}

		if fallbackUsed {
			t.Error("createLogDirectoryWithFallback() should not use fallback for accessible directory")
		}

		if result != logDir {
			t.Errorf("createLogDirectoryWithFallback() = %q, want %q", result, logDir)
		}

		// Verify directory was created
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			t.Error("Log directory was not created")
		}
	})

	t.Run("fallback to current directory", func(t *testing.T) {
		// Set an invalid log directory (non-existent parent)
		invalidDir := "/non/existent/path/that/cannot/be/created"
		os.Setenv("KLONEKIT_LOG_DIR", invalidDir)

		result, fallbackUsed, err := createLogDirectoryWithFallback()
		if err != nil {
			t.Fatalf("createLogDirectoryWithFallback() failed: %v", err)
		}

		if !fallbackUsed {
			t.Error("createLogDirectoryWithFallback() should use fallback for inaccessible directory")
		}

		currentDir, _ := os.Getwd()
		if result != currentDir {
			t.Errorf("createLogDirectoryWithFallback() = %q, want %q", result, currentDir)
		}
	})
}

func TestCheckLogRotation(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	t.Run("no rotation needed for small file", func(t *testing.T) {
		// Create a small file
		content := strings.Repeat("small log entry\n", 10)
		err := os.WriteFile(logPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = checkLogRotation(logPath)
		if err != nil {
			t.Errorf("checkLogRotation() failed: %v", err)
		}

		// File should still exist
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Original log file should still exist")
		}
	})

	t.Run("rotation needed for large file", func(t *testing.T) {
		// Create a file larger than 10MB
		content := strings.Repeat("large log entry that takes up space\n", 300000) // ~10MB
		err := os.WriteFile(logPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}

		err = checkLogRotation(logPath)
		if err != nil {
			t.Errorf("checkLogRotation() failed: %v", err)
		}

		// Original file should be rotated to .1
		rotatedPath := logPath + ".1"
		if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
			t.Error("Rotated log file should exist")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "non-existent.log")
		err := checkLogRotation(nonExistentPath)
		if err != nil {
			t.Errorf("checkLogRotation() should not fail for non-existent file: %v", err)
		}
	})
}

func TestRotateLogFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Create test files
	testFiles := []string{
		logPath,
		logPath + ".1",
		logPath + ".2",
		logPath + ".3",
		logPath + ".4",
	}

	for i, file := range testFiles {
		content := fmt.Sprintf("Log file content %d\n", i)
		err := os.WriteFile(file, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	err := rotateLogFile(logPath)
	if err != nil {
		t.Fatalf("rotateLogFile() failed: %v", err)
	}

	// Check rotation results
	t.Run("current log moved to .1", func(t *testing.T) {
		content, err := os.ReadFile(logPath + ".1")
		if err != nil {
			t.Fatalf("Failed to read rotated file: %v", err)
		}
		if string(content) != "Log file content 0\n" {
			t.Errorf("Rotated file content = %q, want %q", string(content), "Log file content 0\n")
		}
	})

	t.Run("files rotated correctly", func(t *testing.T) {
		for i := 2; i <= 4; i++ {
			expectedContent := fmt.Sprintf("Log file content %d\n", i-1)
			content, err := os.ReadFile(fmt.Sprintf("%s.%d", logPath, i))
			if err != nil {
				t.Fatalf("Failed to read rotated file .%d: %v", i, err)
			}
			if string(content) != expectedContent {
				t.Errorf("Rotated file .%d content = %q, want %q", i, string(content), expectedContent)
			}
		}
	})

	t.Run("oldest file removed", func(t *testing.T) {
		// .5 should not exist (old .4 was removed)
		if _, err := os.Stat(logPath + ".5"); !os.IsNotExist(err) {
			t.Error("Oldest log file should be removed")
		}
	})

	t.Run("original log file removed", func(t *testing.T) {
		if _, err := os.Stat(logPath); !os.IsNotExist(err) {
			t.Error("Original log file should be moved")
		}
	})
}

func TestCreateLogFileWithOSStandardPaths(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "logs")
	os.Setenv("KLONEKIT_LOG_DIR", logDir)

	logFile, err := createLogFile()
	if err != nil {
		t.Fatalf("createLogFile() failed: %v", err)
	}
	defer logFile.Close()

	// Verify log file was created in correct location
	expectedPath := filepath.Join(logDir, "klonekit.log")
	if logFile.Name() != expectedPath {
		t.Errorf("createLogFile() created file at %q, want %q", logFile.Name(), expectedPath)
	}

	// Verify directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Log directory was not created")
	}
}

// Integration tests

func TestIntegrationErrorHandlingWithOSStandardLogging(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "integration-logs")
	os.Setenv("KLONEKIT_LOG_DIR", logDir)

	t.Run("end-to-end error handling with new logging", func(t *testing.T) {
		handler, err := NewErrorHandler()
		if err != nil {
			t.Fatalf("NewErrorHandler() failed: %v", err)
		}

		// Test handling various error types
		testErrors := []error{
			NewBlueprintError("test context", "test cause", "test suggestion", errors.New("blueprint test error")),
			NewSCMError("scm context", "scm cause", "scm suggestion", errors.New("scm test error")),
			errors.New("generic test error"),
		}

		for _, testErr := range testErrors {
			handler.Handle(testErr)
		}

		// Verify log file was created in correct location
		logPath := filepath.Join(logDir, "klonekit.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Fatal("Log file was not created in integration test")
		}

		// Verify log directory structure
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			t.Error("Log directory was not created in integration test")
		}
	})

	t.Run("log rotation integration", func(t *testing.T) {
		logPath := filepath.Join(logDir, "klonekit.log")

		// Create a large log file to trigger rotation
		largeContent := strings.Repeat("large log entry for rotation test\n", 350000) // >10MB
		err := os.WriteFile(logPath, []byte(largeContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create large log file: %v", err)
		}

		// Create new handler (should trigger rotation)
		handler, err := NewErrorHandler()
		if err != nil {
			t.Fatalf("NewErrorHandler() failed during rotation test: %v", err)
		}

		// Handle an error (should write to new log file)
		handler.Handle(errors.New("test error after rotation"))

		// Verify rotation occurred
		rotatedPath := logPath + ".1"
		if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
			t.Error("Log rotation did not occur in integration test")
		}

		// Verify new log file was created and contains new entry
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("New log file was not created after rotation")
		}
	})
}

func TestIntegrationFallbackLogging(t *testing.T) {
	// Save original environment and working directory
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	originalWd, _ := os.Getwd()
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
		os.Chdir(originalWd)
	}()

	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// Set an inaccessible log directory
	os.Setenv("KLONEKIT_LOG_DIR", "/root/inaccessible")

	t.Run("fallback logging integration", func(t *testing.T) {
		handler, err := NewErrorHandler()
		if err != nil {
			t.Fatalf("NewErrorHandler() failed during fallback test: %v", err)
		}

		// Handle an error
		testErr := NewConfigError("config context", "config cause", "config suggestion", errors.New("config test error"))
		handler.Handle(testErr)

		// Verify log file was created in current directory (fallback)
		fallbackLogPath := filepath.Join(tempDir, "klonekit.log")
		if _, err := os.Stat(fallbackLogPath); os.IsNotExist(err) {
			t.Error("Fallback log file was not created")
		}
	})
}

func TestIntegrationEnvironmentVariableOverride(t *testing.T) {
	// Save original environment
	originalLogDir := os.Getenv("KLONEKIT_LOG_DIR")
	defer func() {
		if originalLogDir != "" {
			os.Setenv("KLONEKIT_LOG_DIR", originalLogDir)
		} else {
			os.Unsetenv("KLONEKIT_LOG_DIR")
		}
	}()

	tempDir := t.TempDir()
	customLogDir := filepath.Join(tempDir, "custom", "log", "location")
	os.Setenv("KLONEKIT_LOG_DIR", customLogDir)

	t.Run("environment variable override integration", func(t *testing.T) {
		handler, err := NewErrorHandler()
		if err != nil {
			t.Fatalf("NewErrorHandler() failed during env override test: %v", err)
		}

		// Handle an error
		testErr := NewNetworkError("network context", "network cause", "network suggestion", errors.New("network test error"))
		handler.Handle(testErr)

		// Verify log file was created in custom location
		customLogPath := filepath.Join(customLogDir, "klonekit.log")
		if _, err := os.Stat(customLogPath); os.IsNotExist(err) {
			t.Error("Log file was not created in custom directory")
		}

		// Verify custom directory structure was created
		if _, err := os.Stat(customLogDir); os.IsNotExist(err) {
			t.Error("Custom log directory was not created")
		}
	})
}