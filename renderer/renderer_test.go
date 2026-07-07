package renderer

import (
	"strings"
	"testing"

	"github.com/startvibecoding/GoStreamingMarkdown/parser"
)

// ── Helpers ─────────────────────────────────────────────────────────────────

func render(src string, width int) string {
	return Render(src, width, nil)
}

func renderStripped(src string, width int) string {
	return StripANSI(render(src, width))
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected %q to NOT contain %q", s, substr)
	}
}

func assertHasLine(t *testing.T, s, line string) {
	t.Helper()
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, line) {
			return
		}
	}
	t.Errorf("expected output to have line containing %q", line)
}

func assertNoBackgroundANSI(t *testing.T, s string) {
	t.Helper()
	backgroundCodes := []string{
		"\033[40m", "\033[41m", "\033[42m", "\033[43m",
		"\033[44m", "\033[45m", "\033[46m", "\033[47m",
		"\033[48;", "\033[100m", "\033[101m", "\033[102m",
		"\033[103m", "\033[104m", "\033[105m", "\033[106m",
		"\033[107m",
	}
	for _, code := range backgroundCodes {
		if strings.Contains(s, code) {
			t.Fatalf("expected no background ANSI code %q in %q", code, s)
		}
	}
}

func assertTableLinesWithinWidth(t *testing.T, s string, width int) {
	t.Helper()
	for _, line := range strings.Split(s, "\n") {
		if !strings.ContainsAny(line, "│┌┬┐├┼┤└┴┘") {
			continue
		}
		if w := visualWidth(line); w > width {
			t.Fatalf("table line exceeds width %d: got %d: %q", width, w, line)
		}
	}
}

func assertTableLinesEqualWidth(t *testing.T, s string) {
	t.Helper()
	expected := -1
	for _, line := range strings.Split(s, "\n") {
		if !strings.ContainsAny(line, "│┌┬┐├┼┤└┴┘") {
			continue
		}
		w := visualWidth(line)
		if expected < 0 {
			expected = w
			continue
		}
		if w != expected {
			t.Fatalf("table line width mismatch: expected %d, got %d: %q", expected, w, line)
		}
	}
}

func assertBoxLinesEqualWidth(t *testing.T, s string, width int) {
	t.Helper()
	for _, line := range strings.Split(s, "\n") {
		if !strings.ContainsAny(line, "│┌├└┐┤┘") {
			continue
		}
		if w := visualWidth(line); w != width {
			t.Fatalf("box line width mismatch: expected %d, got %d: %q", width, w, line)
		}
	}
}

// ── Heading Rendering ───────────────────────────────────────────────────────

func TestRenderHeadingH1(t *testing.T) {
	out := renderStripped("# Hello", 80)
	assertContains(t, out, "# Hello")
}

func TestRenderHeadingAllLevels(t *testing.T) {
	src := "# H1\n\n## H2\n\n### H3\n\n#### H4\n\n##### H5\n\n###### H6"
	out := renderStripped(src, 80)
	assertContains(t, out, "# H1")
	assertContains(t, out, "## H2")
	assertContains(t, out, "### H3")
	assertContains(t, out, "#### H4")
	assertContains(t, out, "##### H5")
	assertContains(t, out, "###### H6")
}

func TestRenderHeadingANSI(t *testing.T) {
	out := render("# Hello", 80)
	// Should contain ANSI codes
	if !strings.Contains(out, "\033[") {
		t.Error("expected ANSI codes in heading output")
	}
}

// ── Paragraph Rendering ─────────────────────────────────────────────────────

func TestRenderParagraph(t *testing.T) {
	out := renderStripped("Hello world", 80)
	assertContains(t, out, "Hello world")
}

func TestRenderParagraphMultiple(t *testing.T) {
	out := renderStripped("First.\n\nSecond.\n\nThird.", 80)
	assertContains(t, out, "First.")
	assertContains(t, out, "Second.")
	assertContains(t, out, "Third.")
}

func TestRenderParagraphWrapping(t *testing.T) {
	src := "This is a long paragraph that should wrap at the terminal width boundary when it exceeds the configured column limit"
	out := renderStripped(src, 40)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Errorf("expected wrapping into multiple lines, got %d line(s)", len(lines))
	}
	// No line should exceed width (excluding trailing newline)
	for _, line := range lines {
		w := visualWidth(line)
		if w > 40 {
			t.Errorf("line exceeds width 40: %d chars: %q", w, line)
		}
	}
}

func TestRenderParagraphWrappingPreservesANSI(t *testing.T) {
	src := "Hello **bold text that should wrap across line boundaries** world"
	out := render(src, 40)
	// Bold ANSI code should appear on wrapped line
	if !strings.Contains(out, "\033[1m") {
		t.Error("expected bold ANSI code")
	}
}

func TestRenderParagraphLineSpacing(t *testing.T) {
	src := "First paragraph.\n\nSecond paragraph."
	out := render(src, 80)
	// With default LineSpacing=1, there should be extra blank lines
	assertContains(t, out, "First paragraph.")
	assertContains(t, out, "Second paragraph.")
}

// ── Code Block Rendering ────────────────────────────────────────────────────

func TestRenderFencedCodeBlock(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "fmt.Println")
	assertContains(t, out, "go") // language label
	assertContains(t, out, "┌")  // box top
	assertContains(t, out, "└")  // box bottom
	assertContains(t, out, "│")  // box sides
}

func TestRenderCodeBlockNoLanguage(t *testing.T) {
	src := "```\nhello\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "hello")
	assertContains(t, out, "┌")
}

func TestRenderCodeBlockNoLanguageCJKWidth(t *testing.T) {
	src := "```\n中文内容\n```"
	out := renderStripped(src, 40)
	assertContains(t, out, "中文内容")
	assertBoxLinesEqualWidth(t, out, 40)
}

func TestRenderCodeBlockMultiline(t *testing.T) {
	src := "```python\ndef hello():\n    print('hello')\n    return True\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "def hello():")
	assertContains(t, out, "    print('hello')")
	assertContains(t, out, "    return True")
	assertContains(t, out, "python")
}

func TestRenderCodeBlockPreservesIndentation(t *testing.T) {
	src := "```go\nfunc main() {\n    if true {\n        fmt.Println(\"nested\")\n    }\n}\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "    if true {")
	assertContains(t, out, "        fmt.Println")
}

func TestRenderCodeBlockWithBlankLines(t *testing.T) {
	src := "```go\nline1\n\nline2\n\nline3\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "line1")
	assertContains(t, out, "line2")
	assertContains(t, out, "line3")
}

func TestRenderCodeBlockSpecialCharacters(t *testing.T) {
	src := "```\n<script>alert('xss')</script>\n| not | table |\n> not quote\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "<script>")
	assertContains(t, out, "| not | table |")
	assertContains(t, out, "> not quote")
}

func TestRenderIndentedCodeBlock(t *testing.T) {
	src := "    indented code\n    second line"
	out := renderStripped(src, 80)
	assertContains(t, out, "indented code")
	assertContains(t, out, "second line")
	assertContains(t, out, "┌") // should be boxed
}

func TestRenderCodeBlockWidthRespected(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```"
	out := renderStripped(src, 50)
	assertContains(t, out, "┌")
	assertBoxLinesEqualWidth(t, out, 50)
}

func TestRenderCodeBlockCJKWidthRespected(t *testing.T) {
	src := "```text\n中文代码块\n```"
	out := renderStripped(src, 50)
	assertContains(t, out, "中文代码块")
	assertBoxLinesEqualWidth(t, out, 50)
}

func TestRenderMultipleCodeBlocks(t *testing.T) {
	src := "```go\ncode1\n```\n\nSome text\n\n```python\ncode2\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "code1")
	assertContains(t, out, "code2")
	assertContains(t, out, "go")
	assertContains(t, out, "python")
}

func TestRenderCodeBlockLanguageVariants(t *testing.T) {
	languages := []string{"go", "python", "js", "bash", "sql", "yaml", "json", "html"}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			src := "```" + lang + "\ncode\n```"
			out := renderStripped(src, 80)
			assertContains(t, out, lang)
		})
	}
}

func TestRenderCodeBlockLongLine(t *testing.T) {
	longLine := strings.Repeat("x", 200)
	src := "```\n" + longLine + "\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "x")
	assertBoxLinesEqualWidth(t, out, 80)
}

// ── Blockquote Rendering ────────────────────────────────────────────────────

func TestRenderBlockquoteSimple(t *testing.T) {
	out := renderStripped("> Hello world", 80)
	assertContains(t, out, "│")
	assertContains(t, out, "Hello world")
}

func TestRenderBlockquoteWithFormatting(t *testing.T) {
	out := renderStripped("> **bold** and *italic*", 80)
	assertContains(t, out, "│")
	// Formatting should be rendered inside blockquote
	assertContains(t, out, "bold")
	assertContains(t, out, "italic")
}

func TestRenderBlockquoteNested(t *testing.T) {
	out := renderStripped("> L1\n> > L2", 80)
	assertContains(t, out, "L1")
	assertContains(t, out, "L2")
	// Nested should have double bar
	assertContains(t, out, "│ │")
}

func TestRenderBlockquoteMultiline(t *testing.T) {
	out := renderStripped("> Line 1\n> Line 2\n> Line 3", 80)
	assertContains(t, out, "Line 1")
	assertContains(t, out, "Line 2")
	assertContains(t, out, "Line 3")
}

// ── List Rendering ──────────────────────────────────────────────────────────

func TestRenderUnorderedList(t *testing.T) {
	out := renderStripped("- Item 1\n- Item 2\n- Item 3", 80)
	assertContains(t, out, "•")
	assertContains(t, out, "Item 1")
	assertContains(t, out, "Item 2")
	assertContains(t, out, "Item 3")
}

func TestRenderOrderedList(t *testing.T) {
	out := renderStripped("1. First\n2. Second\n3. Third", 80)
	assertContains(t, out, "1.")
	assertContains(t, out, "2.")
	assertContains(t, out, "3.")
	assertContains(t, out, "First")
}

func TestRenderOrderedListStartNumber(t *testing.T) {
	out := renderStripped("5. Fifth\n6. Sixth", 80)
	assertContains(t, out, "5.")
	assertContains(t, out, "6.")
}

func TestRenderTaskList(t *testing.T) {
	out := renderStripped("- [x] Done\n- [ ] Todo", 80)
	assertContains(t, out, "☑")
	assertContains(t, out, "☐")
	assertContains(t, out, "Done")
	assertContains(t, out, "Todo")
}

func TestRenderListWithBoldStart(t *testing.T) {
	out := renderStripped("- **Term** definition", 80)
	assertContains(t, out, "•")
	assertContains(t, out, "Term")
	assertContains(t, out, "definition")
}

func TestRenderListItemWrapping(t *testing.T) {
	src := "- This is a long list item that should wrap at the terminal width boundary when it exceeds the limit"
	out := renderStripped(src, 40)
	assertContains(t, out, "long list item")
}

func TestRenderOrderedListWrapping(t *testing.T) {
	src := "1. First item with a long description that should wrap at terminal width"
	out := renderStripped(src, 40)
	assertContains(t, out, "1.")
	assertContains(t, out, "long description")
}

// ── Table Rendering ─────────────────────────────────────────────────────────

func TestRenderTableBasic(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |"
	out := renderStripped(src, 80)
	assertContains(t, out, "A")
	assertContains(t, out, "B")
	assertContains(t, out, "1")
	assertContains(t, out, "2")
	assertContains(t, out, "┌")
	assertContains(t, out, "└")
	assertContains(t, out, "│")
}

func TestRenderSmallTableCJKWidth(t *testing.T) {
	src := "| 名称 | 状态 |\n|------|------|\n| 服务 | 正常 |"
	out := renderStripped(src, 40)
	assertContains(t, out, "名称")
	assertContains(t, out, "正常")
	assertTableLinesWithinWidth(t, out, 40)
	assertTableLinesEqualWidth(t, out)
}

func TestRenderTableWithFormatting(t *testing.T) {
	src := "| **bold** | *italic* |\n|----------|----------|\n| `code` | [link](url) |"
	out := renderStripped(src, 80)
	assertContains(t, out, "bold")
	assertContains(t, out, "italic")
	assertContains(t, out, "code")
	assertContains(t, out, "link")
}

func TestRenderTableMultipleRows(t *testing.T) {
	src := "| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |\n| Carol | 35 |"
	out := renderStripped(src, 80)
	assertContains(t, out, "Alice")
	assertContains(t, out, "Bob")
	assertContains(t, out, "Carol")
	assertContains(t, out, "30")
}

func TestRenderTableColumnAlignment(t *testing.T) {
	src := "| Short | Very Long Column |\n|-------|------------------|\n| a | b |"
	out := renderStripped(src, 80)
	assertContains(t, out, "Short")
	assertContains(t, out, "Very Long Column")
}

func TestRenderTableLongManyRows(t *testing.T) {
	// Test table with many rows (long table)
	src := "| ID | Name | Value |\n|----|------|-------|\n"
	for i := 0; i < 50; i++ {
		src += "| " + string(rune('0'+i%10)) + " | Item " + string(rune('A'+i%26)) + " | Value " + string(rune('0'+i%10)) + " |\n"
	}
	out := renderStripped(src, 80)
	// Should render all rows and not crash
	assertContains(t, out, "ID")
	assertContains(t, out, "Name")
	assertContains(t, out, "Value")
	assertContains(t, out, "Item")
}

func TestRenderTableWideManyColumnsLongContent(t *testing.T) {
	// Test wide table with many columns and long content in each cell
	src := `| Command | Description | Default | Required | Example | Notes |
|---------|-------------|---------|----------|---------|-------|
| --input | Path to input file that contains the data to be processed by the application | none | yes | /path/to/input.txt | Must be a valid readable text file with proper formatting |
| --output | Path to output file where the processed results will be written by the application | stdout | no | /path/to/output.json | Directory must exist and be writable by current user |
| --config | Path to configuration file in JSON format that contains application settings | ~/.config/app.json | no | /etc/app/config.json | Supports environment variable expansion |
| --verbose | Enable verbose debug output to see what's happening under the hood when running the application | false | no | true | When enabled logs are written to stderr |
| --timeout | Timeout in seconds after which the application will automatically terminate if not completed | 300 | no | 600 | Must be a positive integer value greater than zero |
`
	out := renderStripped(src, 100)
	assertContains(t, out, "Command")
	assertContains(t, out, "Description")
	assertContains(t, out, "--input")
	assertContains(t, out, "Path to input")
	assertContains(t, out, "file that")
	assertContains(t, out, "Timeout")
	assertTableLinesWithinWidth(t, out, 100)
	assertTableLinesEqualWidth(t, out)
}

func TestRenderTableWrapsLongUnbrokenCell(t *testing.T) {
	src := "| Key | Value |\n|-----|-------|\n| token | SupercalifragilisticexpialidociousSupercalifragilisticexpialidocious |"
	out := renderStripped(src, 40)
	assertContains(t, out, "token")
	assertContains(t, out, "Supercalif")
	assertTableLinesWithinWidth(t, out, 40)
	assertTableLinesEqualWidth(t, out)
}

// ── Thematic Break Rendering ────────────────────────────────────────────────

func TestRenderThematicBreak(t *testing.T) {
	out := renderStripped("---", 80)
	assertContains(t, out, "─")
}

func TestRenderThematicBreakWidth(t *testing.T) {
	out := render("---", 50)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "─") {
			w := visualWidth(line)
			if w != 48 { // width-2
				t.Errorf("thematic break width: expected 48, got %d", w)
			}
		}
	}
}

// ── Inline Rendering ────────────────────────────────────────────────────────

func TestRenderBold(t *testing.T) {
	out := render("Hello **world**", 80)
	if !strings.Contains(out, "\033[1m") {
		t.Error("expected bold ANSI code")
	}
}

func TestRenderItalic(t *testing.T) {
	out := render("Hello *world*", 80)
	if !strings.Contains(out, "\033[3m") {
		t.Error("expected italic ANSI code")
	}
}

func TestRenderItalicDoesNotLeakToAdjacentText(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "both sides",
			src:  "111*222*333",
			want: "111\033[3m222\033[0m333",
		},
		{
			name: "leading text",
			src:  "111*222*",
			want: "111\033[3m222\033[0m",
		},
		{
			name: "trailing text",
			src:  "*222*333",
			want: "\033[3m222\033[0m333",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render(tt.src, 80)
			if !strings.Contains(out, tt.want) {
				t.Fatalf("expected italic ANSI to be scoped as %q, got %q", tt.want, out)
			}
			if stripped := strings.TrimSpace(StripANSI(out)); stripped != strings.ReplaceAll(tt.src, "*", "") {
				t.Fatalf("unexpected stripped output: %q", stripped)
			}
		})
	}
}

func TestRenderCodeSpan(t *testing.T) {
	out := renderStripped("Use `fmt.Println` to print", 80)
	assertContains(t, out, "fmt.Println")
}

func TestRenderCodeSpanNoBackgroundOrPadding(t *testing.T) {
	out := render("Use `fmt.Println` to print", 80)
	assertNoBackgroundANSI(t, out)
	stripped := strings.TrimSpace(StripANSI(out))
	if stripped != "Use fmt.Println to print" {
		t.Fatalf("expected code span without extra padding, got %q", stripped)
	}
}

func TestRenderCodeBlockNoBackground(t *testing.T) {
	out := render("```go\nfmt.Println(\"hello\")\n```", 80)
	assertNoBackgroundANSI(t, out)
}

func TestRenderLink(t *testing.T) {
	out := renderStripped("[Go](https://go.dev)", 80)
	assertContains(t, out, "Go")
	assertContains(t, out, "https://go.dev")
}

func TestRenderAutolink(t *testing.T) {
	out := renderStripped("<https://example.com>", 80)
	assertContains(t, out, "https://example.com")
}

func TestRenderStrikethrough(t *testing.T) {
	out := render("~~deleted~~", 80)
	if !strings.Contains(out, "\033[9m") {
		t.Error("expected strikethrough ANSI code")
	}
}

func TestRenderImage(t *testing.T) {
	out := renderStripped("![alt](https://example.com/img.png)", 80)
	assertContains(t, out, "🖼")
	assertContains(t, out, "alt")
	assertContains(t, out, "https://example.com/img.png")
}

func TestRenderNestedFormatting(t *testing.T) {
	out := render("**bold *and italic* text**", 80)
	// Should have both bold and italic codes
	if !strings.Contains(out, "\033[1m") {
		t.Error("expected bold")
	}
	if !strings.Contains(out, "\033[3m") {
		t.Error("expected italic inside bold")
	}
}

// ── Wrapping Tests ──────────────────────────────────────────────────────────

func TestWrapANSIBasic(t *testing.T) {
	input := "Hello world this is a test"
	result := wrapANSI(input, 15, "", 0)
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Errorf("expected wrapping, got %d line(s)", len(lines))
	}
}

func TestWrapANSIWordBoundary(t *testing.T) {
	input := "Hello world test"
	result := wrapANSI(input, 10, "", 0)
	lines := strings.Split(result, "\n")
	// Should break at word boundaries
	for _, line := range lines {
		w := visualWidth(line)
		if w > 10 {
			t.Errorf("line exceeds width: %d chars: %q", w, line)
		}
	}
}

func TestWrapANSIPreservesStyles(t *testing.T) {
	input := "\033[1mHello world this is bold text\033[0m"
	result := wrapANSI(input, 15, "", 0)
	// Bold code should appear on continuation lines
	if !strings.Contains(result, "\033[1m") {
		t.Error("expected bold code preserved")
	}
}

func TestWrapANSIWithIndent(t *testing.T) {
	input := "Hello world this is a test"
	result := wrapANSI(input, 20, "  ", 0)
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("expected wrapping")
	}
	// Continuation lines should have indent
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) > 0 && !strings.HasPrefix(lines[i], "  ") {
			t.Errorf("continuation line %d missing indent: %q", i, lines[i])
		}
	}
}

func TestWrapANSIWithLineSpacing(t *testing.T) {
	input := "Hello world this is a test that wraps"
	result := wrapANSI(input, 15, "", 1)
	// Should have extra blank lines between wrapped segments
	if !strings.Contains(result, "\n\n") {
		t.Error("expected blank lines for line spacing")
	}
}

func TestWrapANSISingleWord(t *testing.T) {
	input := "Supercalifragilisticexpialidocious"
	result := wrapANSI(input, 10, "", 0)
	// Single long word should be on its own line (no good break point)
	assertContains(t, result, "Supercalifragilisticexpialidocious")
}

func TestWrapANSIEmpty(t *testing.T) {
	result := wrapANSI("", 10, "", 0)
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestWrapANSIZeroWidth(t *testing.T) {
	input := "Hello"
	result := wrapANSI(input, 0, "", 0)
	if result != input {
		t.Error("zero width should return input unchanged")
	}
}

func TestRenderThematicBreakNarrowWidth(t *testing.T) {
	out := renderStripped("---", 1)
	assertContains(t, out, "─")
}

// ── Theme Tests ─────────────────────────────────────────────────────────────

func TestDefaultThemeNotNil(t *testing.T) {
	theme := DefaultTheme()
	if theme == nil {
		t.Fatal("default theme should not be nil")
	}
}

func TestLightThemeNotNil(t *testing.T) {
	theme := LightTheme()
	if theme == nil {
		t.Fatal("light theme should not be nil")
	}
}

func TestAutoThemeNotNil(t *testing.T) {
	theme := AutoTheme()
	if theme == nil {
		t.Fatal("auto theme should not be nil")
	}
}

func TestThemeHeadingStyles(t *testing.T) {
	theme := DefaultTheme()
	if theme.Heading1 == "" {
		t.Error("Heading1 style should not be empty")
	}
	if theme.Heading6 == "" {
		t.Error("Heading6 style should not be empty")
	}
}

func TestThemeLineSpacing(t *testing.T) {
	theme := DefaultTheme()
	if theme.LineSpacing < 0 {
		t.Error("LineSpacing should be non-negative")
	}
}

// ── StripANSI Tests ─────────────────────────────────────────────────────────

func TestStripANSI(t *testing.T) {
	input := "\033[1mBold\033[0m text"
	result := StripANSI(input)
	if result != "Bold text" {
		t.Errorf("expected 'Bold text', got %q", result)
	}
}

func TestStripANSIMultiple(t *testing.T) {
	input := "\033[1m\033[3mBold italic\033[0m"
	result := StripANSI(input)
	if result != "Bold italic" {
		t.Errorf("expected 'Bold italic', got %q", result)
	}
}

func TestStripANSINone(t *testing.T) {
	input := "Plain text"
	result := StripANSI(input)
	if result != input {
		t.Errorf("no ANSI to strip: %q", result)
	}
}

func TestStripANSIEmpty(t *testing.T) {
	result := StripANSI("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestStripANSI256Color(t *testing.T) {
	input := "\033[38;5;196mRed\033[0m"
	result := StripANSI(input)
	if result != "Red" {
		t.Errorf("expected 'Red', got %q", result)
	}
}

// ── VisualWidth Tests ───────────────────────────────────────────────────────

func TestVisualWidthPlain(t *testing.T) {
	if visualWidth("hello") != 5 {
		t.Errorf("expected 5, got %d", visualWidth("hello"))
	}
}

func TestVisualWidthANSI(t *testing.T) {
	w := visualWidth("\033[1mhello\033[0m")
	if w != 5 {
		t.Errorf("expected 5 (ignoring ANSI), got %d", w)
	}
}

func TestVisualWidthEmpty(t *testing.T) {
	if visualWidth("") != 0 {
		t.Error("expected 0 for empty string")
	}
}

func TestVisualWidthUnicode(t *testing.T) {
	w := visualWidth("你好")
	if w != 4 {
		t.Errorf("expected 4, got %d", w)
	}
}

func TestVisualWidthCombiningAndEmoji(t *testing.T) {
	if w := visualWidth("e\u0301"); w != 1 {
		t.Errorf("expected combining mark width 1, got %d", w)
	}
	if w := visualWidth("🙂"); w != 2 {
		t.Errorf("expected emoji width 2, got %d", w)
	}
}

// ── Full Document Rendering Tests ───────────────────────────────────────────

func TestRenderFullDocument(t *testing.T) {
	src := `# Title

A paragraph with **bold** and *italic*.

> A blockquote

- List item

` + "```go\nfmt.Println(\"hello\")\n```" + `

| A | B |
|---|---|
| 1 | 2 |

---

[Link](https://go.dev)`
	out := renderStripped(src, 80)
	// All major elements should be present
	assertContains(t, out, "Title")
	assertContains(t, out, "bold")
	assertContains(t, out, "blockquote")
	assertContains(t, out, "List item")
	assertContains(t, out, "fmt.Println")
	assertContains(t, out, "A")
	assertContains(t, out, "─")
	assertContains(t, out, "Link")
}

func TestRenderFullDocumentWide(t *testing.T) {
	src := `# Title

` + "```python\ndef hello():\n    print('world')\n```" + `

| Name | Value |
|------|-------|
| foo  | bar   |`
	out := renderStripped(src, 120)
	assertContains(t, out, "Title")
	assertContains(t, out, "def hello()")
	assertContains(t, out, "foo")
}

func TestRenderFullDocumentNarrow(t *testing.T) {
	src := `# Title

A paragraph with **bold** text.

- List item`
	out := renderStripped(src, 30)
	assertContains(t, out, "Title")
	assertContains(t, out, "bold")
	assertContains(t, out, "•")
}

func TestRenderMultipleCodeBlocksInDocument(t *testing.T) {
	src := "```go\ncode1\n```\n\nSome text\n\n```python\ncode2\n```\n\nMore text\n\n```bash\ncode3\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "code1")
	assertContains(t, out, "code2")
	assertContains(t, out, "code3")
	assertContains(t, out, "Some text")
	assertContains(t, out, "More text")
}

func TestRenderCodeBlockBetweenParagraphs(t *testing.T) {
	src := "Before paragraph.\n\n```\ncode block\n```\n\nAfter paragraph."
	out := renderStripped(src, 80)
	assertContains(t, out, "Before paragraph.")
	assertContains(t, out, "code block")
	assertContains(t, out, "After paragraph.")
}

func TestRenderTableBetweenCodeBlocks(t *testing.T) {
	src := "```go\nfunc main() {}\n```\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n```python\ndef hello(): pass\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "func main()")
	assertContains(t, out, "A")
	assertContains(t, out, "def hello()")
}

func TestRenderEmptyDocument(t *testing.T) {
	out := render("", 80)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestRenderBlankLinesOnly(t *testing.T) {
	out := render("\n\n\n", 80)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output for blank lines, got %q", out)
	}
}

// ── Parse → Render Integration ──────────────────────────────────────────────

func TestParseRenderCodeBlockWithLanguage(t *testing.T) {
	src := "```javascript\nconst x = 42;\nconsole.log(x);\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "javascript")
	assertContains(t, out, "const x = 42;")
	assertContains(t, out, "console.log(x);")
}

func TestParseRenderNestedBlockquote(t *testing.T) {
	src := "> Outer\n> > Inner\n> > > Deep"
	out := renderStripped(src, 80)
	assertContains(t, out, "Outer")
	assertContains(t, out, "Inner")
	assertContains(t, out, "Deep")
}

func TestParseRenderListWithCodeSpan(t *testing.T) {
	src := "- Use `fmt.Println` for output\n- Use `log.Printf` for logging"
	out := renderStripped(src, 80)
	assertContains(t, out, "fmt.Println")
	assertContains(t, out, "log.Printf")
}

func TestParseRenderTableWithLinks(t *testing.T) {
	src := "| Name | URL |\n|------|-----|\n| Go | [go.dev](https://go.dev) |"
	out := renderStripped(src, 80)
	assertContains(t, out, "Go")
	assertContains(t, out, "go.dev")
}

func TestParseRenderHeadingWithCodeSpan(t *testing.T) {
	src := "# Using `fmt` package"
	out := renderStripped(src, 80)
	assertContains(t, out, "Using")
	assertContains(t, out, "fmt")
}

func TestParseRenderBlockquoteWithCodeBlock(t *testing.T) {
	src := "> Example:\n> ```\n> code here\n> ```"
	out := renderStripped(src, 80)
	assertContains(t, out, "Example:")
	assertContains(t, out, "code here")
}

func TestParseRenderTaskListWithFormatting(t *testing.T) {
	src := "- [x] **Done** task\n- [ ] *Pending* task"
	out := renderStripped(src, 80)
	assertContains(t, out, "☑")
	assertContains(t, out, "☐")
	assertContains(t, out, "Done")
	assertContains(t, out, "Pending")
}

func TestParseRenderLaTeXBlockMath(t *testing.T) {
	src := "$$E = mc^2$$"
	out := renderStripped(src, 80)
	assertContains(t, out, "E = mc^2")
	assertContains(t, out, "blockmath")
}

func TestParseRenderLaTeXInline(t *testing.T) {
	src := "Equation: $a^2 + b^2$"
	out := renderStripped(src, 80)
	assertContains(t, out, "Equation:")
	assertContains(t, out, "a^2 + b^2")
}

func TestParseRenderStreamingMode(t *testing.T) {
	src := "Hello **world"
	doc := parser.Parse(src, parser.StreamOption())
	r := New(nil, 80)
	out := StripANSI(r.Render(doc))
	// Speculative rewrite should close the bold
	assertContains(t, out, "Hello")
	assertContains(t, out, "world")
}

func TestParseRenderComplexDocument(t *testing.T) {
	src := `# API Reference

## Installation

` + "```bash\ngo get github.com/example/pkg\n```" + `

## Usage

Import the package:

` + "```go\nimport \"github.com/example/pkg\"\n\nfunc main() {\n    pkg.DoSomething()\n}\n```" + `

### Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --verbose | bool | false | Enable verbose output |
| --output | string | stdout | Output file path |

### Examples

> **Note:** Make sure to check the [documentation](https://docs.example.com) for more details.

1. First step
2. Second step with ` + "`code`" + ` inside
3. Third step

---

*Last updated: 2024*`
	out := renderStripped(src, 80)
	assertContains(t, out, "API Reference")
	assertContains(t, out, "Installation")
	assertContains(t, out, "go get")
	assertContains(t, out, "import")
	assertContains(t, out, "DoSomething()")
	assertContains(t, out, "Options")
	assertContains(t, out, "verbose")
	assertContains(t, out, "Note:")
	assertContains(t, out, "documentation")
	assertContains(t, out, "1.")
	assertContains(t, out, "─")
	assertContains(t, out, "Last updated")
}

// ── Word Wrapping in Lists ──────────────────────────────────────────────────

func TestRenderListItemWrappingContinuationIndent(t *testing.T) {
	src := "- This is a very long list item that wraps"
	out := renderStripped(src, 30)
	lines := strings.Split(out, "\n")
	// First line should start with bullet
	if len(lines) > 0 && !strings.Contains(lines[0], "•") {
		t.Error("first line should have bullet")
	}
	// Continuation lines should be indented past the bullet
	if len(lines) > 1 {
		secondLine := lines[1]
		if len(secondLine) > 0 && secondLine[0] != ' ' {
			t.Error("continuation line should be indented")
		}
	}
}

func TestRenderOrderedListWrappingContinuationIndent(t *testing.T) {
	src := "1. First item with long text"
	out := renderStripped(src, 25)
	assertContains(t, out, "1.")
	assertContains(t, out, "First item")
}

// ── Edge Cases ──────────────────────────────────────────────────────────────

func TestRenderSingleCharacter(t *testing.T) {
	out := renderStripped("x", 80)
	assertContains(t, out, "x")
}

func TestRenderOnlyFormatting(t *testing.T) {
	out := renderStripped("**bold**", 80)
	assertContains(t, out, "bold")
}

func TestRenderEmptyCodeBlock(t *testing.T) {
	out := renderStripped("```\n```", 80)
	assertContains(t, out, "┌")
}

func TestRenderEmptyTable(t *testing.T) {
	out := renderStripped("| |\n|--|\n| |", 80)
	// Should not crash
	assertContains(t, out, "│")
}

func TestRenderUnicode(t *testing.T) {
	out := renderStripped("# 你好世界\n\n这是**粗体**文本", 80)
	assertContains(t, out, "你好世界")
	assertContains(t, out, "粗体")
}

func TestRenderEmoji(t *testing.T) {
	out := renderStripped("# 🚀 Rocket\n\nHello 🌍 World", 80)
	assertContains(t, out, "🚀")
	assertContains(t, out, "🌍")
}

func TestRenderVeryNarrow(t *testing.T) {
	// Very narrow terminal shouldn't crash
	out := renderStripped("# Hello\n\nSome text", 10)
	assertContains(t, out, "Hello")
}

func TestRenderVeryWide(t *testing.T) {
	// Very wide terminal shouldn't crash
	out := renderStripped("# Hello", 500)
	assertContains(t, out, "Hello")
}

func TestRenderParagraphWrappingWithInlineCode(t *testing.T) {
	src := "Use `fmt.Println` for output and `log.Printf` for logging in your Go applications"
	out := renderStripped(src, 40)
	assertContains(t, out, "fmt.Println")
	assertContains(t, out, "log.Printf")
}

func TestRenderParagraphWrappingWithBoldItalic(t *testing.T) {
	src := "This has **bold text** and *italic text* and ***both*** in a paragraph that wraps"
	out := renderStripped(src, 40)
	assertContains(t, out, "bold text")
	assertContains(t, out, "italic text")
}

func TestRenderCodeBlockFollowedByList(t *testing.T) {
	src := "```\ncode\n```\n\n- Item 1\n- Item 2"
	out := renderStripped(src, 80)
	assertContains(t, out, "code")
	assertContains(t, out, "Item 1")
}

func TestRenderListFollowedByCodeBlock(t *testing.T) {
	src := "- Item 1\n- Item 2\n\n```\ncode\n```"
	out := renderStripped(src, 80)
	assertContains(t, out, "Item 1")
	assertContains(t, out, "code")
}

func TestRenderTableFollowedByBlockquote(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |\n\n> Quote"
	out := renderStripped(src, 80)
	assertContains(t, out, "A")
	assertContains(t, out, "Quote")
}

func TestRenderBlockquoteFollowedByTable(t *testing.T) {
	src := "> Quote\n\n| A | B |\n|---|---|\n| 1 | 2 |"
	out := renderStripped(src, 80)
	assertContains(t, out, "Quote")
	assertContains(t, out, "A")
}
