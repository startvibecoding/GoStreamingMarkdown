# GoStreamingMarkdown Library Usage

This example demonstrates how to use GoStreamingMarkdown as a library in your Go projects.

## Installation

```bash
go get github.com/startvibecoding/GoStreamingMarkdown
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/startvibecoding/GoStreamingMarkdown/gsm"
)

func main() {
    // One-shot rendering
    output := gsm.Render("# Hello **world**", 80, nil)
    fmt.Println(output)
}
```

## API Reference

### One-shot Rendering

```go
func Render(src string, width int, theme *renderer.Theme) string
```

Render Markdown to ANSI terminal output in one call.

- `src`: Markdown source text
- `width`: Terminal width in columns (0 for default 80)
- `theme`: Color theme (nil for auto-detect)

### Streaming Rendering

```go
stream := gsm.NewStream(80, nil)

// Feed content incrementally
stream.Update("partial markdown...")

// Get rendered output
output := stream.Output()
```

### Themes

```go
// Dark theme (default)
theme := gsm.DefaultTheme()

// Light theme
theme := gsm.LightTheme()
```

## Examples

### Basic Usage

```go
output := gsm.Render("# Hello", 80, nil)
fmt.Print(output)
```

### Streaming Mode

```go
stream := gsm.NewStream(80, nil)

for chunk := range incomingMarkdown {
    stream.Update(chunk)
    clearScreen()
    fmt.Print(stream.Output())
}
```

### Custom Width

```go
// Narrow terminal (40 columns)
output := gsm.Render(markdown, 40, nil)

// Wide terminal (120 columns)
output := gsm.Render(markdown, 120, nil)
```

## Running the Example

```bash
cd examples/library-usage
go run main.go
```
