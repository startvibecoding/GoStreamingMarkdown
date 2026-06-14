// Example: Using GoStreamingMarkdown as a library
package main

import (
	"fmt"
	"strings"
	"time"

	"GoStreamingMarkdown/gsm"
)

func main() {
	fmt.Println("=== GoStreamingMarkdown Library Usage Examples ===")
	fmt.Println()

	// Example 1: One-shot rendering
	exampleOneShot()

	// Example 2: Streaming rendering
	exampleStreaming()

	// Example 3: Custom theme
	exampleCustomTheme()
}

func exampleOneShot() {
	fmt.Println("--- Example 1: One-shot Rendering ---")

	markdown := `# Hello World

This is a **bold** and *italic* text.

` + "```go" + `
func main() {
    fmt.Println("Hello!")
}
` + "```" + `

- Item 1
- Item 2
- Item 3
`

	// Render with default theme, 80 columns width
	output := gsm.Render(markdown, 80, nil)
	fmt.Println(output)
}

func exampleStreaming() {
	fmt.Println("--- Example 2: Streaming Rendering ---")
	fmt.Println("(Simulating incremental content arrival)")
	fmt.Println()

	stream := gsm.NewStream(80, nil)

	// Simulate streaming content
	lines := []string{
		"# Streaming Demo",
		"",
		"This text appears ",
		"This text appears **word** ",
		"This text appears **word** by ",
		"This text appears **word** by word.",
		"",
		"```go",
		"func main() {",
		"    fmt.Println(\"Hello!\")",
		"}",
		"```",
	}

	for _, line := range lines {
		stream.Update(line)
		output := stream.Output()
		// In a real application, you would clear and redraw
		fmt.Print("\033[H\033[2J") // Clear screen
		fmt.Println(output)
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println()
}

func exampleCustomTheme() {
	fmt.Println("--- Example 3: Custom Themes ---")

	markdown := `# Theme Demo

` + "```" + `
Code block
` + "```" + `

> Blockquote with **bold** text
`

	// Dark theme (default)
	fmt.Println("Dark theme:")
	fmt.Println(gsm.Render(markdown, 60, gsm.DefaultTheme()))

	// Light theme
	fmt.Println("Light theme:")
	fmt.Println(gsm.Render(markdown, 60, gsm.LightTheme()))
}

// Helper to repeat strings
func repeat(s string, n int) string {
	return strings.Repeat(s, n)
}
