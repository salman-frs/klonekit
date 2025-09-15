package ui

import (
	"strings"
	"testing"
)

func TestNewConsole(t *testing.T) {
	console := NewConsole()
	if console == nil {
		t.Fatal("NewConsole() returned nil")
	}
}

func TestConsole_formatMessage(t *testing.T) {
	console := &Console{useColors: true}

	tests := []struct {
		style    ConsoleStyle
		message  string
		expected bool // true if the result should contain color codes
	}{
		{StyleNormal, "test message", false},
		{StyleError, "error message", true},
		{StyleWarning, "warning message", true},
		{StyleSuccess, "success message", true},
		{StyleInfo, "info message", true},
	}

	for _, test := range tests {
		result := console.formatMessage(test.style, test.message)

		if test.expected {
			// Should contain color codes and reset code
			if !strings.Contains(result, test.message) {
				t.Errorf("formatMessage(%v, %q) should contain original message", test.style, test.message)
			}
			if !strings.Contains(result, colorReset) {
				t.Errorf("formatMessage(%v, %q) should contain reset code", test.style, test.message)
			}
		} else {
			// Should return original message unchanged
			if result != test.message {
				t.Errorf("formatMessage(%v, %q) = %q, want %q", test.style, test.message, result, test.message)
			}
		}
	}
}

func TestConsole_formatMessage_NoColors(t *testing.T) {
	console := &Console{useColors: false}

	result := console.formatMessage(StyleError, "test message")
	if result != "test message" {
		t.Errorf("formatMessage with useColors=false should return original message, got %q", result)
	}
}

func TestConsole_FormatErrorMessage(t *testing.T) {
	console := NewConsole()

	tests := []struct {
		context    string
		cause      string
		suggestion string
		expected   []string // parts that should be present
	}{
		{
			context:    "Test context",
			cause:      "Test cause",
			suggestion: "Test suggestion",
			expected:   []string{"Test context", "Cause: Test cause", "Suggestion: Test suggestion"},
		},
		{
			context:    "Only context",
			cause:      "",
			suggestion: "",
			expected:   []string{"Only context"},
		},
		{
			context:    "",
			cause:      "Only cause",
			suggestion: "",
			expected:   []string{"Cause: Only cause"},
		},
		{
			context:    "",
			cause:      "",
			suggestion: "Only suggestion",
			expected:   []string{"Suggestion: Only suggestion"},
		},
		{
			context:    "Context",
			cause:      "",
			suggestion: "Suggestion",
			expected:   []string{"Context", "Suggestion: Suggestion"},
		},
	}

	for _, test := range tests {
		result := console.FormatErrorMessage(test.context, test.cause, test.suggestion)

		for _, expected := range test.expected {
			if !strings.Contains(result, expected) {
				t.Errorf("FormatErrorMessage(%q, %q, %q) = %q, should contain %q",
					test.context, test.cause, test.suggestion, result, expected)
			}
		}

		// Verify the number of lines matches expected parts
		lines := strings.Split(result, "\n")
		if len(lines) != len(test.expected) {
			t.Errorf("FormatErrorMessage(%q, %q, %q) returned %d lines, want %d",
				test.context, test.cause, test.suggestion, len(lines), len(test.expected))
		}
	}
}

func TestConsole_FormatErrorMessage_Empty(t *testing.T) {
	console := NewConsole()

	result := console.FormatErrorMessage("", "", "")
	if result != "" {
		t.Errorf("FormatErrorMessage with all empty strings should return empty string, got %q", result)
	}
}

func TestStyleConstants(t *testing.T) {
	// Ensure style constants are properly defined
	styles := []ConsoleStyle{StyleNormal, StyleError, StyleWarning, StyleSuccess, StyleInfo}

	// Check that all styles have unique values
	styleMap := make(map[ConsoleStyle]bool)
	for _, style := range styles {
		if styleMap[style] {
			t.Errorf("Duplicate style value found: %d", style)
		}
		styleMap[style] = true
	}
}

func TestColorConstants(t *testing.T) {
	// Ensure color constants are not empty
	colors := map[string]string{
		"colorReset":  colorReset,
		"colorRed":    colorRed,
		"colorYellow": colorYellow,
		"colorGreen":  colorGreen,
		"colorBlue":   colorBlue,
		"colorBold":   colorBold,
	}

	for name, color := range colors {
		if color == "" {
			t.Errorf("Color constant %s is empty", name)
		}
		if !strings.HasPrefix(color, "\033[") {
			t.Errorf("Color constant %s (%q) does not start with ANSI escape sequence", name, color)
		}
	}
}