package renderer

import (
	"fmt"

	"github.com/startvibecoding/GoStreamingMarkdown/parser"
)

// Example demonstrates rendering a Markdown AST to ANSI terminal output.
func Example() {
	doc := parser.Parse("# Title", parser.DefaultOption())
	r := New(DefaultTheme(), 80)
	_ = r.Render(doc)
	// Output:
}

// ExampleDefaultTheme returns the default dark-terminal-friendly theme.
func ExampleDefaultTheme() {
	_ = DefaultTheme()
	// Output:
}

// ExampleLightTheme returns a light-terminal-friendly theme.
func ExampleLightTheme() {
	_ = LightTheme()
	// Output:
}

// ExampleAutoTheme returns the auto-detected theme.
func ExampleAutoTheme() {
	_ = AutoTheme()
	// Output:
}

// ExampleNew creates a new Renderer with the given theme and terminal width.
func ExampleNew() {
	r := New(nil, 80)
	fmt.Println(r != nil)
	// Output: true
}

// ExampleRender is a convenience function that parses and renders in one step.
func ExampleRender() {
	_ = Render("hello world", 80, nil)
	// Output:
}

// ExampleRenderer_Render renders an entire document AST to a string.
func ExampleRenderer_Render() {
	doc := parser.Parse("Hello **world**", parser.DefaultOption())
	r := New(DefaultTheme(), 80)
	_ = r.Render(doc)
	// Output:
}

// ExampleStripANSI removes all ANSI escape sequences from a string.
func ExampleStripANSI() {
	plain := StripANSI("\033[1mhello\033[0m")
	fmt.Println(plain)
	// Output: hello
}
