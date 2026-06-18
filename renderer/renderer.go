// Package renderer converts a Markdown AST into ANSI-styled terminal output.
// Zero external dependencies — only stdlib.
//
// Mirrors the Swift rendering pipeline:
//   - ParagraphView lineSpacing → wrapANSI with lineSpacing param
//   - TextFonts kern → Theme.LetterSpacing (character spacing)
//   - InlineConvertible attribute merging → styleStack (cascading ANSI codes)
//   - CodeBlockView → boxed code blocks with language label
//   - TableView → box-drawing table with column alignment
//   - BlockQuoteView → nested │ prefix with depth-aware styling
package renderer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/startvibecoding/GoStreamingMarkdown/parser"
)

// ── ANSI Escape Codes ───────────────────────────────────────────────────────

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiItalic    = "\033[3m"
	ansiUnderline = "\033[4m"
	ansiStrike    = "\033[9m"

	ansiFgBlack   = "\033[30m"
	ansiFgRed     = "\033[31m"
	ansiFgGreen   = "\033[32m"
	ansiFgYellow  = "\033[33m"
	ansiFgBlue    = "\033[34m"
	ansiFgMagenta = "\033[35m"
	ansiFgCyan    = "\033[36m"
	ansiFgWhite   = "\033[37m"

	ansiFgBrightBlack   = "\033[90m"
	ansiFgBrightRed     = "\033[91m"
	ansiFgBrightGreen   = "\033[92m"
	ansiFgBrightYellow  = "\033[93m"
	ansiFgBrightBlue    = "\033[94m"
	ansiFgBrightMagenta = "\033[95m"
	ansiFgBrightCyan    = "\033[96m"
	ansiFgBrightWhite   = "\033[97m"

	ansiBgBlack   = "\033[40m"
	ansiBgRed     = "\033[41m"
	ansiBgGreen   = "\033[42m"
	ansiBgYellow  = "\033[43m"
	ansiBgBlue    = "\033[44m"
	ansiBgMagenta = "\033[45m"
	ansiBgCyan    = "\033[46m"
	ansiBgWhite   = "\033[47m"

	ansiBgBrightBlack = "\033[100m"
	ansiBgBrightWhite = "\033[107m"
)

// ── Theme ───────────────────────────────────────────────────────────────────

// Theme defines ANSI styles for every Markdown element.
// Mirrors Swift MarkdownRenderConfig with Typography + Colors.
type Theme struct {
	Heading         string
	Heading1        string
	Heading2        string
	Heading3        string
	Heading4        string
	Heading5        string
	Heading6        string
	BlockQuote      string
	BlockQuoteBar   string
	CodeText        string
	CodeBg          string
	CodeLang        string
	TableBorder     string
	TableHeader     string
	TableHeaderText string
	TableCell       string
	Horizontal      string
	ListBullet      string
	ListNumber      string
	Bold            string
	Italic          string
	Code            string
	CodeBgInline    string
	Link            string
	LinkURL         string
	Strike          string
	TaskChecked     string
	TaskUnchecked   string
	LetterSpacing   int // extra spaces between letters (mirrors kern)
	LineSpacing     int // blank lines between wrapped lines (mirrors lineSpacing: 5)
}

// DefaultTheme returns a dark-terminal-friendly theme.
func DefaultTheme() *Theme {
	return &Theme{
		Heading:         ansiBold + ansiFgBrightCyan,
		Heading1:        ansiBold + ansiFgBrightCyan,
		Heading2:        ansiBold + ansiFgBrightBlue,
		Heading3:        ansiBold + ansiFgBrightMagenta,
		Heading4:        ansiBold + ansiFgBrightGreen,
		Heading5:        ansiBold + ansiFgYellow,
		Heading6:        ansiBold + ansiFgBrightRed,
		BlockQuote:      ansiItalic + ansiFgBrightBlack,
		BlockQuoteBar:   ansiFgBrightBlack,
		CodeText:        ansiFgBrightWhite,
		CodeBg:          "",
		CodeLang:        ansiFgBrightBlack + ansiDim,
		TableBorder:     ansiFgBrightBlack,
		TableHeader:     ansiBold + ansiFgBrightWhite,
		TableHeaderText: ansiBold + ansiFgBrightWhite,
		TableCell:       "",
		Horizontal:      ansiFgBrightBlack,
		ListBullet:      ansiFgBrightYellow,
		ListNumber:      ansiFgBrightYellow,
		Bold:            ansiBold,
		Italic:          ansiItalic,
		Code:            ansiFgBrightRed,
		CodeBgInline:    "",
		Link:            ansiFgBrightBlue + ansiUnderline,
		LinkURL:         ansiFgBrightBlack + ansiDim,
		Strike:          ansiStrike,
		TaskChecked:     ansiFgGreen,
		TaskUnchecked:   ansiFgBrightBlack,
		LetterSpacing:   0,
		LineSpacing:     1,
	}
}

// LightTheme returns a light-terminal-friendly theme.
func LightTheme() *Theme {
	return &Theme{
		Heading:         ansiBold + ansiFgCyan,
		Heading1:        ansiBold + ansiFgCyan,
		Heading2:        ansiBold + ansiFgBlue,
		Heading3:        ansiBold + ansiFgMagenta,
		Heading4:        ansiBold + ansiFgGreen,
		Heading5:        ansiBold + ansiFgYellow,
		Heading6:        ansiBold + ansiFgRed,
		BlockQuote:      ansiItalic + ansiFgBlack,
		BlockQuoteBar:   ansiFgBlack,
		CodeText:        ansiFgBlack,
		CodeBg:          "",
		CodeLang:        ansiFgBlack + ansiDim,
		TableBorder:     ansiFgBlack,
		TableHeader:     ansiBold + ansiFgBlack,
		TableHeaderText: ansiBold + ansiFgBlack,
		TableCell:       "",
		Horizontal:      ansiFgBlack,
		ListBullet:      ansiFgYellow,
		ListNumber:      ansiFgYellow,
		Bold:            ansiBold,
		Italic:          ansiItalic,
		Code:            ansiFgRed,
		CodeBgInline:    "",
		Link:            ansiFgBlue + ansiUnderline,
		LinkURL:         ansiFgBlack + ansiDim,
		Strike:          ansiStrike,
		TaskChecked:     ansiFgGreen,
		TaskUnchecked:   ansiFgBlack,
		LetterSpacing:   0,
		LineSpacing:     1,
	}
}

// AutoTheme returns the default theme. Actual auto-detection is done in main.go
// via COLORFGBG and TERM_PROGRAM environment variables.
func AutoTheme() *Theme {
	return DefaultTheme()
}

// ── Renderer ────────────────────────────────────────────────────────────────

// Renderer renders a Markdown AST to ANSI terminal output.
type Renderer struct {
	theme      *Theme
	width      int
	buf        bytes.Buffer
	styleStack []string // tracks active ANSI codes for attribute cascading
}

// New creates a Renderer with the given theme and terminal width.
func New(theme *Theme, width int) *Renderer {
	if theme == nil {
		theme = DefaultTheme()
	}
	if width <= 0 {
		width = 80
	}
	return &Renderer{theme: theme, width: width}
}

// Render renders the entire document AST to an ANSI string.
func (r *Renderer) Render(doc *parser.Node) string {
	r.buf.Reset()
	r.renderChildren(doc)
	return r.buf.String()
}

func (r *Renderer) renderChildren(n *parser.Node) {
	for _, child := range n.Children {
		r.renderNode(child)
	}
}

// pushStyle / popStyle implement attribute cascading (mirrors Swift NSAttributeContainer merging).
func (r *Renderer) pushStyle(style string) {
	r.styleStack = append(r.styleStack, style)
	r.buf.WriteString(style)
}

func (r *Renderer) popStyle() {
	if len(r.styleStack) > 0 {
		r.styleStack = r.styleStack[:len(r.styleStack)-1]
	}
	r.buf.WriteString(ansiReset)
	// Re-emit parent styles
	for _, s := range r.styleStack {
		r.buf.WriteString(s)
	}
}

func (r *Renderer) renderNode(n *parser.Node) {
	switch n.Type {
	case parser.NodeHeading:
		r.renderHeading(n)
	case parser.NodeParagraph:
		r.renderParagraph(n)
	case parser.NodeFencedCodeBlock:
		r.renderFencedCodeBlock(n)
	case parser.NodeIndentedCodeBlock:
		r.renderIndentedCodeBlock(n)
	case parser.NodeBlockquote:
		r.renderBlockquote(n)
	case parser.NodeThematicBreak:
		r.renderThematicBreak()
	case parser.NodeUnorderedList:
		r.renderList(n)
	case parser.NodeListItem:
		r.renderListItem(n)
	case parser.NodeTable:
		r.renderTable(n)
	case parser.NodeTableRow:
		// handled by renderTable
	case parser.NodeTableCell:
		// handled by renderTable

	// Inline — with style cascading
	case parser.NodeText:
		r.buf.WriteString(n.Text)
	case parser.NodeEmphasis:
		r.pushStyle(r.theme.Italic)
		r.renderInlineContent(n)
		r.popStyle()
	case parser.NodeStrong:
		r.pushStyle(r.theme.Bold)
		r.renderInlineContent(n)
		r.popStyle()
	case parser.NodeCodeSpan:
		r.pushStyle(r.theme.Code)
		r.buf.WriteString(n.Text)
		r.popStyle()
	case parser.NodeLink:
		r.pushStyle(r.theme.Link)
		r.renderInlineContent(n)
		r.popStyle()
		r.buf.WriteString(" ")
		r.buf.WriteString(r.theme.LinkURL)
		r.buf.WriteString("(" + n.URL + ")")
		r.buf.WriteString(ansiReset)
	case parser.NodeImage:
		alt := ""
		if len(n.Children) > 0 {
			alt = n.Children[0].TextContent()
		}
		r.buf.WriteString(r.theme.Code)
		r.buf.WriteString("🖼 " + alt)
		r.buf.WriteString(ansiReset)
		if n.URL != "" {
			r.buf.WriteString(" ")
			r.buf.WriteString(r.theme.LinkURL)
			r.buf.WriteString("(" + n.URL + ")")
			r.buf.WriteString(ansiReset)
		}
	case parser.NodeStrikethrough:
		r.pushStyle(r.theme.Strike)
		r.renderInlineContent(n)
		r.popStyle()
	case parser.NodeAutolink:
		r.buf.WriteString(r.theme.Link)
		r.buf.WriteString(n.URL)
		r.buf.WriteString(ansiReset)
	case parser.NodeSoftBreak:
		r.buf.WriteByte('\n')
	case parser.NodeHardBreak, parser.NodeLineBreak:
		r.buf.WriteString("\n")
	default:
		r.renderChildren(n)
	}
}

// ── Block Renderers ─────────────────────────────────────────────────────────

func (r *Renderer) getHeadingStyle(level int) string {
	switch level {
	case 1:
		return r.theme.Heading1
	case 2:
		return r.theme.Heading2
	case 3:
		return r.theme.Heading3
	case 4:
		return r.theme.Heading4
	case 5:
		return r.theme.Heading5
	case 6:
		return r.theme.Heading6
	default:
		return r.theme.Heading
	}
}

func (r *Renderer) renderHeading(n *parser.Node) {
	prefix := strings.Repeat("#", n.Level) + " "
	style := r.getHeadingStyle(n.Level)
	var tmp bytes.Buffer
	oldBuf := r.buf
	r.buf = tmp
	r.buf.WriteString(style)
	r.buf.WriteString(prefix)
	r.renderInlineContent(n)
	r.buf.WriteString(ansiReset)
	tmp = r.buf
	r.buf = oldBuf
	wrapped := wrapANSI(tmp.String(), r.width, "", r.theme.LineSpacing)
	r.buf.WriteString(wrapped)
	r.buf.WriteString("\n\n")
}

func (r *Renderer) renderParagraph(n *parser.Node) {
	// Mirrors Swift ParagraphView with lineBreakMode = .byWordWrapping, lineSpacing: 5
	var tmp bytes.Buffer
	oldBuf := r.buf
	r.buf = tmp
	r.renderInlineContent(n)
	tmp = r.buf
	r.buf = oldBuf
	wrapped := wrapANSI(tmp.String(), r.width, "", r.theme.LineSpacing)
	r.buf.WriteString(wrapped)
	r.buf.WriteString("\n\n")
}

func (r *Renderer) renderFencedCodeBlock(n *parser.Node) {
	lang := n.Language
	lines := strings.Split(n.Code, "\n")
	r.renderCodeBox(lang, lines)
}

func (r *Renderer) renderIndentedCodeBlock(n *parser.Node) {
	lines := strings.Split(n.Code, "\n")
	r.renderCodeBox("", lines)
}

func (r *Renderer) renderCodeBox(lang string, lines []string) {
	boxW := r.width - 2
	if boxW < 20 {
		boxW = 20
	}

	// Top border
	r.buf.WriteString(r.theme.TableBorder)
	r.buf.WriteString("┌" + strings.Repeat("─", boxW) + "┐")
	r.buf.WriteString(ansiReset + "\n")

	drawCodeLine := func(line string, style string) {
		r.buf.WriteString(r.theme.TableBorder + "│")
		r.buf.WriteString(" " + style + line)
		padding := boxW - 1 - visualWidth(line)
		if padding > 0 {
			r.buf.WriteString(strings.Repeat(" ", padding))
		}
		r.buf.WriteString(ansiReset + r.theme.TableBorder + "│" + ansiReset + "\n")
	}

	// Language label
	if lang != "" {
		label := lang + " "
		for _, line := range hardWrapANSI(label, boxW-1) {
			drawCodeLine(line, r.theme.CodeLang)
		}

		r.buf.WriteString(r.theme.TableBorder)
		r.buf.WriteString("├" + strings.Repeat("─", boxW) + "┤")
		r.buf.WriteString(ansiReset + "\n")
	}

	// Code lines
	for _, line := range lines {
		for _, wrapped := range hardWrapANSI(line, boxW-1) {
			drawCodeLine(wrapped, r.theme.CodeText)
		}
	}

	// Bottom border
	r.buf.WriteString(r.theme.TableBorder)
	r.buf.WriteString("└" + strings.Repeat("─", boxW) + "┘")
	r.buf.WriteString(ansiReset + "\n\n")
}

func (r *Renderer) renderBlockquote(n *parser.Node) {
	// Recursive rendering with depth-aware bar (mirrors Swift BlockQuoteView)
	depth := n.QuoteLevel
	bar := strings.Repeat("│ ", depth+1)
	barW := visualWidth(bar)

	sub := &Renderer{theme: r.theme, width: r.width - barW}
	sub.renderChildren(n)
	content := strings.TrimRight(sub.buf.String(), "\n")

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		r.buf.WriteString(r.theme.BlockQuoteBar + bar)
		r.buf.WriteString(r.theme.BlockQuote + line + ansiReset + "\n")
	}
	r.buf.WriteString("\n")
}

func (r *Renderer) renderThematicBreak() {
	w := r.width - 2
	if w < 1 {
		w = 1
	}
	r.buf.WriteString(r.theme.Horizontal + strings.Repeat("─", w) + ansiReset + "\n\n")
}

func (r *Renderer) renderList(n *parser.Node) {
	for _, child := range n.Children {
		r.renderListItem(child)
	}
	r.buf.WriteString("\n")
}

func (r *Renderer) renderListItem(n *parser.Node) {
	indent := ""
	depth := 0
	p := n.Parent
	for p != nil {
		if p.Type == parser.NodeUnorderedList || p.Type == parser.NodeOrderedList {
			depth++
		}
		p = p.Parent
	}
	indent = strings.Repeat("  ", depth-1)

	isOrdered := n.Parent != nil && n.Parent.Type == parser.NodeOrderedList
	marker := ""
	if isOrdered {
		startNum := 1
		if n.Parent != nil {
			startNum = n.Parent.StartNum
		}
		idx := n.IndexInParent()
		if idx < 0 {
			idx = 0
		}
		num := startNum + idx
		marker = fmt.Sprintf("%d. ", num)
	} else {
		marker = "• "
	}

	// Task checkbox (uses IsTaskItem field from parser, since [ ] prefix is stripped)
	if n.IsTaskItem {
		if n.Checked {
			r.buf.WriteString(indent + r.theme.TaskChecked + "☑ " + ansiReset)
		} else {
			r.buf.WriteString(indent + r.theme.TaskUnchecked + "☐ " + ansiReset)
		}
	} else {
		if isOrdered {
			r.buf.WriteString(indent + r.theme.ListNumber + marker + ansiReset)
		} else {
			r.buf.WriteString(indent + r.theme.ListBullet + marker + ansiReset)
		}
	}

	// Render paragraph content with wrapping
	markerW := visualWidth(indent) + visualWidth(marker)
	contIndent := indent + strings.Repeat(" ", visualWidth(marker))

	for _, child := range n.Children {
		if child.Type == parser.NodeParagraph {
			var tmp bytes.Buffer
			oldBuf := r.buf
			r.buf = tmp
			r.renderInlineContent(child)
			tmp = r.buf
			r.buf = oldBuf

			wrapped := wrapANSI(tmp.String(), r.width-markerW, "", r.theme.LineSpacing)
			lines := strings.Split(wrapped, "\n")
			for i, line := range lines {
				if i == 0 {
					r.buf.WriteString(line)
				} else {
					r.buf.WriteString("\n" + contIndent + line)
				}
			}
		} else {
			r.renderNode(child)
		}
	}
	r.buf.WriteString("\n")
}

// ── Table Renderer (with inline formatting in cells) ────────────────────────

func (r *Renderer) renderTable(n *parser.Node) {
	if len(n.Children) == 0 {
		return
	}

	type row struct {
		cells    []string
		isHeader bool
	}
	var rows []row
	for _, child := range n.Children {
		if child.Type == parser.NodeTableRow {
			var cells []string
			for _, cell := range child.Children {
				if cell.Type == parser.NodeTableCell {
					// Render cell WITH inline formatting (not just plain text)
					sub := &Renderer{theme: r.theme, width: r.width}
					sub.renderInlineContent(cell)
					cells = append(cells, strings.TrimSpace(strings.TrimRight(sub.buf.String(), "\n")))
				}
			}
			rows = append(rows, row{cells: cells, isHeader: child.IsTableHeader})
		}
	}

	if len(rows) == 0 {
		return
	}

	colCount := 0
	for _, rw := range rows {
		if len(rw.cells) > colCount {
			colCount = len(rw.cells)
		}
	}
	colWidths := make([]int, colCount)
	for _, rw := range rows {
		for i, cell := range rw.cells {
			if i < colCount {
				w := visualWidth(cell)
				if w > colWidths[i] {
					colWidths[i] = w
				}
			}
		}
	}
	for i := range colWidths {
		if colWidths[i] < 3 {
			colWidths[i] = 3
		}
	}
	colWidths = fitTableColumnWidths(colWidths, r.width)

	drawSep := func(left, mid, right, fill string) {
		r.buf.WriteString(r.theme.TableBorder + left)
		for i, w := range colWidths {
			r.buf.WriteString(strings.Repeat(fill, w+2))
			if i < colCount-1 {
				r.buf.WriteString(mid)
			}
		}
		r.buf.WriteString(right + ansiReset + "\n")
	}

	drawRow := func(cells []string, isHeader bool) {
		wrappedCells := make([][]string, colCount)
		rowHeight := 1
		for i := 0; i < colCount; i++ {
			text := ""
			if i < len(cells) {
				text = cells[i]
			}
			wrapped := wrapTableCell(text, colWidths[i])
			wrappedCells[i] = wrapped
			if len(wrapped) > rowHeight {
				rowHeight = len(wrapped)
			}
		}

		for lineIdx := 0; lineIdx < rowHeight; lineIdx++ {
			r.buf.WriteString(r.theme.TableBorder + "│")
			for i := 0; i < colCount; i++ {
				text := ""
				if lineIdx < len(wrappedCells[i]) {
					text = wrappedCells[i][lineIdx]
				}
				pad := colWidths[i] - visualWidth(text)
				if pad < 0 {
					pad = 0
				}
				if isHeader {
					r.buf.WriteString(r.theme.TableHeaderText + " " + text + strings.Repeat(" ", pad+1) + ansiReset)
				} else {
					r.buf.WriteString(" " + text + strings.Repeat(" ", pad+1) + ansiReset)
				}
				r.buf.WriteString(r.theme.TableBorder + "│")
			}
			r.buf.WriteString(ansiReset + "\n")
		}
	}

	drawSep("┌", "┬", "┐", "─")
	firstRow := true
	for _, rw := range rows {
		drawRow(rw.cells, rw.isHeader)
		if firstRow && rw.isHeader {
			drawSep("├", "┼", "┤", "─")
			firstRow = false
		}
	}
	drawSep("└", "┴", "┘", "─")
	r.buf.WriteString("\n")
}

func fitTableColumnWidths(desired []int, maxTableWidth int) []int {
	widths := append([]int(nil), desired...)
	if len(widths) == 0 || maxTableWidth <= 0 {
		return widths
	}

	colCount := len(widths)
	available := maxTableWidth - (3*colCount + 1)
	minWidth := 3
	if available < colCount*minWidth {
		minWidth = 1
	}
	if available < colCount*minWidth {
		for i := range widths {
			widths[i] = 1
		}
		return widths
	}

	total := 0
	for i, w := range widths {
		if widths[i] < minWidth {
			widths[i] = minWidth
			w = minWidth
		}
		total += w
	}
	if total <= available {
		return widths
	}

	for i := range widths {
		widths[i] = minWidth
	}

	remaining := available - colCount*minWidth
	for remaining > 0 {
		progressed := false
		for i := range widths {
			if widths[i] >= desired[i] {
				continue
			}
			widths[i]++
			remaining--
			progressed = true
			if remaining == 0 {
				break
			}
		}
		if !progressed {
			break
		}
	}

	return widths
}

func wrapTableCell(text string, maxWidth int) []string {
	if text == "" {
		return []string{""}
	}
	if maxWidth <= 0 {
		return []string{text}
	}

	wrapped := wrapANSI(text, maxWidth, "", 0)
	var lines []string
	for _, line := range strings.Split(wrapped, "\n") {
		if visualWidth(line) <= maxWidth {
			lines = append(lines, line)
			continue
		}
		lines = append(lines, hardWrapANSI(line, maxWidth)...)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func hardWrapANSI(text string, maxWidth int) []string {
	if maxWidth <= 0 || visualWidth(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	var line bytes.Buffer
	col := 0
	activeStyles := ""

	flushLine := func() {
		lines = append(lines, line.String())
		line.Reset()
		col = 0
		if activeStyles != "" {
			line.WriteString(activeStyles)
		}
	}

	for _, seg := range parseANSI(text) {
		if !seg.visible {
			line.WriteString(seg.text)
			if seg.text == ansiReset {
				activeStyles = ""
			} else {
				activeStyles += seg.text
			}
			continue
		}
		for _, ch := range seg.text {
			chW := runeVisualWidth(ch)
			if col > 0 && col+chW > maxWidth {
				flushLine()
			}
			line.WriteRune(ch)
			col += chW
		}
	}
	if line.Len() > 0 || len(lines) == 0 {
		lines = append(lines, line.String())
	}
	return lines
}

// ── Inline Renderers ────────────────────────────────────────────────────────

func (r *Renderer) renderInlineContent(n *parser.Node) {
	for _, child := range n.Children {
		r.renderNode(child)
	}
}

func getTextNode(n *parser.Node) (*parser.Node, bool) {
	for _, c := range n.Children {
		if c.Type == parser.NodeText {
			return c, true
		}
		if t, ok := getTextNode(c); ok {
			return t, true
		}
	}
	return nil, false
}

// ── ANSI-aware Word Wrapping ────────────────────────────────────────────────
// Mirrors Swift UITextView .byWordWrapping with lineSpacing.

type ansiSegment struct {
	text    string
	visible bool
}

func parseANSI(s string) []ansiSegment {
	var segs []ansiSegment
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) {
					c := s[j]
					j++
					if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
						break
					}
				}
			} else {
				j = i + 1
			}
			segs = append(segs, ansiSegment{text: s[i:j], visible: false})
			i = j
		} else {
			j := i + 1
			for j < len(s) && s[j] != '\033' {
				j++
			}
			segs = append(segs, ansiSegment{text: s[i:j], visible: true})
			i = j
		}
	}
	return segs
}

// wrapANSI wraps ANSI-styled text at word boundaries.
// lineSpacing controls blank lines between wrapped segments (mirrors Swift lineSpacing: 5).
func wrapANSI(text string, maxWidth int, indent string, lineSpacing int) string {
	if maxWidth <= 0 {
		return text
	}
	segs := parseANSI(text)
	if len(segs) == 0 {
		return ""
	}
	indentW := visualWidth(indent)
	availWidth := maxWidth - indentW
	if availWidth < 10 {
		availWidth = 10
	}

	var out bytes.Buffer
	var line bytes.Buffer
	col := 0
	activeStyles := ""
	lineHasContent := false

	flushLine := func() {
		out.Write(line.Bytes())
		out.WriteByte('\n')
		for i := 0; i < lineSpacing; i++ {
			out.WriteByte('\n')
		}
		line.Reset()
		col = 0
		lineHasContent = false
		line.WriteString(indent)
		if activeStyles != "" {
			line.WriteString(activeStyles)
		}
	}

	var word bytes.Buffer
	wordCol := 0

	flushWord := func() {
		if wordCol == 0 {
			return
		}
		if lineHasContent && col+wordCol > availWidth {
			flushLine()
		}
		line.Write(word.Bytes())
		col += wordCol
		lineHasContent = true
		word.Reset()
		wordCol = 0
	}

	for _, seg := range segs {
		if !seg.visible {
			word.WriteString(seg.text)
			if seg.text == ansiReset {
				activeStyles = ""
			} else {
				activeStyles += seg.text
			}
			continue
		}
		for _, ch := range seg.text {
			if ch == ' ' || ch == '\t' {
				flushWord()
				chW := runeVisualWidth(ch)
				if col+chW > availWidth && lineHasContent {
					flushLine()
				} else {
					line.WriteRune(ch)
					col += chW
					lineHasContent = true
				}
			} else {
				word.WriteRune(ch)
				wordCol += runeVisualWidth(ch)
			}
		}
	}
	flushWord()
	out.Write(line.Bytes())
	return out.String()
}

// ── Utilities ───────────────────────────────────────────────────────────────

// visualWidth returns the terminal display width of a string, ignoring ANSI escape sequences.
func visualWidth(s string) int {
	w := 0
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		w += runeVisualWidth(r)
	}
	return w
}

func runeVisualWidth(r rune) int {
	if r == 0 {
		return 0
	}
	if r < 32 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if isCombiningRune(r) {
		return 0
	}
	if isWideRune(r) {
		return 2
	}
	return 1
}

func isCombiningRune(r rune) bool {
	return (r >= 0x0300 && r <= 0x036f) ||
		(r >= 0x1ab0 && r <= 0x1aff) ||
		(r >= 0x1dc0 && r <= 0x1dff) ||
		(r >= 0x20d0 && r <= 0x20ff) ||
		(r >= 0xfe00 && r <= 0xfe0f) ||
		(r >= 0xfe20 && r <= 0xfe2f)
}

func isWideRune(r rune) bool {
	return (r >= 0x1100 && r <= 0x115f) ||
		(r >= 0x2329 && r <= 0x232a) ||
		(r >= 0x2e80 && r <= 0xa4cf) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x1f300 && r <= 0x1f64f) ||
		(r >= 0x1f900 && r <= 0x1f9ff) ||
		(r >= 0x20000 && r <= 0x3fffd)
}

// StripANSI removes all ANSI escape sequences from a string.
func StripANSI(s string) string {
	var buf strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		buf.WriteRune(r)
	}
	return buf.String()
}

// Render is a convenience function: parse + render in one call.
func Render(src string, width int, theme *Theme) string {
	doc := parser.Parse(src, parser.DefaultOption())
	r := New(theme, width)
	return r.Render(doc)
}
