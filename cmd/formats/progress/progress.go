// Package progress provides progress tracking functionality
package progress

import (
	"fmt"
	"io"

	"github.com/cheggaaa/pb/v3"
)

const (
	progressTemplate = `` +
		`{{ string . "prefix" | printf "%-12v" }}` +
		`{{ if string . "format" }}` +
		`[{{ string . "format" | printf "%-5v" }}]` +
		`{{ else }}` +
		`{{ printf "%-7v" "" }}` +
		`{{ end }}` +
		`{{ bar . "|" "█" "▌" " " "|" }}` + `{{ " " }}` +
		`{{ if string . "message" }}` +
		`{{   string . "message" | printf "%-15v" }}` +
		`{{ else }}` +
		`{{   counters . | printf "%-15v" }}` +
		`{{ end }}` + `{{ " |" }}`
)

type Progress interface {
	Increase(int)
	Add(int)
	NewProxyWriter(io.Writer) io.Writer
}

type CliProgress struct {
	bar       *pb.ProgressBar
	firstCall bool
}

func (p CliProgress) Increase(n int) {
	p.bar.AddTotal(int64(n))
}

func (p CliProgress) Add(n int) {
	p.bar.Add(n)
}

func (p CliProgress) NewProxyWriter(w io.Writer) io.Writer {
	return p.bar.NewProxyWriter(w)
}

func (p CliProgress) Done() {
	p.bar.Finish()
}

// SetFormat sets the format indicator in the progress bar
func (p *CliProgress) SetFormat(format string) {
	p.bar.Set("format", format)
}

// Cancel cancels the progress bar with a message
func (p *CliProgress) Cancel(message string) {
	p.bar.Set("message", message)
	p.bar.SetTotal(1).SetCurrent(1)
	p.Done()
}

// CancelWithFormat cancels the progress bar with a format-specific message
func (p *CliProgress) CancelWithFormat(format, message string) {
	p.SetFormat(format)
	p.Cancel(message)
}

// SetFormatMessage sets a message for the current format
func (p *CliProgress) SetFormatMessage(format, message string) {
	p.SetFormat(format)
	p.bar.Set("message", message)
}

// TitledProgress creates a new progress bar with a title
func TitledProgress(title string) CliProgress {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Start()

	return CliProgress{bar, true}
}

// FormatTitledProgress creates a new progress bar with a title and format indicator
func FormatTitledProgress(title string, format string) CliProgress {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Set("format", format)
	bar.Start()

	return CliProgress{bar, true}
}

// VanishingProgress creates a new progress bar that disappears when complete
func VanishingProgress(title string) CliProgress {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()

	return CliProgress{bar, true}
}

// FormatVanishingProgress creates a new progress bar with a format that disappears when complete
func FormatVanishingProgress(title string, format string) CliProgress {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Set("format", format)
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()

	return CliProgress{bar, true}
}

// MultiFormatStatusProgress creates a progress bar for tracking multiple formats
// and displays a final status message
func MultiFormatStatusProgress(title string, formats []string) CliProgress {
	bar := pb.New(len(formats)).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Start()

	return CliProgress{bar, true}
}

// FormatCompleted marks a format as completed in a multi-format progress bar
func (p *CliProgress) FormatCompleted(format string, status string) {
	currentMsg := fmt.Sprintf("%s: %s", format, status)
	prevMsg, hasPrevMsg := p.bar.Get("message").(string)

	if hasPrevMsg && prevMsg != "" {
		p.bar.Set("message", fmt.Sprintf("%s, %s", prevMsg, currentMsg))
	} else {
		p.bar.Set("message", currentMsg)
	}

	p.bar.Increment()
}
