// Package gsm provides streaming Markdown rendering for terminal output.
//
// This package offers a simple API for rendering Markdown to ANSI-styled terminal
// output, with support for both one-shot and streaming modes.
//
// Quick Start:
//
//	// One-shot rendering
//	output := gsm.Render("# Hello **world**", 80, nil)
//
//	// Streaming mode
//	stream := gsm.NewStream(80, nil)
//	stream.Update("partial markdown...")
//	fmt.Print(stream.Output())
package gsm

import (
	"github.com/startvibecoding/GoStreamingMarkdown/parser"
	"github.com/startvibecoding/GoStreamingMarkdown/renderer"
)

// Stream provides incremental Markdown rendering for streaming use cases.
type Stream struct {
	width    int
	theme    *renderer.Theme
	renderer *renderer.Renderer
	builder  string
}

// NewStream creates a new streaming renderer.
// width is the terminal width (use 0 for default 80).
// theme is the color theme (use nil for auto-detect).
func NewStream(width int, theme *renderer.Theme) *Stream {
	if width <= 0 {
		width = 80
	}
	if theme == nil {
		theme = renderer.DefaultTheme()
	}
	return &Stream{
		width:    width,
		theme:    theme,
		renderer: renderer.New(theme, width),
	}
}

// Update feeds new Markdown content to the stream.
// The content should be the accumulated Markdown text so far.
func (s *Stream) Update(content string) {
	s.builder = content
}

// Output renders the current accumulated Markdown to ANSI terminal output.
func (s *Stream) Output() string {
	if s.builder == "" {
		return ""
	}
	doc := parser.Parse(s.builder, parser.StreamOption())
	return s.renderer.Render(doc)
}

// Render is a convenience function for one-shot Markdown rendering.
// It parses the entire Markdown string and renders it to ANSI terminal output.
//
// Parameters:
//   - src: The Markdown source text
//   - width: Terminal width in columns (use 0 for default 80)
//   - theme: Color theme (use nil for auto-detect)
func Render(src string, width int, theme *renderer.Theme) string {
	return renderer.Render(src, width, theme)
}

// RenderWithStreamOption renders Markdown with streaming-optimized parsing.
// This is useful when rendering partial/incomplete Markdown documents.
func RenderWithStreamOption(src string, width int, theme *renderer.Theme) string {
	if width <= 0 {
		width = 80
	}
	if theme == nil {
		theme = renderer.DefaultTheme()
	}
	doc := parser.Parse(src, parser.StreamOption())
	r := renderer.New(theme, width)
	return r.Render(doc)
}

// DefaultTheme returns the default dark terminal theme.
func DefaultTheme() *renderer.Theme {
	return renderer.DefaultTheme()
}

// LightTheme returns a theme suitable for light terminal backgrounds.
func LightTheme() *renderer.Theme {
	return renderer.LightTheme()
}
