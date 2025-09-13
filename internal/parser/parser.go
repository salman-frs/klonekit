package parser

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"klonekit/pkg/blueprint"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Parse reads and validates a blueprint YAML file, returning the parsed Blueprint struct or an error.
func Parse(filePath string) (*blueprint.Blueprint, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("blueprint file not found: %s", filePath)
	}

	// Configure Viper
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	// Read the file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("blueprint file not found: %s", filePath)
		}
		return nil, fmt.Errorf("failed to read blueprint file: %w", err)
	}

	// Unmarshal into Blueprint struct
	var bp blueprint.Blueprint
	if err := v.Unmarshal(&bp); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint file - malformed YAML: %w", err)
	}

	// Validate the structure
	if err := validate.Struct(&bp); err != nil {
		return nil, formatValidationError(err)
	}

	return &bp, nil
}

// formatValidationError converts validator errors into user-friendly messages.
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMessages []string
		for _, e := range validationErrors {
			errorMessages = append(errorMessages, formatFieldError(e))
		}

		if len(errorMessages) == 1 {
			return fmt.Errorf("validation error: %s", errorMessages[0])
		}

		result := "validation errors:\n"
		for _, msg := range errorMessages {
			result += fmt.Sprintf("  - %s\n", msg)
		}
		return fmt.Errorf("%s", result)
	}
	return fmt.Errorf("validation failed: %w", err)
}

// formatFieldError formats a single validation error into a user-friendly message.
func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("field '%s' is required but missing", field)
	case "eq":
		return fmt.Sprintf("field '%s' must be '%s'", field, e.Param())
	case "oneof":
		return fmt.Sprintf("field '%s' must be one of: %s", field, e.Param())
	case "url":
		return fmt.Sprintf("field '%s' must be a valid URL", field)
	default:
		return fmt.Sprintf("field '%s' failed validation (%s)", field, tag)
	}
}
