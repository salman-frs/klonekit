package ui

import (
	"fmt"
	"os"
	"strings"
)

type ConsoleStyle int

const (
	StyleNormal ConsoleStyle = iota
	StyleError
	StyleWarning
	StyleSuccess
	StyleInfo
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
)

type Console struct {
	useColors bool
}

func NewConsole() *Console {
	return &Console{
		useColors: isTerminal(),
	}
}

func isTerminal() bool {
	stat, _ := os.Stderr.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func (c *Console) formatMessage(style ConsoleStyle, message string) string {
	if !c.useColors {
		return message
	}

	var color string
	switch style {
	case StyleError:
		color = colorRed + colorBold
	case StyleWarning:
		color = colorYellow
	case StyleSuccess:
		color = colorGreen
	case StyleInfo:
		color = colorBlue
	default:
		return message
	}

	return color + message + colorReset
}

func (c *Console) PrintError(message string) {
	fmt.Fprintf(os.Stderr, "%s\n", c.formatMessage(StyleError, "Error: "+message))
}

func (c *Console) PrintWarning(message string) {
	fmt.Fprintf(os.Stderr, "%s\n", c.formatMessage(StyleWarning, "Warning: "+message))
}

func (c *Console) PrintSuccess(message string) {
	fmt.Printf("%s\n", c.formatMessage(StyleSuccess, message))
}

func (c *Console) PrintInfo(message string) {
	fmt.Printf("%s\n", c.formatMessage(StyleInfo, message))
}

func (c *Console) FormatErrorMessage(context, cause, suggestion string) string {
	var parts []string

	if context != "" {
		parts = append(parts, context)
	}

	if cause != "" {
		parts = append(parts, fmt.Sprintf("Cause: %s", cause))
	}

	if suggestion != "" {
		parts = append(parts, fmt.Sprintf("Suggestion: %s", suggestion))
	}

	return strings.Join(parts, "\n")
}