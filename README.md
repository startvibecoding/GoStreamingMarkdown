# GoStreamingMarkdown

[中文文档](README_zh.md)

**Zero-dependency** Go CLI streaming Markdown → ANSI terminal renderer.

Refactored from the Swift project [SwiftStreamingMarkdown](https://github.com/SwiftStreamingMarkdown),
fully implementing the Preprocess → Parse → Rewrite → Render four-stage pipeline. Supports streaming rendering mode.

## Features

| Category | Support |
|----------|---------|
| **Headings** | ATX `# ~ ######` with colored levels |
| **Emphasis** | `*italic*`, `**bold**`, `~~strikethrough~~` |
| **Code** | Inline `` `code` ``, fenced code blocks ` ```lang ``` ` with language label |
| **Lists** | Ordered `1.`, unordered `- * +`, task lists `- [x]` |
| **Blockquotes** | `> blockquote` with nesting support |
| **Tables** | `\| a \| b \|` GFM tables with borders |
| **Links** | `[text](url)`, `<autolink>` |
| **Images** | `![alt](url)` terminal rendering |
| **Horizontal Rules** | `---`, `***`, `___` |
| **LaTeX Math** | `$$block$$`, `$inline$`, `\(paren\)` preprocessed to code blocks/inline code |
| **Streaming** | `--stream` mode with speculative emphasis closure (anti-jitter) |

## Installation

### As a CLI tool

```bash
go install github.com/startvibecoding/GoStreamingMarkdown@latest
```

### As a library

```bash
go get github.com/startvibecoding/GoStreamingMarkdown
```

### Build from source

```bash
git clone https://github.com/startvibecoding/GoStreamingMarkdown.git
cd GoStreamingMarkdown
go build -o GoStreamingMarkdown .
```

**Zero external dependencies** — uses only Go standard library.

## Usage

```bash
# Render a file
GoStreamingMarkdown README.md

# Read from pipe
cat README.md | GoStreamingMarkdown

# Streaming mode (real-time rendering, incremental updates)
cat stream.txt | GoStreamingMarkdown --stream --delay 100ms

# Specify theme and width
GoStreamingMarkdown -t light -w 100 doc.md

# Render Markdown string
echo '# Hello **world**' | GoStreamingMarkdown
```

### Flags

| Flag | Description |
|------|-------------|
| `-h, --help` | Show help |
| `-s, --stream` | Streaming mode (in-place redraw) |
| `-d, --delay <dur>` | Stream update interval (e.g. `50ms`, `1s`) |
| `-t, --theme <name>` | Theme: `dark` (default) or `light` |
| `-w, --width <cols>` | Terminal width (default: 80) |

## Architecture

Complete replication of the Swift project's four-stage pipeline:

```
Input Text
  │
  ▼
┌─────────────────────┐
│ 1. Preprocess (LaTeX)│  $$...$$ → fenced code block
│                      │  $...$   → inline code
│                      │  \(...\) → inline code
└──────────┬──────────┘
           ▼
┌─────────────────────┐
│ 2. Parse (AST)       │  CommonMark compatible parser
│                      │  Block: heading, paragraph, code,
│                      │  blockquote, list, table, hr
│                      │  Inline: emphasis, strong, code,
│                      │  link, image, strikethrough, autolink
└──────────┬──────────┘
           ▼
┌─────────────────────┐
│ 3. Rewrite           │  Speculative emphasis closure
│ (speculative)        │  Prevents text jitter during streaming
└──────────┬──────────┘
           ▼
┌─────────────────────┐
│ 4. Render (ANSI)     │  AST → ANSI terminal output
│                      │  Dark/Light themes
│                      │  Unicode box-drawing characters
└─────────────────────┘
```

## Project Structure

```
GoStreamingMarkdown/
├── main.go              # CLI entry point
├── go.mod               # Go module definition (zero dependencies)
├── gsm/                 # Convenience API package
│   └── gsm.go           # Simplified streaming rendering API
├── parser/
│   ├── node.go          # AST node type definitions
│   └── parser.go        # Markdown parser + LaTeX preprocessor + rewriter
├── renderer/
│   └── renderer.go      # ANSI terminal renderer + theme system
├── examples/            # Usage examples
└── README.md
```

## Library Usage

### Option 1: Using gsm convenience package (Recommended)

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
    
    // Streaming rendering
    stream := gsm.NewStream(80, nil)
    stream.Update("partial markdown...")
    fmt.Print(stream.Output())
}
```

### Option 2: Using parser/renderer directly

```go
package main

import (
    "fmt"
    "github.com/startvibecoding/GoStreamingMarkdown/parser"
    "github.com/startvibecoding/GoStreamingMarkdown/renderer"
)

func main() {
    src := "# Hello **world**"
    
    // Parse
    doc := parser.Parse(src, parser.DefaultOption())
    
    // Render
    theme := renderer.DefaultTheme()
    r := renderer.New(theme, 80)
    fmt.Println(r.Render(doc))
    
    // Or one-step
    fmt.Println(renderer.Render(src, 80, theme))
}
```

### Streaming Parsing

```go
opt := parser.StreamOption() // Enable speculative rewrite
doc := parser.Parse(partialText, opt)
```

For more examples, see `examples/library-usage/`

## Swift to Go Mapping

| Swift Component | Go Equivalent |
|----------------|---------------|
| `LaTexPreProcessor` | `parser.preprocessLaTeX()` |
| `swift-markdown` (cmark-gfm) | `parser.Parse()` |
| `PartialEmphasisRewriter` | `parser.rewriteSpeculative()` |
| `MarkdownRenderable` | `parser.Node` AST |
| `DocumentView` | `renderer.Render()` |
| `MarkdownRenderConfig` + `Colors` | `renderer.Theme` |
| `StreamedMarkdownSource` | `--stream` CLI mode |

## License

MIT
