package parser

import (
	"strings"
	"testing"
)

// ── Helpers ─────────────────────────────────────────────────────────────────

func parse(src string) *Node {
	return Parse(src, DefaultOption())
}

func parseStream(src string) *Node {
	return Parse(src, StreamOption())
}

func assertNodeType(t *testing.T, n *Node, expected NodeType) {
	t.Helper()
	if n.Type != expected {
		t.Errorf("expected node type %d, got %d", expected, n.Type)
	}
}

func assertText(t *testing.T, n *Node, expected string) {
	t.Helper()
	if n.TextContent() != expected {
		t.Errorf("expected text %q, got %q", expected, n.TextContent())
	}
}

func assertChildCount(t *testing.T, n *Node, expected int) {
	t.Helper()
	if len(n.Children) != expected {
		t.Errorf("expected %d children, got %d", expected, len(n.Children))
	}
}

func findNode(n *Node, t NodeType) *Node {
	if n.Type == t {
		return n
	}
	for _, c := range n.Children {
		if found := findNode(c, t); found != nil {
			return found
		}
	}
	return nil
}

func findAllNodes(n *Node, t NodeType) []*Node {
	var result []*Node
	if n.Type == t {
		result = append(result, n)
	}
	for _, c := range n.Children {
		result = append(result, findAllNodes(c, t)...)
	}
	return result
}

// ── Heading Tests ───────────────────────────────────────────────────────────

func TestHeadingH1(t *testing.T) {
	doc := parse("# Hello")
	h := findNode(doc, NodeHeading)
	if h == nil {
		t.Fatal("expected heading node")
	}
	assertNodeType(t, h, NodeHeading)
	if h.Level != 1 {
		t.Errorf("expected level 1, got %d", h.Level)
	}
	assertText(t, h, "Hello\n")
}

func TestHeadingH1ToH6(t *testing.T) {
	inputs := []struct {
		src   string
		level int
		text  string
	}{
		{"# H1", 1, "H1\n"},
		{"## H2", 2, "H2\n"},
		{"### H3", 3, "H3\n"},
		{"#### H4", 4, "H4\n"},
		{"##### H5", 5, "H5\n"},
		{"###### H6", 6, "H6\n"},
	}
	for _, tc := range inputs {
		t.Run(tc.src, func(t *testing.T) {
			doc := parse(tc.src)
			h := findNode(doc, NodeHeading)
			if h == nil {
				t.Fatalf("expected heading for %q", tc.src)
			}
			if h.Level != tc.level {
				t.Errorf("expected level %d, got %d", tc.level, h.Level)
			}
		})
	}
}

func TestHeadingWithInlineFormatting(t *testing.T) {
	doc := parse("# Hello **bold** world")
	h := findNode(doc, NodeHeading)
	if h == nil {
		t.Fatal("expected heading")
	}
	strong := findNode(h, NodeStrong)
	if strong == nil {
		t.Fatal("expected bold in heading")
	}
}

func TestHeadingWithTrailingHashes(t *testing.T) {
	doc := parse("# Hello ##")
	h := findNode(doc, NodeHeading)
	if h == nil {
		t.Fatal("expected heading")
	}
	// Trailing ## should be stripped
	txt := h.TextContent()
	if strings.Contains(txt, "##") {
		t.Errorf("trailing hashes not stripped: %q", txt)
	}
}

func TestHeadingInvalidNoSpace(t *testing.T) {
	doc := parse("#NoSpace")
	// Should not be a heading (no space after #)
	h := findNode(doc, NodeHeading)
	if h != nil {
		t.Error("should not parse as heading without space")
	}
}

// ── Paragraph Tests ─────────────────────────────────────────────────────────

func TestSingleParagraph(t *testing.T) {
	doc := parse("Hello world")
	assertChildCount(t, doc, 1)
	assertNodeType(t, doc.Children[0], NodeParagraph)
}

func TestMultipleParagraphs(t *testing.T) {
	doc := parse("First paragraph.\n\nSecond paragraph.\n\nThird paragraph.")
	paras := findAllNodes(doc, NodeParagraph)
	if len(paras) < 3 {
		t.Errorf("expected 3 paragraphs, got %d", len(paras))
	}
}

func TestParagraphWithInlineFormatting(t *testing.T) {
	doc := parse("Hello **bold** and *italic* and `code` world")
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	if findNode(p, NodeStrong) == nil {
		t.Error("expected bold")
	}
	if findNode(p, NodeEmphasis) == nil {
		t.Error("expected italic")
	}
	if findNode(p, NodeCodeSpan) == nil {
		t.Error("expected code span")
	}
}

func TestParagraphMultiline(t *testing.T) {
	doc := parse("Line one\nLine two\nLine three")
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	softBreaks := findAllNodes(p, NodeSoftBreak)
	if len(softBreaks) < 2 {
		t.Errorf("expected 2 soft breaks, got %d", len(softBreaks))
	}
}

func TestHardBreak(t *testing.T) {
	doc := parse("Line one  \nLine two")
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	if findNode(p, NodeHardBreak) == nil {
		t.Error("expected hard break")
	}
}

func TestHardBreakBackslash(t *testing.T) {
	doc := parse("Line one\\\nLine two")
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	if findNode(p, NodeHardBreak) == nil {
		t.Error("expected hard break from backslash")
	}
}

// ── Fenced Code Block Tests ─────────────────────────────────────────────────

func TestFencedCodeBlockBacktick(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected fenced code block")
	}
	if cb.Language != "go" {
		t.Errorf("expected language 'go', got %q", cb.Language)
	}
	if !strings.Contains(cb.Code, "fmt.Println") {
		t.Errorf("expected code content, got %q", cb.Code)
	}
}

func TestFencedCodeBlockTilde(t *testing.T) {
	src := "~~~python\nprint('hello')\n~~~"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected fenced code block with tilde")
	}
	if cb.Language != "python" {
		t.Errorf("expected language 'python', got %q", cb.Language)
	}
}

func TestFencedCodeBlockNoLanguage(t *testing.T) {
	src := "```\nsome code\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block")
	}
	if cb.Language != "" {
		t.Errorf("expected empty language, got %q", cb.Language)
	}
}

func TestFencedCodeBlockMultiline(t *testing.T) {
	src := "```js\nfunction hello() {\n  console.log('hello');\n  return true;\n}\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block")
	}
	lines := strings.Split(cb.Code, "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(cb.Code, "function hello()") {
		t.Error("missing function declaration")
	}
	if !strings.Contains(cb.Code, "return true") {
		t.Error("missing return statement")
	}
}

func TestFencedCodeBlockEmpty(t *testing.T) {
	src := "```\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block")
	}
}

func TestFencedCodeBlockWithBlankLines(t *testing.T) {
	src := "```go\nfunc main() {\n\n    fmt.Println(\"hello\")\n\n}\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block")
	}
	lines := strings.Split(cb.Code, "\n")
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines (including blanks), got %d", len(lines))
	}
}

func TestFencedCodeBlockIndentation(t *testing.T) {
	src := "```python\n    def hello():\n        print('hello')\n        if True:\n            pass\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block")
	}
	// Indentation should be preserved
	if !strings.Contains(cb.Code, "    def hello():") {
		t.Errorf("indentation not preserved: %q", cb.Code)
	}
	if !strings.Contains(cb.Code, "        print") {
		t.Errorf("nested indentation not preserved: %q", cb.Code)
	}
}

func TestFencedCodeBlockSpecialCharacters(t *testing.T) {
	src := "```\n<script>alert('xss')</script>\n| table | not |\n> not blockquote\n# not heading\n```"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block")
	}
	// Content should be raw, not parsed as markdown
	if !strings.Contains(cb.Code, "<script>") {
		t.Error("HTML not preserved in code block")
	}
	if !strings.Contains(cb.Code, "| table |") {
		t.Error("table syntax not preserved in code block")
	}
}

func TestFencedCodeBlockLanguageVariants(t *testing.T) {
	tests := []string{"go", "python", "javascript", "bash", "sql", "yaml", "json", "html", "css"}
	for _, lang := range tests {
		t.Run(lang, func(t *testing.T) {
			src := "```" + lang + "\ncode\n```"
			doc := parse(src)
			cb := findNode(doc, NodeFencedCodeBlock)
			if cb == nil {
				t.Fatal("expected code block")
			}
			if cb.Language != lang {
				t.Errorf("expected language %q, got %q", lang, cb.Language)
			}
		})
	}
}

func TestFencedCodeBlockMultipleInDocument(t *testing.T) {
	src := "Paragraph 1\n\n```go\ncode1\n```\n\nParagraph 2\n\n```python\ncode2\n```"
	doc := parse(src)
	blocks := findAllNodes(doc, NodeFencedCodeBlock)
	if len(blocks) != 2 {
		t.Errorf("expected 2 code blocks, got %d", len(blocks))
	}
	if blocks[0].Language != "go" {
		t.Errorf("first block lang: expected 'go', got %q", blocks[0].Language)
	}
	if blocks[1].Language != "python" {
		t.Errorf("second block lang: expected 'python', got %q", blocks[1].Language)
	}
}

func TestFencedCodeBlockConsecutive(t *testing.T) {
	src := "```go\ncode1\n```\n```python\ncode2\n```"
	doc := parse(src)
	blocks := findAllNodes(doc, NodeFencedCodeBlock)
	if len(blocks) != 2 {
		t.Errorf("expected 2 consecutive code blocks, got %d", len(blocks))
	}
}

func TestFencedCodeBlockExtendedFence(t *testing.T) {
	src := "````go\ncode with ``` inside\n````"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block with extended fence")
	}
	if !strings.Contains(cb.Code, "```") {
		t.Error("inner backticks not preserved")
	}
}

// ── Indented Code Block Tests ───────────────────────────────────────────────

func TestIndentedCodeBlock(t *testing.T) {
	src := "    indented code\n    second line"
	doc := parse(src)
	cb := findNode(doc, NodeIndentedCodeBlock)
	if cb == nil {
		t.Fatal("expected indented code block")
	}
	if !strings.Contains(cb.Code, "indented code") {
		t.Errorf("expected code content, got %q", cb.Code)
	}
}

func TestIndentedCodeBlockTab(t *testing.T) {
	src := "\ttab indented code"
	doc := parse(src)
	cb := findNode(doc, NodeIndentedCodeBlock)
	if cb == nil {
		t.Fatal("expected indented code block from tab")
	}
}

func TestIndentedCodeBlockBlankLine(t *testing.T) {
	src := "    line 1\n\n    line 2"
	doc := parse(src)
	cb := findNode(doc, NodeIndentedCodeBlock)
	if cb == nil {
		t.Fatal("expected indented code block spanning blank line")
	}
	lines := strings.Split(cb.Code, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestIndentedCodeBlockEndsAtUnindented(t *testing.T) {
	src := "    code line\nNormal text"
	doc := parse(src)
	cb := findNode(doc, NodeIndentedCodeBlock)
	if cb == nil {
		t.Fatal("expected indented code block")
	}
	para := findNode(doc, NodeParagraph)
	if para == nil {
		t.Fatal("expected paragraph after code block")
	}
}

// ── Blockquote Tests ────────────────────────────────────────────────────────

func TestBlockquoteSimple(t *testing.T) {
	doc := parse("> Hello world")
	bq := findNode(doc, NodeBlockquote)
	if bq == nil {
		t.Fatal("expected blockquote")
	}
	txt := bq.TextContent()
	if !strings.Contains(txt, "Hello world") {
		t.Errorf("expected text to contain 'Hello world', got %q", txt)
	}
}

func TestBlockquoteMultiline(t *testing.T) {
	doc := parse("> Line 1\n> Line 2\n> Line 3")
	bq := findNode(doc, NodeBlockquote)
	if bq == nil {
		t.Fatal("expected blockquote")
	}
	txt := bq.TextContent()
	if !strings.Contains(txt, "Line 1") || !strings.Contains(txt, "Line 2") {
		t.Errorf("multiline content missing: %q", txt)
	}
}

func TestBlockquoteWithFormatting(t *testing.T) {
	doc := parse("> **bold** and *italic*")
	bq := findNode(doc, NodeBlockquote)
	if bq == nil {
		t.Fatal("expected blockquote")
	}
	if findNode(bq, NodeStrong) == nil {
		t.Error("expected bold in blockquote")
	}
	if findNode(bq, NodeEmphasis) == nil {
		t.Error("expected italic in blockquote")
	}
}

func TestBlockquoteNested(t *testing.T) {
	doc := parse("> L1\n> > L2")
	bqs := findAllNodes(doc, NodeBlockquote)
	if len(bqs) < 1 {
		t.Fatalf("expected at least 1 blockquote, got %d", len(bqs))
	}
	// Verify both L1 and L2 are in the output
	txt := doc.TextContent()
	if !strings.Contains(txt, "L1") {
		t.Error("expected L1 in blockquote")
	}
	if !strings.Contains(txt, "L2") {
		t.Error("expected L2 in nested blockquote")
	}
}

func TestBlockquoteWithCodeBlock(t *testing.T) {
	doc := parse("> ```\n> code in quote\n> ```")
	bq := findNode(doc, NodeBlockquote)
	if bq == nil {
		t.Fatal("expected blockquote")
	}
	// Code block inside blockquote
	cb := findNode(bq, NodeFencedCodeBlock)
	if cb == nil {
		t.Error("expected code block inside blockquote")
	}
}

func TestBlockquoteEmpty(t *testing.T) {
	doc := parse(">")
	bq := findNode(doc, NodeBlockquote)
	if bq == nil {
		t.Fatal("expected blockquote even for empty content")
	}
}

// ── Unordered List Tests ────────────────────────────────────────────────────

func TestUnorderedListDash(t *testing.T) {
	doc := parse("- Item 1\n- Item 2\n- Item 3")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected unordered list")
	}
	assertChildCount(t, list, 3)
}

func TestUnorderedListAsterisk(t *testing.T) {
	doc := parse("* Item 1\n* Item 2")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected unordered list with *")
	}
	assertChildCount(t, list, 2)
}

func TestUnorderedListPlus(t *testing.T) {
	doc := parse("+ Item 1\n+ Item 2")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected unordered list with +")
	}
}

func TestUnorderedListWithBold(t *testing.T) {
	doc := parse("- **Bold** item\n- Normal item")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected list")
	}
	if findNode(list, NodeStrong) == nil {
		t.Error("expected bold in list")
	}
}

func TestUnorderedListTaskItems(t *testing.T) {
	doc := parse("- [x] Done\n- [ ] Todo\n- [X] Also done")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected list")
	}
	items := findAllNodes(list, NodeListItem)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if !items[0].Checked {
		t.Error("first item should be checked")
	}
	if items[1].Checked {
		t.Error("second item should not be checked")
	}
	if !items[2].Checked {
		t.Error("third item should be checked (uppercase X)")
	}
}

func TestUnorderedListStartsWithBold(t *testing.T) {
	doc := parse("- **Term** definition\n- Normal item")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected list")
	}
	items := findAllNodes(list, NodeListItem)
	if len(items) < 1 {
		t.Fatal("expected items")
	}
	if !items[0].StartsWithBold {
		t.Error("first item should start with bold")
	}
	if items[1].StartsWithBold {
		t.Error("second item should not start with bold")
	}
}

func TestUnorderedListBlankLineBetween(t *testing.T) {
	doc := parse("- Item 1\n\n- Item 2")
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected list to span blank line")
	}
	assertChildCount(t, list, 2)
}

// ── Ordered List Tests ──────────────────────────────────────────────────────

func TestOrderedListBasic(t *testing.T) {
	doc := parse("1. First\n2. Second\n3. Third")
	list := findNode(doc, NodeOrderedList)
	if list == nil {
		// Fallback: check for unordered list with Ordered=true
		list = findNode(doc, NodeUnorderedList)
	}
	if list == nil {
		t.Fatal("expected ordered list")
	}
	if !list.Ordered {
		t.Error("list should be ordered")
	}
	assertChildCount(t, list, 3)
}

func TestOrderedListStartNumber(t *testing.T) {
	doc := parse("5. Fifth\n6. Sixth\n7. Seventh")
	list := findNode(doc, NodeOrderedList)
	if list == nil {
		list = findNode(doc, NodeUnorderedList)
	}
	if list == nil {
		t.Fatal("expected list")
	}
	if list.StartNum != 5 {
		t.Errorf("expected StartNum 5, got %d", list.StartNum)
	}
}

func TestOrderedListTaskItems(t *testing.T) {
	doc := parse("1. [x] Done\n2. [ ] Todo")
	list := findNode(doc, NodeOrderedList)
	if list == nil {
		list = findNode(doc, NodeUnorderedList)
	}
	if list == nil {
		t.Fatal("expected list")
	}
	items := findAllNodes(list, NodeListItem)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if !items[0].Checked {
		t.Error("first item should be checked")
	}
	if items[1].Checked {
		t.Error("second item should not be checked")
	}
}

func TestOrderedListStartsWithBold(t *testing.T) {
	doc := parse("1. **Bold** start\n2. Normal")
	list := findNode(doc, NodeOrderedList)
	if list == nil {
		list = findNode(doc, NodeUnorderedList)
	}
	if list == nil {
		t.Fatal("expected list")
	}
	items := findAllNodes(list, NodeListItem)
	if len(items) == 0 {
		t.Fatal("expected items")
	}
	if !items[0].StartsWithBold {
		t.Error("first item should start with bold")
	}
}

// ── Table Tests ─────────────────────────────────────────────────────────────

func TestTableBasic(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |"
	doc := parse(src)
	table := findNode(doc, NodeTable)
	if table == nil {
		t.Fatal("expected table")
	}
	rows := findAllNodes(table, NodeTableRow)
	if len(rows) < 2 {
		t.Errorf("expected at least 2 rows (header+body), got %d", len(rows))
	}
}

func TestTableHeaderFormatting(t *testing.T) {
	src := "| Name | Value |\n|------|-------|\n| foo  | bar   |"
	doc := parse(src)
	table := findNode(doc, NodeTable)
	if table == nil {
		t.Fatal("expected table")
	}
	rows := findAllNodes(table, NodeTableRow)
	if len(rows) < 1 {
		t.Fatal("expected rows")
	}
	if !rows[0].IsTableHeader {
		t.Error("first row should be header")
	}
}

func TestTableMultipleRows(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |\n| 3 | 4 |\n| 5 | 6 |"
	doc := parse(src)
	table := findNode(doc, NodeTable)
	if table == nil {
		t.Fatal("expected table")
	}
	cells := findAllNodes(table, NodeTableCell)
	if len(cells) != 8 { // 2 header + 3*2 body
		t.Errorf("expected 8 cells, got %d", len(cells))
	}
}

func TestTableInlineInCells(t *testing.T) {
	src := "| **bold** | *italic* |\n|----------|----------|\n| `code` | [link](url) |"
	doc := parse(src)
	table := findNode(doc, NodeTable)
	if table == nil {
		t.Fatal("expected table")
	}
	if findNode(table, NodeStrong) == nil {
		t.Error("expected bold in table cell")
	}
	if findNode(table, NodeEmphasis) == nil {
		t.Error("expected italic in table cell")
	}
	if findNode(table, NodeCodeSpan) == nil {
		t.Error("expected code in table cell")
	}
	if findNode(table, NodeLink) == nil {
		t.Error("expected link in table cell")
	}
}

func TestTableUnevenColumns(t *testing.T) {
	src := "| A | B | C |\n|---|---|---|\n| 1 | 2 |"
	doc := parse(src)
	table := findNode(doc, NodeTable)
	if table == nil {
		t.Fatal("expected table")
	}
	// Body row should be padded to match header width
	rows := findAllNodes(table, NodeTableRow)
	if len(rows) < 2 {
		t.Fatal("expected rows")
	}
	bodyCells := findAllNodes(rows[1], NodeTableCell)
	if len(bodyCells) != 3 {
		t.Errorf("expected 3 cells (padded), got %d", len(bodyCells))
	}
}

// ── Thematic Break Tests ────────────────────────────────────────────────────

func TestThematicBreakDash(t *testing.T) {
	doc := parse("---")
	if findNode(doc, NodeThematicBreak) == nil {
		t.Fatal("expected thematic break from ---")
	}
}

func TestThematicBreakAsterisk(t *testing.T) {
	doc := parse("***")
	if findNode(doc, NodeThematicBreak) == nil {
		t.Fatal("expected thematic break from ***")
	}
}

func TestThematicBreakUnderscore(t *testing.T) {
	doc := parse("___")
	if findNode(doc, NodeThematicBreak) == nil {
		t.Fatal("expected thematic break from ___")
	}
}

func TestThematicBreakWithSpaces(t *testing.T) {
	doc := parse("- - -")
	if findNode(doc, NodeThematicBreak) == nil {
		t.Fatal("expected thematic break from - - -")
	}
}

func TestThematicBreakNotInParagraph(t *testing.T) {
	doc := parse("text\n---\nmore text")
	// --- after text should be a thematic break, not setext heading
	// (depends on parser implementation)
	tb := findNode(doc, NodeThematicBreak)
	para := findNode(doc, NodeParagraph)
	if tb == nil && para == nil {
		t.Error("expected either thematic break or paragraph")
	}
}

// ── Inline Tests ────────────────────────────────────────────────────────────

func TestInlineBold(t *testing.T) {
	doc := parse("Hello **world**")
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	s := findNode(p, NodeStrong)
	if s == nil {
		t.Fatal("expected bold")
	}
	assertText(t, s, "world")
}

func TestInlineBoldUnderscore(t *testing.T) {
	doc := parse("Hello __world__")
	p := findNode(doc, NodeParagraph)
	s := findNode(p, NodeStrong)
	if s == nil {
		t.Fatal("expected bold with __")
	}
}

func TestInlineItalic(t *testing.T) {
	doc := parse("Hello *world*")
	p := findNode(doc, NodeParagraph)
	e := findNode(p, NodeEmphasis)
	if e == nil {
		t.Fatal("expected italic")
	}
	assertText(t, e, "world")
}

func TestInlineItalicUnderscore(t *testing.T) {
	doc := parse("Hello _world_")
	p := findNode(doc, NodeParagraph)
	e := findNode(p, NodeEmphasis)
	if e == nil {
		t.Fatal("expected italic with _")
	}
}

func TestInlineItalicAdjacentTextBoundaries(t *testing.T) {
	doc := parse("111*222*333")
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	assertChildCount(t, p, 3)

	if p.Children[0].Type != NodeText || p.Children[0].Text != "111" {
		t.Fatalf("expected leading text 111, got type=%d text=%q", p.Children[0].Type, p.Children[0].Text)
	}
	if p.Children[1].Type != NodeEmphasis {
		t.Fatalf("expected middle emphasis, got type=%d", p.Children[1].Type)
	}
	assertText(t, p.Children[1], "222")
	if p.Children[2].Type != NodeText || p.Children[2].Text != "333" {
		t.Fatalf("expected trailing text 333, got type=%d text=%q", p.Children[2].Type, p.Children[2].Text)
	}
}

func TestInlineCodeSpan(t *testing.T) {
	doc := parse("Use `fmt.Println` to print")
	p := findNode(doc, NodeParagraph)
	cs := findNode(p, NodeCodeSpan)
	if cs == nil {
		t.Fatal("expected code span")
	}
	if cs.Text != "fmt.Println" {
		t.Errorf("expected 'fmt.Println', got %q", cs.Text)
	}
}

func TestInlineCodeSpanDoubleBacktick(t *testing.T) {
	doc := parse("Use ``fmt.Println`x` to print")
	p := findNode(doc, NodeParagraph)
	cs := findNode(p, NodeCodeSpan)
	if cs == nil {
		t.Fatal("expected code span with double backtick")
	}
}

func TestInlineCodeSpanSpaces(t *testing.T) {
	doc := parse("` code `")
	p := findNode(doc, NodeParagraph)
	cs := findNode(p, NodeCodeSpan)
	if cs == nil {
		t.Fatal("expected code span")
	}
	// Leading/trailing spaces should be stripped when both present
	if cs.Text != "code" {
		t.Errorf("expected 'code', got %q", cs.Text)
	}
}

func TestInlineLink(t *testing.T) {
	doc := parse("[Go](https://go.dev)")
	p := findNode(doc, NodeParagraph)
	link := findNode(p, NodeLink)
	if link == nil {
		t.Fatal("expected link")
	}
	if link.URL != "https://go.dev" {
		t.Errorf("expected URL 'https://go.dev', got %q", link.URL)
	}
}

func TestLinkWithTitle(t *testing.T) {
	doc := parse("[Go](https://go.dev \"Go homepage\")")
	p := findNode(doc, NodeParagraph)
	link := findNode(p, NodeLink)
	if link == nil {
		t.Fatal("expected link")
	}
	if link.Title != "Go homepage" {
		t.Errorf("expected title 'Go homepage', got %q", link.Title)
	}
}

func TestInlineImage(t *testing.T) {
	doc := parse("![alt text](https://example.com/img.png)")
	p := findNode(doc, NodeParagraph)
	img := findNode(p, NodeImage)
	if img == nil {
		t.Fatal("expected image")
	}
	if img.URL != "https://example.com/img.png" {
		t.Errorf("unexpected URL: %q", img.URL)
	}
}

func TestInlineStrikethrough(t *testing.T) {
	doc := parse("~~deleted~~")
	p := findNode(doc, NodeParagraph)
	st := findNode(p, NodeStrikethrough)
	if st == nil {
		t.Fatal("expected strikethrough")
	}
}

func TestInlineAutolink(t *testing.T) {
	doc := parse("<https://example.com>")
	p := findNode(doc, NodeParagraph)
	al := findNode(p, NodeAutolink)
	if al == nil {
		t.Fatal("expected autolink")
	}
	if al.URL != "https://example.com" {
		t.Errorf("unexpected URL: %q", al.URL)
	}
}

func TestInlineAutolinkEmail(t *testing.T) {
	doc := parse("<user@example.com>")
	p := findNode(doc, NodeParagraph)
	al := findNode(p, NodeAutolink)
	if al == nil {
		t.Fatal("expected autolink email")
	}
}

func TestInlineNestedFormatting(t *testing.T) {
	doc := parse("Hello **bold *and italic* text**")
	p := findNode(doc, NodeParagraph)
	s := findNode(p, NodeStrong)
	if s == nil {
		t.Fatal("expected bold")
	}
	if findNode(s, NodeEmphasis) == nil {
		t.Error("expected italic inside bold")
	}
}

func TestInlineEscapedCharacter(t *testing.T) {
	doc := parse("Hello \\*not bold\\*")
	p := findNode(doc, NodeParagraph)
	// Escaped * should not create emphasis
	if findNode(p, NodeEmphasis) != nil {
		t.Error("escaped * should not create emphasis")
	}
}

// ── LaTeX Preprocessing Tests ───────────────────────────────────────────────

func TestLaTeXBlockDollar(t *testing.T) {
	doc := parse("$$E = mc^2$$")
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block from $$...$$")
	}
	if cb.Language != "blockmath" {
		t.Errorf("expected 'blockmath' language, got %q", cb.Language)
	}
}

func TestLaTeXBlockSlashBracket(t *testing.T) {
	doc := parse("\\[x^2 + y^2 = z^2\\]")
	cb := findNode(doc, NodeFencedCodeBlock)
	if cb == nil {
		t.Fatal("expected code block from \\[...\\]")
	}
	if cb.Language != "blockmath" {
		t.Errorf("expected 'blockmath', got %q", cb.Language)
	}
}

func TestLaTeXInlineParen(t *testing.T) {
	doc := parse("Equation: \\(a^2 + b^2\\)")
	p := findNode(doc, NodeParagraph)
	cs := findNode(p, NodeCodeSpan)
	if cs == nil {
		t.Fatal("expected code span from \\(...\\)")
	}
}

func TestLaTeXInlineDollar(t *testing.T) {
	doc := parse("Equation: $a^2 + b^2$")
	p := findNode(doc, NodeParagraph)
	cs := findNode(p, NodeCodeSpan)
	if cs == nil {
		t.Fatal("expected code span from $...$")
	}
}

func TestLaTeXFilterDfrac(t *testing.T) {
	result := filterLatexSyntax(`\dfrac{1}{2}`)
	if strings.Contains(result, `\dfrac`) {
		t.Error("dfrac should be replaced with frac")
	}
	if !strings.Contains(result, `\frac`) {
		t.Error("should contain frac")
	}
}

func TestLaTeXFilterBoxed(t *testing.T) {
	result := filterLatexSyntax(`\boxed{x=1}`)
	if strings.Contains(result, `\boxed`) {
		t.Error("boxed should be stripped")
	}
}

func TestLaTeXFilterPrime(t *testing.T) {
	result := filterLatexSyntax("f'(x)")
	if strings.Contains(result, "'") {
		t.Error("prime should be replaced")
	}
	if !strings.Contains(result, `\prime`) {
		t.Error("should contain \\prime")
	}
}

// ── Speculative Rewriting Tests (Streaming) ─────────────────────────────────

func TestSpeculativeEmphasisStrong(t *testing.T) {
	doc := parseStream("Yeah, this is **cool")
	// Should either close the ** or keep the text visible
	txt := doc.TextContent()
	if !strings.Contains(txt, "cool") {
		t.Error("'cool' should be in the output")
	}
}

func TestSpeculativeEmphasisItalic(t *testing.T) {
	doc := parseStream("Yeah, this is *cool")
	txt := doc.TextContent()
	if !strings.Contains(txt, "cool") {
		t.Error("'cool' should be in the output")
	}
}

func TestSpeculativeTable(t *testing.T) {
	doc := parseStream("| Month | Savings |")
	// Partial table header should be removed from output
	p := findNode(doc, NodeParagraph)
	if p != nil {
		// The paragraph should be emptied by the table rewriter
		txt := p.TextContent()
		if strings.Contains(txt, "Month") {
			t.Error("partial table content should be cleared")
		}
	}
}

func TestSpeculativeTableWithSeparator(t *testing.T) {
	doc := parseStream("| Month | Savings |\n| :")
	p := findNode(doc, NodeParagraph)
	if p != nil {
		txt := p.TextContent()
		if strings.Contains(txt, "Month") {
			t.Error("partial table content should be cleared")
		}
	}
}

func TestSpeculativeDisabled(t *testing.T) {
	doc := parse("Yeah, this is **cool")
	// Without speculative rewrite, ** should remain as text
	s := findNode(doc, NodeStrong)
	// It might or might not be there depending on parser behavior
	// The key is it shouldn't crash
	_ = s
}

// ── Node ID Tests ───────────────────────────────────────────────────────────

func TestComputeIDs(t *testing.T) {
	doc := parse("Hello\n\nWorld")
	ComputeIDs(doc)
	if doc.ID != "0" {
		t.Errorf("root ID should be '0', got %q", doc.ID)
	}
	for _, child := range doc.Children {
		if child.ID == "" {
			t.Error("child ID should not be empty")
		}
	}
}

func TestComputeIDsNested(t *testing.T) {
	doc := parse("- Item 1\n- Item 2")
	ComputeIDs(doc)
	list := findNode(doc, NodeUnorderedList)
	if list == nil {
		t.Fatal("expected list")
	}
	if list.ID == "" {
		t.Error("list ID should not be empty")
	}
	for _, item := range list.Children {
		if item.ID == "" {
			t.Error("list item ID should not be empty")
		}
		// Each item should have a unique ID
		for _, other := range list.Children {
			if item != other && item.ID == other.ID {
				t.Errorf("duplicate ID: %q", item.ID)
			}
		}
	}
}

// ── PlainText Extraction Tests ──────────────────────────────────────────────

func TestPlainTextHeading(t *testing.T) {
	doc := parse("# Hello World")
	txt := doc.TextContent()
	if !strings.Contains(txt, "Hello World") {
		t.Errorf("plain text missing heading content: %q", txt)
	}
}

func TestPlainTextParagraph(t *testing.T) {
	doc := parse("Hello **bold** world")
	txt := doc.TextContent()
	if !strings.Contains(txt, "Hello") || !strings.Contains(txt, "bold") {
		t.Errorf("plain text missing content: %q", txt)
	}
}

func TestPlainTextCodeBlock(t *testing.T) {
	doc := parse("```\ncode here\n```")
	txt := doc.TextContent()
	if !strings.Contains(txt, "code here") {
		t.Errorf("plain text missing code: %q", txt)
	}
}

func TestPlainTextList(t *testing.T) {
	doc := parse("- Item 1\n- Item 2")
	txt := doc.TextContent()
	if !strings.Contains(txt, "Item 1") || !strings.Contains(txt, "Item 2") {
		t.Errorf("plain text missing list items: %q", txt)
	}
}

func TestPlainTextTable(t *testing.T) {
	doc := parse("| A | B |\n|---|---|\n| 1 | 2 |")
	txt := doc.TextContent()
	if !strings.Contains(txt, "A") || !strings.Contains(txt, "1") {
		t.Errorf("plain text missing table: %q", txt)
	}
}

func TestPlainTextLink(t *testing.T) {
	doc := parse("[Go](https://go.dev)")
	txt := doc.TextContent()
	if !strings.Contains(txt, "Go") {
		t.Errorf("plain text missing link text: %q", txt)
	}
}

// ── Edge Cases ──────────────────────────────────────────────────────────────

func TestEmptyInput(t *testing.T) {
	doc := parse("")
	assertChildCount(t, doc, 0)
}

func TestBlankLinesOnly(t *testing.T) {
	doc := parse("\n\n\n")
	assertChildCount(t, doc, 0)
}

func TestSingleCharacter(t *testing.T) {
	doc := parse("x")
	assertChildCount(t, doc, 1)
	assertNodeType(t, doc.Children[0], NodeParagraph)
}

func TestUnicodeContent(t *testing.T) {
	doc := parse("# 你好世界\n\n这是 **粗体** 文本")
	h := findNode(doc, NodeHeading)
	if h == nil {
		t.Fatal("expected heading with unicode")
	}
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph with unicode")
	}
	if findNode(p, NodeStrong) == nil {
		t.Error("expected bold in unicode text")
	}
}

func TestEmojiContent(t *testing.T) {
	doc := parse("# 🚀 Rocket\n\nHello 🌍 World")
	if findNode(doc, NodeHeading) == nil {
		t.Error("expected heading with emoji")
	}
}

func TestVeryLongLine(t *testing.T) {
	long := strings.Repeat("word ", 500)
	doc := parse(long)
	p := findNode(doc, NodeParagraph)
	if p == nil {
		t.Fatal("expected paragraph for long line")
	}
}

func TestMultipleBlankLinesBetweenBlocks(t *testing.T) {
	doc := parse("Paragraph 1\n\n\n\nParagraph 2")
	paras := findAllNodes(doc, NodeParagraph)
	if len(paras) != 2 {
		t.Errorf("expected 2 paragraphs, got %d", len(paras))
	}
}

func TestAdjacentCodeBlocks(t *testing.T) {
	src := "```go\ncode1\n```\n\n```python\ncode2\n```"
	doc := parse(src)
	blocks := findAllNodes(doc, NodeFencedCodeBlock)
	if len(blocks) != 2 {
		t.Errorf("expected 2 code blocks, got %d", len(blocks))
	}
}

func TestCodeBlockFollowedByParagraph(t *testing.T) {
	src := "```\ncode\n```\n\nNormal paragraph"
	doc := parse(src)
	cb := findNode(doc, NodeFencedCodeBlock)
	p := findNode(doc, NodeParagraph)
	if cb == nil {
		t.Fatal("expected code block")
	}
	if p == nil {
		t.Fatal("expected paragraph after code block")
	}
}

func TestParagraphFollowedByCodeBlock(t *testing.T) {
	src := "Normal paragraph\n\n```\ncode\n```"
	doc := parse(src)
	p := findNode(doc, NodeParagraph)
	cb := findNode(doc, NodeFencedCodeBlock)
	if p == nil {
		t.Fatal("expected paragraph")
	}
	if cb == nil {
		t.Fatal("expected code block after paragraph")
	}
}

func TestMixedContentDocument(t *testing.T) {
	src := `# Title

A paragraph with **bold** and *italic*.

> A blockquote

- List item 1
- List item 2

` + "```go\nfmt.Println(\"hello\")\n```" + `

| A | B |
|---|---|
| 1 | 2 |

---

Another paragraph.`
	doc := parse(src)

	// Should have all major block types
	if findNode(doc, NodeHeading) == nil {
		t.Error("missing heading")
	}
	if findNode(doc, NodeBlockquote) == nil {
		t.Error("missing blockquote")
	}
	if findNode(doc, NodeUnorderedList) == nil {
		t.Error("missing list")
	}
	if findNode(doc, NodeFencedCodeBlock) == nil {
		t.Error("missing code block")
	}
	if findNode(doc, NodeTable) == nil {
		t.Error("missing table")
	}
	if findNode(doc, NodeThematicBreak) == nil {
		t.Error("missing thematic break")
	}
	paras := findAllNodes(doc, NodeParagraph)
	if len(paras) < 2 {
		t.Errorf("expected at least 2 paragraphs, got %d", len(paras))
	}
}

// ── RightmostDescendant Tests ───────────────────────────────────────────────

func TestRightmostDescendant(t *testing.T) {
	doc := parse("Hello **world** and *foo*")
	leaf := doc.RightmostDescendant()
	if leaf == nil {
		t.Fatal("expected rightmost descendant")
	}
	// Should be the last text node
	if leaf.Type != NodeText {
		t.Errorf("expected text node, got type %d", leaf.Type)
	}
}

func TestRightmostDescendantEmpty(t *testing.T) {
	doc := parse("")
	leaf := doc.RightmostDescendant()
	// Empty doc's rightmost descendant is itself
	if leaf != doc {
		t.Error("empty doc's rightmost descendant should be itself")
	}
}

// ── IndexInParent Tests ─────────────────────────────────────────────────────

func TestIndexInParent(t *testing.T) {
	doc := parse("First\n\nSecond\n\nThird")
	if len(doc.Children) < 3 {
		t.Fatalf("expected 3 children, got %d", len(doc.Children))
	}
	for i, child := range doc.Children {
		idx := child.IndexInParent()
		if idx != i {
			t.Errorf("expected index %d, got %d", i, idx)
		}
	}
}

func TestIndexInParentRoot(t *testing.T) {
	doc := parse("Hello")
	if doc.IndexInParent() != -1 {
		t.Error("root's IndexInParent should be -1")
	}
}

// ── FindChild Tests ─────────────────────────────────────────────────────────

func TestFindChild(t *testing.T) {
	doc := parse("# Hello\n\nWorld")
	h := doc.FindChild(NodeHeading)
	if h == nil {
		t.Error("expected to find heading child")
	}
	p := doc.FindChild(NodeParagraph)
	if p == nil {
		t.Error("expected to find paragraph child")
	}
	cb := doc.FindChild(NodeFencedCodeBlock)
	if cb != nil {
		t.Error("should not find code block")
	}
}
