package errors

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"klonekit/internal/ui"
)

type ErrorHandler struct {
	logger  *slog.Logger
	console *ui.Console
}

func NewErrorHandler() (*ErrorHandler, error) {
	logFile, err := createLogFile()
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	console := ui.NewConsole()

	return &ErrorHandler{
		logger:  logger,
		console: console,
	}, nil
}

// getOSStandardLogDir returns the OS-standard log directory path
func getOSStandardLogDir() (string, error) {
	// Check for environment variable override first
	if customLogDir := os.Getenv("KLONEKIT_LOG_DIR"); customLogDir != "" {
		return customLogDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Logs/KloneKit/
		return filepath.Join(homeDir, "Library", "Logs", "KloneKit"), nil
	case "linux", "freebsd", "openbsd", "netbsd":
		// Linux/Unix: ~/.local/share/klonekit/logs/ (XDG Base Directory)
		return filepath.Join(homeDir, ".local", "share", "klonekit", "logs"), nil
	case "windows":
		// Windows: %APPDATA%\KloneKit\logs\
		appDataDir := os.Getenv("APPDATA")
		if appDataDir == "" {
			return filepath.Join(homeDir, "AppData", "Roaming", "KloneKit", "logs"), nil
		}
		return filepath.Join(appDataDir, "KloneKit", "logs"), nil
	default:
		// Fallback for unknown OS
		return filepath.Join(homeDir, ".klonekit", "logs"), nil
	}
}

// createLogDirectoryWithFallback creates the log directory with fallback to current directory
func createLogDirectoryWithFallback() (string, bool, error) {
	var warnings []string
	var fallbackUsed bool

	// Try OS-standard directory first
	logDir, err := getOSStandardLogDir()
	if err == nil {
		if err := os.MkdirAll(logDir, 0750); err == nil {
			// Check if we can write to the directory
			testFile := filepath.Join(logDir, ".test_write")
			if f, testErr := os.Create(testFile); testErr == nil {
				if err := f.Close(); err != nil {
					slog.Warn("Failed to close test file", "path", testFile, "error", err)
				}
				if err := os.Remove(testFile); err != nil {
					slog.Warn("Failed to remove test file", "path", testFile, "error", err)
				}
				return logDir, fallbackUsed, nil
			}
		}
		warnings = append(warnings, fmt.Sprintf("Cannot access standard log directory %s: %v", logDir, err))
	} else {
		warnings = append(warnings, fmt.Sprintf("Cannot determine standard log directory: %v", err))
	}

	// Fallback to current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", true, fmt.Errorf("cannot determine current directory for fallback logging: %w", err)
	}

	fallbackUsed = true
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %s. Falling back to current directory for logging.\n", warnings[0])
	}

	return currentDir, fallbackUsed, nil
}

// rotateLogFile rotates log files when size limit is exceeded
func rotateLogFile(logPath string) error {
	const maxFiles = 5

	// Rotate existing files (.4 -> .5, .3 -> .4, etc.)
	for i := maxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", logPath, i)
		newPath := fmt.Sprintf("%s.%d", logPath, i+1)

		if i == maxFiles-1 {
			// Remove the oldest file
			if _, err := os.Stat(oldPath); err == nil {
				if err := os.Remove(oldPath); err != nil {
					slog.Warn("Failed to remove old log file", "path", oldPath, "error", err)
				}
			}
		} else {
			// Rotate file
			if _, err := os.Stat(oldPath); err == nil {
				if err := os.Rename(oldPath, newPath); err != nil {
					slog.Warn("Failed to rotate log file", "old", oldPath, "new", newPath, "error", err)
				}
			}
		}
	}

	// Move current log to .1
	if _, err := os.Stat(logPath); err == nil {
		return os.Rename(logPath, logPath+".1")
	}

	return nil
}

// checkLogRotation checks if log rotation is needed and performs it
func checkLogRotation(logPath string) error {
	const maxSizeBytes = 10 * 1024 * 1024 // 10MB

	info, err := os.Stat(logPath)
	if err != nil {
		// File doesn't exist or other error, no rotation needed
		return nil
	}

	if info.Size() >= maxSizeBytes {
		return rotateLogFile(logPath)
	}

	return nil
}

func createLogFile() (*os.File, error) {
	logDir, _, err := createLogDirectoryWithFallback()
	if err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFileName := "klonekit.log"

	logPath := filepath.Join(logDir, logFileName)

	// Check if log rotation is needed before opening the file
	if err := checkLogRotation(logPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to rotate log file: %v\n", err)
	}

	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
}

func (h *ErrorHandler) Handle(err error) {
	if err == nil {
		return
	}

	var kloneKitErr *KloneKitError
	if errors.As(err, &kloneKitErr) {
		h.handleKloneKitError(kloneKitErr)
	} else {
		h.handleGenericError(err)
	}
}

func (h *ErrorHandler) handleKloneKitError(err *KloneKitError) {
	h.logStructuredError(err)

	message := h.console.FormatErrorMessage(err.Context, err.Cause, err.Suggestion)
	h.console.PrintError(message)
}

func (h *ErrorHandler) handleGenericError(err error) {
	h.logger.Error("Unhandled error occurred",
		"error", err.Error(),
		"type", "generic",
	)

	h.console.PrintError(err.Error())
}

func (h *ErrorHandler) logStructuredError(err *KloneKitError) {
	logAttrs := []slog.Attr{
		slog.String("error", err.OriginalErr.Error()),
		slog.String("type", getErrorTypeName(err.Type)),
		slog.String("context", err.Context),
	}

	if err.Cause != "" {
		logAttrs = append(logAttrs, slog.String("cause", err.Cause))
	}

	if err.Suggestion != "" {
		logAttrs = append(logAttrs, slog.String("suggestion", err.Suggestion))
	}

	h.logger.LogAttrs(context.TODO(), slog.LevelError, "KloneKit error occurred", logAttrs...)
}

func getErrorTypeName(errType error) string {
	switch errType {
	case ErrBlueprintNotFound:
		return "blueprint_not_found"
	case ErrBlueprintParseFailed:
		return "blueprint_parse_failed"
	case ErrScaffoldFailed:
		return "scaffold_failed"
	case ErrSCMFailed:
		return "scm_failed"
	case ErrProvisionFailed:
		return "provision_failed"
	case ErrRuntimeFailed:
		return "runtime_failed"
	case ErrConfigInvalid:
		return "config_invalid"
	case ErrNetworkFailed:
		return "network_failed"
	case ErrFileSystemFailed:
		return "filesystem_failed"
	default:
		return "unknown"
	}
}