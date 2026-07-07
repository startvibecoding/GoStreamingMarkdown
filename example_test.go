package main

import (
	"github.com/startvibecoding/GoStreamingMarkdown/gsm"
)

// Example demonstrates using the library API directly.
//
// The gsm package provides the simplest way to render Markdown to ANSI
// terminal output in one call.
func Example() {
	_ = gsm.Render("# Hello **world**", 80, nil)
	// Output:
}
