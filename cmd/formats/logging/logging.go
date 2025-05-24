// Package logging provides logging utilities for format generation
package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/leotaku/kojirou/cmd/formats"
)

var (
	// Log levels for format generation
	debugMode    = false
	colorEnabled = true
)

// EnableDebug enables debug logging
func EnableDebug(enable bool) {
	debugMode = enable
}

// EnableColor enables colored output
func EnableColor(enable bool) {
	colorEnabled = enable
	color.NoColor = !enable
}

// FormatInfo logs information about format generation
func FormatInfo(format formats.FormatType, message string) {
	prefix := ""
	if colorEnabled {
		prefix = color.BlueString("[%s]", format)
	} else {
		prefix = fmt.Sprintf("[%s]", format)
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", prefix, message)
}

// FormatSuccess logs a successful format generation
func FormatSuccess(format formats.FormatType, message string) {
	prefix := ""
	if colorEnabled {
		prefix = color.GreenString("[%s]", format)
	} else {
		prefix = fmt.Sprintf("[%s]", format)
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", prefix, message)
}

// FormatError logs an error during format generation
func FormatError(format formats.FormatType, err error) {
	prefix := ""
	if colorEnabled {
		prefix = color.RedString("[%s]", format)
	} else {
		prefix = fmt.Sprintf("[%s]", format)
	}
	fmt.Fprintf(os.Stderr, "%s Error: %v\n", prefix, err)
}

// FormatDebug logs debug information if debug mode is enabled
func FormatDebug(format formats.FormatType, message string) {
	if !debugMode {
		return
	}

	prefix := ""
	if colorEnabled {
		prefix = color.YellowString("[%s]", format)
	} else {
		prefix = fmt.Sprintf("[%s]", format)
	}
	fmt.Fprintf(os.Stderr, "%s DEBUG: %s\n", prefix, message)
}

// TimedOperation executes a function and logs the time it took
func TimedOperation(formatType formats.FormatType, operation string, fn func() error) error {
	if debugMode {
		FormatDebug(formatType, fmt.Sprintf("Starting %s", operation))
	}

	start := time.Now()
	err := fn()
	elapsed := time.Since(start)

	if err != nil {
		FormatError(formatType, fmt.Errorf("%s: %w (took %s)", operation, err, elapsed))
		return err
	}

	if debugMode {
		FormatDebug(formatType, fmt.Sprintf("Completed %s in %s", operation, elapsed))
	}

	return nil
}

// FormatSummary logs a summary of format generation
func FormatSummary(formatStatuses map[formats.FormatType]string) {
	var successFormats, errorFormats, skippedFormats []string

	for format, status := range formatStatuses {
		if strings.HasPrefix(status, "Error") {
			errorFormats = append(errorFormats, string(format))
		} else if strings.HasPrefix(status, "Skipped") {
			skippedFormats = append(skippedFormats, string(format))
		} else {
			successFormats = append(successFormats, string(format))
		}
	}

	if len(successFormats) > 0 {
		if colorEnabled {
			fmt.Fprintf(os.Stderr, "%s %s\n",
				color.GreenString("✓ Success:"),
				strings.Join(successFormats, ", "))
		} else {
			fmt.Fprintf(os.Stderr, "✓ Success: %s\n",
				strings.Join(successFormats, ", "))
		}
	}

	if len(skippedFormats) > 0 {
		if colorEnabled {
			fmt.Fprintf(os.Stderr, "%s %s\n",
				color.BlueString("↷ Skipped:"),
				strings.Join(skippedFormats, ", "))
		} else {
			fmt.Fprintf(os.Stderr, "↷ Skipped: %s\n",
				strings.Join(skippedFormats, ", "))
		}
	}

	if len(errorFormats) > 0 {
		if colorEnabled {
			fmt.Fprintf(os.Stderr, "%s %s\n",
				color.RedString("✗ Errors:"),
				strings.Join(errorFormats, ", "))
		} else {
			fmt.Fprintf(os.Stderr, "✗ Errors: %s\n",
				strings.Join(errorFormats, ", "))
		}
	}
}
