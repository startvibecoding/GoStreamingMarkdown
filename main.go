// GoStreamingMarkdown is a zero-dependency CLI tool that renders Markdown to styled
// ANSI terminal output. It supports both one-shot and streaming modes.
//
// Usage:
//
//	GoStreamingMarkdown [flags] [file]
//	cat README.md | GoStreamingMarkdown
//	GoStreamingMarkdown --stream --delay 50ms input.md
//
// If no file is given, reads from stdin.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"GoStreamingMarkdown/parser"
	"GoStreamingMarkdown/renderer"
)

func main() {
	args := os.Args[1:]

	// Parse flags
	filePath := ""
	streamMode := false
	delay := 0
	themeName := "auto"
	width := 0

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		case "-s", "--stream":
			streamMode = true
		case "-d", "--delay":
			if i+1 < len(args) {
				i++
				d, err := time.ParseDuration(args[i])
				if err != nil {
					fatal("invalid duration: %s", args[i])
				}
				delay = int(d)
			}
		case "-t", "--theme":
			if i+1 < len(args) {
				i++
				themeName = args[i]
			}
		case "-w", "--width":
			if i+1 < len(args) {
				i++
				w := 0
				fmt.Sscanf(args[i], "%d", &w)
				if w > 0 {
					width = w
				}
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				filePath = args[i]
			} else {
				fatal("unknown flag: %s", args[i])
			}
		}
	}

	// Detect terminal width
	if width == 0 {
		width = detectWidth()
	}

	// Select theme
	var theme *renderer.Theme
	switch themeName {
	case "dark":
		theme = renderer.DefaultTheme()
	case "light":
		theme = renderer.LightTheme()
	case "auto":
		theme = autoDetectTheme()
	default:
		theme = renderer.DefaultTheme()
	}

	// Open input
	var input io.Reader
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			fatal("cannot open %s: %v", filePath, err)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	if streamMode {
		runStream(input, width, theme, time.Duration(delay))
	} else {
		runOneShot(input, width, theme)
	}
}

func runOneShot(input io.Reader, width int, theme *renderer.Theme) {
	src, err := io.ReadAll(input)
	if err != nil {
		fatal("read error: %v", err)
	}
	doc := parser.Parse(string(src), parser.DefaultOption())
	r := renderer.New(theme, width)
	fmt.Print(r.Render(doc))
}

func runStream(input io.Reader, width int, theme *renderer.Theme, delay time.Duration) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var accumulated strings.Builder

	fmt.Print("\033[?1049h")       // enter alternate screen
	defer fmt.Print("\033[?1049l") // leave alternate screen

	for scanner.Scan() {
		line := scanner.Text()
		if accumulated.Len() > 0 {
			accumulated.WriteByte('\n')
		}
		accumulated.WriteString(line)

		// Parse with speculative rewrite (mirrors Swift StreamedMarkdownController)
		doc := parser.Parse(accumulated.String(), parser.StreamOption())
		r := renderer.New(theme, width)
		output := r.Render(doc)

		fmt.Print("\033[H\033[2J") // move cursor to top-left, clear screen
		fmt.Print(output)

		if delay > 0 {
			time.Sleep(delay)
		}
	}
	if err := scanner.Err(); err != nil {
		fatal("read error: %v", err)
	}
}

// autoDetectTheme checks COLORFGBG to detect light/dark terminal.
func autoDetectTheme() *renderer.Theme {
	fgbg := os.Getenv("COLORFGBG")
	if fgbg != "" {
		parts := strings.Split(fgbg, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			switch bg {
			case "0", "1", "2", "3", "4", "5", "6", "7":
				return renderer.DefaultTheme() // dark background
			default:
				return renderer.LightTheme() // light background
			}
		}
	}
	// Check TERM_PROGRAM for known dark-default terminals
	term := os.Getenv("TERM_PROGRAM")
	switch term {
	case "iTerm.app", "Apple_Terminal", "WezTerm", "Alacritty", "kitty", "ghostty":
		return renderer.DefaultTheme()
	}
	return renderer.DefaultTheme()
}

func detectWidth() int {
	// Default fallback; could use ioctl TIOCGWINSZ but keeping it simple
	return 80
}

func printUsage() {
	fmt.Print(`GoStreamingMarkdown — zero-dependency Markdown → ANSI terminal renderer

USAGE
  GoStreamingMarkdown [flags] [file]
  cat README.md | GoStreamingMarkdown

FLAGS
  -h, --help          Show this help
  -s, --stream        Streaming mode (renders growing document in-place)
  -d, --delay <dur>   Delay between stream updates (e.g. 50ms, 1s)
  -t, --theme <name>  Theme: auto (default), dark, or light
  -w, --width <cols>  Terminal width (default: 80)

EXAMPLES
  GoStreamingMarkdown README.md
  GoStreamingMarkdown -w 100 -t light doc.md
  cat stream.txt | GoStreamingMarkdown --stream --delay 100ms
  echo '# Hello **world**' | GoStreamingMarkdown
`)
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "GoStreamingMarkdown: "+format+"\n", args...)
	os.Exit(1)
}
