package gsm

import (
	"fmt"
)

// Example demonstrates one-shot Markdown rendering with the top-level
// Render function.
func Example() {
	_ = Render("# Hello **world**", 80, nil)
	// Output:
}

// ExampleRender demonstrates rendering Markdown inline formatting.
func ExampleRender() {
	_ = Render("**bold** and *italic* text", 80, nil)
	// Output:
}

// ExampleNewStream demonstrates creating a new streaming renderer.
func ExampleNewStream() {
	stream := NewStream(80, nil)
	fmt.Println(stream != nil)
	// Output: true
}

// ExampleStream_Update demonstrates feeding new Markdown content to the stream.
func ExampleStream_Update() {
	stream := NewStream(80, nil)
	stream.Update("partial ")
	stream.Update("markdown")
	_ = stream.Output()
	// Output:
}

// ExampleStream_Output demonstrates rendering the accumulated content.
func ExampleStream_Output() {
	stream := NewStream(80, nil)
	stream.Update("# Heading")
	_ = stream.Output()
	// Output:
}

// ExampleRenderWithStreamOption demonstrates rendering with speculative rewrite
// enabled, which is useful for partial/incomplete Markdown documents.
func ExampleRenderWithStreamOption() {
	_ = RenderWithStreamOption("# Hello **wor", 80, nil)
	// Output:
}

// ExampleDefaultTheme returns the default dark terminal theme.
func ExampleDefaultTheme() {
	_ = DefaultTheme()
	// Output:
}

// ExampleLightTheme returns a theme suitable for light terminal backgrounds.
func ExampleLightTheme() {
	_ = LightTheme()
	// Output:
}
