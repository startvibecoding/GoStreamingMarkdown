package parser

import (
	"regexp"
	"strings"
	"unicode"
)

// ── Preprocessor (LaTeX) ────────────────────────────────────────────────────
// Mirrors Swift LaTexPreProcessorImpl.

var (
	reDollarBlockMath = regexp.MustCompile(`(?m)^\s*\$\$([\s\S]*?)\$\$\s*$`)
	reDollarInline    = regexp.MustCompile(`\$([^$\n]+?)\$`)
	reSlashBracket    = regexp.MustCompile(`(?m)^\s*\\\[([\s\S]*?)\\\]\s*$`)
	reInlineParen     = regexp.MustCompile(`\\\([\s\S]*?\\\)`)

	// LaTeX syntax filters (matching Swift filteringUnsupportedSyntaxes)
	reBoxed     = regexp.MustCompile(`\\boxed\s*\{`)
	reDfrac     = regexp.MustCompile(`\\dfrac`)
	reTfrac     = regexp.MustCompile(`\\tfrac`)
	rePrime     = regexp.MustCompile(`'`)
	reOverright = regexp.MustCompile(`\\overrightarrow`)
	reImplies   = regexp.MustCompile(`\\implies`)
	reHarpoons  = regexp.MustCompile(`\\rightleftharpoons`)
	reDots      = regexp.MustCompile(`\\dots`)
	reBracketSz = regexp.MustCompile(`\\(?:big|Big|bigg|Bigg)[lr]?`)
)

func preprocessLaTeX(src string) string {
	src = reDollarBlockMath.ReplaceAllStringFunc(src, func(m string) string {
		inner := reDollarBlockMath.FindStringSubmatch(m)
		if len(inner) < 2 {
			return m
		}
		code := filterLatexSyntax(inner[1])
		return "```blockmath\n" + strings.TrimSpace(code) + "\n```"
	})
	src = reSlashBracket.ReplaceAllStringFunc(src, func(m string) string {
		inner := reSlashBracket.FindStringSubmatch(m)
		if len(inner) < 2 {
			return m
		}
		code := filterLatexSyntax(inner[1])
		return "```blockmath\n" + strings.TrimSpace(code) + "\n```"
	})
	src = reInlineParen.ReplaceAllStringFunc(src, func(m string) string {
		code := filterLatexSyntax(m)
		return "`" + code + "`"
	})
	src = reDollarInline.ReplaceAllStringFunc(src, func(m string) string {
		inner := reDollarInline.FindStringSubmatch(m)
		if len(inner) < 2 {
			return m
		}
		code := filterLatexSyntax(inner[1])
		return "`$" + code + "$`"
	})
	return src
}

func filterLatexSyntax(s string) string {
	s = reBoxed.ReplaceAllString(s, "{")
	s = reDfrac.ReplaceAllString(s, `\frac`)
	s = reTfrac.ReplaceAllString(s, `\frac`)
	s = rePrime.ReplaceAllString(s, `\prime`)
	s = reOverright.ReplaceAllString(s, `\vec`)
	s = reImplies.ReplaceAllString(s, `\Rightarrow`)
	s = reHarpoons.ReplaceAllString(s, `\Leftrightarrow`)
	s = reDots.ReplaceAllString(s, `\ldots`)
	s = reBracketSz.ReplaceAllString(s, "")
	return s
}

// ── Speculative Rewriting (mirrors Swift PartialEmphasisRewriter) ────────────

var (
	rePartialStrong = regexp.MustCompile(`(?:\*\*|__)\S*$`)
	rePartialItalic = regexp.MustCompile(`(?:\*|_)\S*$`)
)

// rewriteSpeculative applies both emphasis and table speculative rewriting.
func rewriteSpeculative(doc *Node) {
	rewriteSpeculativeEmphasis(doc)
	rewriteSpeculativeTable(doc)
}

// rewriteSpeculativeEmphasis closes partial emphasis at the end of streamed text.
func rewriteSpeculativeEmphasis(doc *Node) {
	textNode := doc.RightmostDescendant()
	if textNode == nil || textNode.Type != NodeText {
		return
	}
	text := textNode.Text
	if text == "" {
		return
	}

	parent := textNode.Parent
	if parent == nil {
		return
	}

	// Check partial strong: **text or __text at end
	if loc := rePartialStrong.FindStringIndex(text); loc != nil && loc[1] == len(text) {
		delimLen := 2
		inner := text[loc[0]+delimLen:]
		prefix := text[:loc[0]]
		idx := textNode.IndexInParent()
		if idx < 0 {
			return
		}
		var newNodes []*Node
		if prefix != "" {
			pn := NewNode(NodeText)
			pn.Text = prefix
			newNodes = append(newNodes, pn)
		}
		sn := NewNode(NodeStrong)
		in := NewNode(NodeText)
		in.Text = inner
		sn.Append(in)
		newNodes = append(newNodes, sn)
		parent.Children = append(parent.Children[:idx], append(newNodes, parent.Children[idx+1:]...)...)
		for _, c := range newNodes {
			c.Parent = parent
		}
		return
	}

	// Check partial italic: *text or _text at end
	if loc := rePartialItalic.FindStringIndex(text); loc != nil && loc[1] == len(text) {
		delimLen := 1
		inner := text[loc[0]+delimLen:]
		prefix := text[:loc[0]]
		idx := textNode.IndexInParent()
		if idx < 0 {
			return
		}
		var newNodes []*Node
		if prefix != "" {
			pn := NewNode(NodeText)
			pn.Text = prefix
			newNodes = append(newNodes, pn)
		}
		en := NewNode(NodeEmphasis)
		in := NewNode(NodeText)
		in.Text = inner
		en.Append(in)
		newNodes = append(newNodes, en)
		parent.Children = append(parent.Children[:idx], append(newNodes, parent.Children[idx+1:]...)...)
		for _, c := range newNodes {
			c.Parent = parent
		}
	}
}

// rewriteSpeculativeTable detects half-formed tables in paragraphs and removes them.
// Mirrors Swift PartialTableRewriter + PartialTableScanner.
func rewriteSpeculativeTable(doc *Node) {
	// Find the rightmost paragraph
	para := findRightmostParagraph(doc)
	if para == nil {
		return
	}
	// Check if paragraph content looks like a partial table
	text := para.TextContent()
	if isPartialTable(text) {
		// Empty the paragraph's children so nothing renders
		para.Children = nil
	}
}

func findRightmostParagraph(n *Node) *Node {
	leaf := n.RightmostDescendant()
	// Walk up to find the enclosing Paragraph
	p := leaf
	for p != nil {
		if p.Type == NodeParagraph {
			return p
		}
		p = p.Parent
	}
	return nil
}

func isPartialTable(text string) bool {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 0 {
		return false
	}
	// Single line starting with | could be a table header in progress
	if len(lines) == 1 {
		trimmed := strings.TrimSpace(lines[0])
		return len(trimmed) > 1 && trimmed[0] == '|' && !strings.Contains(trimmed, "\n")
	}
	// Two lines: header + partial separator
	if len(lines) == 2 {
		first := strings.TrimSpace(lines[0])
		second := strings.TrimSpace(lines[1])
		if len(first) > 0 && first[0] == '|' {
			// Second line should look like a separator in progress
			if len(second) > 0 && (second[0] == '|' || second[0] == '-' || second[0] == ':') {
				return true
			}
		}
	}
	return false
}

// ── Public API ──────────────────────────────────────────────────────────────

// ParseOption controls parsing behavior.
//
// Example:
//
//	opt := parser.DefaultOption()
//	fmt.Println(opt.PreprocessLaTeX) // true
type ParseOption struct {
	SpeculativeRewrite bool
	PreprocessLaTeX    bool
}

// DefaultOption returns sensible defaults.
//
// Example:
//
//	opt := parser.DefaultOption()
//	fmt.Println(opt.SpeculativeRewrite) // false
func DefaultOption() ParseOption {
	return ParseOption{
		SpeculativeRewrite: false,
		PreprocessLaTeX:    true,
	}
}

// StreamOption returns options suitable for streaming/incremental rendering.
//
// Example:
//
//	opt := parser.StreamOption()
//	fmt.Println(opt.SpeculativeRewrite) // true
func StreamOption() ParseOption {
	return ParseOption{
		SpeculativeRewrite: true,
		PreprocessLaTeX:    true,
	}
}

// Parse parses markdown text into an AST.
//
// Example:
//
//	doc := parser.Parse("# Hello world", parser.DefaultOption())
//	fmt.Println(doc.Children[0].Type) // prints 1 (NodeHeading)
func Parse(src string, opt ParseOption) *Node {
	if opt.PreprocessLaTeX {
		src = preprocessLaTeX(src)
	}
	lines := splitLines(src)
	doc := parseDocument(lines)
	ComputeIDs(doc)
	if opt.SpeculativeRewrite {
		rewriteSpeculative(doc)
	}
	return doc
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// ── Block Parser ────────────────────────────────────────────────────────────

type blockParser struct {
	lines []string
	pos   int
}

func parseDocument(lines []string) *Node {
	bp := &blockParser{lines: lines, pos: 0}
	doc := NewNode(NodeDocument)
	bp.parseBlocks(doc)
	return doc
}

func (bp *blockParser) parseBlocks(parent *Node) {
	for bp.pos < len(bp.lines) {
		bp.parseBlock(parent)
	}
}

func (bp *blockParser) parseBlock(parent *Node) {
	if bp.pos >= len(bp.lines) {
		return
	}
	line := bp.lines[bp.pos]
	trimmed := strings.TrimLeft(line, " ")

	// Blank line
	if trimmed == "" {
		bp.pos++
		return
	}
	// Thematic break
	if isThematicBreak(trimmed) {
		parent.Append(NewNode(NodeThematicBreak))
		bp.pos++
		return
	}
	// ATX Heading
	if heading, consumed := bp.tryParseHeading(); heading != nil {
		bp.pos += consumed
		parent.Append(heading)
		return
	}
	// Fenced code block
	if isFenceStart(trimmed) {
		bp.parseFencedCodeBlock(parent, trimmed[0])
		return
	}
	// Blockquote
	if trimmed[0] == '>' {
		bp.parseBlockquote(parent, 0)
		return
	}
	// Unordered list
	if isUnorderedListStart(trimmed) {
		bp.parseUnorderedList(parent)
		return
	}
	// Ordered list
	if isOrderedListStart(trimmed) {
		bp.parseOrderedList(parent)
		return
	}
	// Table: line starts with | and next line is separator
	if trimmed[0] == '|' && bp.pos+1 < len(bp.lines) && isTableSeparator(bp.lines[bp.pos+1]) {
		bp.parseTable(parent)
		return
	}
	// Indented code block (4 spaces or 1 tab)
	if isIndentedCodeStart(line) {
		bp.parseIndentedCodeBlock(parent)
		return
	}
	// Paragraph
	bp.parseParagraph(parent)
}

// ── Heading ─────────────────────────────────────────────────────────────────

func (bp *blockParser) tryParseHeading() (*Node, int) {
	line := bp.lines[bp.pos]
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	if i == 0 || i > 6 {
		return nil, 0
	}
	if i < len(line) && line[i] != ' ' {
		return nil, 0
	}
	j := i
	for j < len(line) && line[j] == ' ' {
		j++
	}
	text := strings.TrimSpace(line[j:])
	text = strings.TrimRight(text, "# ")
	text = strings.TrimRight(text, " ")

	node := NewNode(NodeHeading)
	node.Level = i
	parseInline(node, text)
	return node, 1
}

// ── Thematic Break ──────────────────────────────────────────────────────────

func isThematicBreak(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	ch := trimmed[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	count := 0
	for _, r := range trimmed {
		if byte(r) == ch {
			count++
		} else if r != ' ' && r != '\t' {
			return false
		}
	}
	return count >= 3
}

// ── Fenced Code Block ───────────────────────────────────────────────────────

func isFenceStart(line string) bool {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) < 3 {
		return false
	}
	ch := trimmed[0]
	if ch != '`' && ch != '~' {
		return false
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == ch {
		count++
	}
	return count >= 3
}

func (bp *blockParser) parseFencedCodeBlock(parent *Node, fenceChar byte) {
	firstLine := bp.lines[bp.pos]
	trimmed := strings.TrimLeft(firstLine, " ")
	fenceLen := 0
	for fenceLen < len(trimmed) && trimmed[fenceLen] == fenceChar {
		fenceLen++
	}
	info := strings.TrimSpace(trimmed[fenceLen:])
	lang := ""
	if idx := strings.IndexAny(info, " \t"); idx >= 0 {
		lang = info[:idx]
	} else {
		lang = info
	}
	bp.pos++

	var codeLines []string
	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		trimmedLine := strings.TrimLeft(line, " ")
		if len(trimmedLine) >= fenceLen {
			allFence := true
			for k := 0; k < fenceLen; k++ {
				if trimmedLine[k] != fenceChar {
					allFence = false
					break
				}
			}
			if allFence && strings.TrimSpace(trimmedLine[fenceLen:]) == "" {
				bp.pos++
				goto done
			}
		}
		codeLines = append(codeLines, line)
		bp.pos++
	}
done:
	node := NewNode(NodeFencedCodeBlock)
	node.Language = lang
	node.Code = strings.Join(codeLines, "\n")
	parent.Append(node)
}

// ── Indented Code Block (mirrors CommonMark 4-space rule) ───────────────────

func isIndentedCodeStart(line string) bool {
	if len(line) == 0 {
		return false
	}
	// 4 spaces or 1 tab
	if len(line) >= 4 && line[:4] == "    " {
		return true
	}
	if line[0] == '\t' {
		return true
	}
	return false
}

func (bp *blockParser) parseIndentedCodeBlock(parent *Node) {
	var codeLines []string
	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		if strings.TrimSpace(line) == "" {
			// Blank line: check if next non-blank line is still indented
			peek := bp.pos + 1
			for peek < len(bp.lines) && strings.TrimSpace(bp.lines[peek]) == "" {
				peek++
			}
			if peek < len(bp.lines) && isIndentedCodeStart(bp.lines[peek]) {
				codeLines = append(codeLines, "")
				bp.pos++
				continue
			}
			break
		}
		if !isIndentedCodeStart(line) {
			break
		}
		// Strip leading 4 spaces or 1 tab
		if len(line) >= 4 && line[:4] == "    " {
			codeLines = append(codeLines, line[4:])
		} else if line[0] == '\t' {
			codeLines = append(codeLines, line[1:])
		} else {
			codeLines = append(codeLines, line)
		}
		bp.pos++
	}
	if len(codeLines) == 0 {
		return
	}
	node := NewNode(NodeIndentedCodeBlock)
	node.Code = strings.Join(codeLines, "\n")
	parent.Append(node)
}

// ── Blockquote (recursive nesting support) ──────────────────────────────────

func (bp *blockParser) parseBlockquote(parent *Node, depth int) {
	node := NewNode(NodeBlockquote)
	node.QuoteLevel = depth

	var innerLines []string
	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" {
			if bp.pos+1 < len(bp.lines) {
				nextTrim := strings.TrimLeft(bp.lines[bp.pos+1], " ")
				if len(nextTrim) > 0 && nextTrim[0] == '>' {
					bp.pos++
					innerLines = append(innerLines, "")
					continue
				}
			}
			break
		}
		if trimmed[0] != '>' {
			break
		}
		content := trimmed[1:]
		if len(content) > 0 && content[0] == ' ' {
			content = content[1:]
		}
		innerLines = append(innerLines, content)
		bp.pos++
	}

	// Recursively parse inner content
	subParser := &blockParser{lines: innerLines, pos: 0}
	subParser.parseBlocks(node)
	parent.Append(node)
}

// ── Unordered List ──────────────────────────────────────────────────────────

func isUnorderedListStart(line string) bool {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) < 2 {
		return false
	}
	ch := trimmed[0]
	if ch != '-' && ch != '*' && ch != '+' {
		return false
	}
	return len(trimmed) > 1 && (trimmed[1] == ' ' || trimmed[1] == '\t')
}

func (bp *blockParser) parseUnorderedList(parent *Node) {
	node := NewNode(NodeUnorderedList)
	node.Ordered = false

	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" {
			if bp.pos+1 < len(bp.lines) && isUnorderedListStart(bp.lines[bp.pos+1]) {
				bp.pos++
				continue
			}
			break
		}
		if !isUnorderedListStart(trimmed) {
			if len(node.Children) > 0 {
				lastItem := node.Children[len(node.Children)-1]
				contText := strings.TrimLeft(line, " ")
				contNode := NewNode(NodeText)
				contNode.Text = "\n" + contText
				lastItem.Append(contNode)
				bp.pos++
				continue
			}
			break
		}
		markerLen := 1
		content := trimmed[markerLen:]
		if len(content) > 0 && content[0] == ' ' {
			content = content[1:]
		}
		listItem := NewNode(NodeListItem)

		// Task list checkbox — strip BEFORE parseInline so '[' isn't misinterpreted as link
		if strings.HasPrefix(content, "[ ] ") {
			listItem.Checked = false
			listItem.IsTaskItem = true
			content = content[4:]
		} else if strings.HasPrefix(content, "[x] ") || strings.HasPrefix(content, "[X] ") {
			listItem.Checked = true
			listItem.IsTaskItem = true
			content = content[4:]
		}

		para := NewNode(NodeParagraph)
		parseInline(para, content)
		listItem.Append(para)

		// Detect startsWithBold (mirrors Swift ListItem+.swift)
		if len(para.Children) > 0 && para.Children[0].Type == NodeStrong {
			listItem.StartsWithBold = true
		}

		node.Append(listItem)
		bp.pos++
	}
	parent.Append(node)
}

// ── Ordered List ────────────────────────────────────────────────────────────

func isOrderedListStart(line string) bool {
	trimmed := strings.TrimLeft(line, " ")
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	if i == 0 || i > 9 || i >= len(trimmed) {
		return false
	}
	return (trimmed[i] == '.' || trimmed[i] == ')') && i+1 < len(trimmed) && trimmed[i+1] == ' '
}

func parseOrderedListStart(line string) (int, int, string) {
	trimmed := strings.TrimLeft(line, " ")
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	num := 0
	for j := 0; j < i; j++ {
		num = num*10 + int(trimmed[j]-'0')
	}
	delimLen := 2
	content := trimmed[i+delimLen:]
	return num, i + delimLen, content
}

func (bp *blockParser) parseOrderedList(parent *Node) {
	node := NewNode(NodeOrderedList)
	node.Ordered = true
	first := true

	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" {
			if bp.pos+1 < len(bp.lines) && isOrderedListStart(bp.lines[bp.pos+1]) {
				bp.pos++
				continue
			}
			break
		}
		if !isOrderedListStart(trimmed) {
			if len(node.Children) > 0 {
				lastItem := node.Children[len(node.Children)-1]
				contText := strings.TrimLeft(line, " ")
				contNode := NewNode(NodeText)
				contNode.Text = "\n" + contText
				lastItem.Append(contNode)
				bp.pos++
				continue
			}
			break
		}
		num, _, content := parseOrderedListStart(trimmed)
		if first {
			node.StartNum = num
			first = false
		}
		listItem := NewNode(NodeListItem)

		// Task list — strip BEFORE parseInline
		if strings.HasPrefix(content, "[ ] ") {
			listItem.Checked = false
			listItem.IsTaskItem = true
			content = content[4:]
		} else if strings.HasPrefix(content, "[x] ") || strings.HasPrefix(content, "[X] ") {
			listItem.Checked = true
			listItem.IsTaskItem = true
			content = content[4:]
		}

		para := NewNode(NodeParagraph)
		parseInline(para, content)
		listItem.Append(para)

		// Detect startsWithBold
		if len(para.Children) > 0 && para.Children[0].Type == NodeStrong {
			listItem.StartsWithBold = true
		}

		node.Append(listItem)
		bp.pos++
	}
	parent.Append(node)
}

// ── Table ───────────────────────────────────────────────────────────────────

func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	hasDash := false
	hasPipe := false
	for _, r := range trimmed {
		if r == '-' {
			hasDash = true
		} else if r == '|' {
			hasPipe = true
		} else if r != ' ' && r != ':' {
			return false
		}
	}
	return hasDash && hasPipe
}

func parseTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) > 0 && trimmed[0] == '|' {
		trimmed = trimmed[1:]
	}
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '|' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	parts := strings.Split(trimmed, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

func (bp *blockParser) parseTable(parent *Node) {
	node := NewNode(NodeTable)

	headerCells := parseTableRow(bp.lines[bp.pos])
	bp.pos++

	if bp.pos < len(bp.lines) && isTableSeparator(bp.lines[bp.pos]) {
		bp.pos++
	}

	var rows [][]string
	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.Contains(trimmed, "|") {
			break
		}
		cells := parseTableRow(line)
		for len(cells) < len(headerCells) {
			cells = append(cells, "")
		}
		if len(cells) > len(headerCells) {
			cells = cells[:len(headerCells)]
		}
		rows = append(rows, cells)
		bp.pos++
	}

	// Header row
	headerNode := NewNode(NodeTableRow)
	headerNode.IsTableHeader = true
	for _, cell := range headerCells {
		cellNode := NewNode(NodeTableCell)
		cellNode.IsTableHeader = true
		parseInline(cellNode, cell)
		headerNode.Append(cellNode)
	}
	node.Append(headerNode)

	// Body rows
	for _, row := range rows {
		rowNode := NewNode(NodeTableRow)
		for _, cell := range row {
			cellNode := NewNode(NodeTableCell)
			parseInline(cellNode, cell)
			rowNode.Append(cellNode)
		}
		node.Append(rowNode)
	}
	parent.Append(node)
}

// ── Paragraph ───────────────────────────────────────────────────────────────

func (bp *blockParser) parseParagraph(parent *Node) {
	var lines []string
	for bp.pos < len(bp.lines) {
		line := bp.lines[bp.pos]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		if isThematicBreak(trimmed) {
			break
		}
		if isAtxHeading(trimmed) {
			break
		}
		if isFenceStart(trimmed) {
			break
		}
		if trimmed[0] == '>' {
			break
		}
		if isUnorderedListStart(trimmed) {
			break
		}
		if isOrderedListStart(trimmed) {
			break
		}
		if isIndentedCodeStart(line) {
			break
		}
		if trimmed[0] == '|' && bp.pos+1 < len(bp.lines) && isTableSeparator(bp.lines[bp.pos+1]) {
			break
		}
		lines = append(lines, line)
		bp.pos++
	}
	if len(lines) == 0 {
		return
	}
	node := NewNode(NodeParagraph)
	text := strings.Join(lines, "\n")
	parseInline(node, text)
	parent.Append(node)
}

func isAtxHeading(line string) bool {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	return i >= 1 && i <= 6 && i < len(line) && line[i] == ' '
}

// ── Inline Parser ───────────────────────────────────────────────────────────

type inlineParser struct {
	text string
	pos  int
}

func parseInline(parent *Node, text string) {
	ip := &inlineParser{text: text, pos: 0}
	ip.parse(parent)
}

func (ip *inlineParser) parse(parent *Node) {
	for ip.pos < len(ip.text) {
		ch := ip.text[ip.pos]
		switch ch {
		case '`':
			ip.parseCodeSpan(parent)
		case '*':
			ip.parseEmphasis(parent, '*')
		case '_':
			ip.parseEmphasis(parent, '_')
		case '~':
			if ip.pos+1 < len(ip.text) && ip.text[ip.pos+1] == '~' {
				ip.parseStrikethrough(parent)
			} else {
				ip.emitText(parent, "~")
			}
		case '[':
			ip.parseLinkOrImage(parent)
		case '!':
			if ip.pos+1 < len(ip.text) && ip.text[ip.pos+1] == '[' {
				ip.parseImage(parent)
			} else {
				ip.emitText(parent, "!")
			}
		case '<':
			if !ip.tryParseAutolink(parent) {
				ip.emitText(parent, "<")
			}
		case '\\':
			if ip.pos+1 < len(ip.text) {
				next := ip.text[ip.pos+1]
				if next == '\n' {
					parent.Append(NewNode(NodeHardBreak))
					ip.pos += 2
				} else {
					ip.emitText(parent, string(next))
					ip.pos += 2
				}
			} else {
				ip.emitText(parent, "\\")
			}
		case '\n':
			if ip.pos >= 2 && ip.text[ip.pos-2] == ' ' && ip.text[ip.pos-1] == ' ' {
				removeTrailingSpaces(parent)
				parent.Append(NewNode(NodeHardBreak))
				ip.pos++
			} else {
				parent.Append(NewNode(NodeSoftBreak))
				ip.pos++
			}
		default:
			start := ip.pos
			for ip.pos < len(ip.text) {
				c := ip.text[ip.pos]
				if c == '`' || c == '*' || c == '_' || c == '~' || c == '[' || c == '!' || c == '<' || c == '\\' || c == '\n' {
					break
				}
				ip.pos++
			}
			if ip.pos > start {
				node := NewNode(NodeText)
				node.Text = ip.text[start:ip.pos]
				parent.Append(node)
			}
		}
	}
}

func (ip *inlineParser) emitText(parent *Node, text string) {
	node := NewNode(NodeText)
	node.Text = text
	parent.Append(node)
	ip.pos += len(text)
}

// ── Code Span ───────────────────────────────────────────────────────────────

func (ip *inlineParser) parseCodeSpan(parent *Node) {
	start := ip.pos
	backtickCount := 0
	for ip.pos < len(ip.text) && ip.text[ip.pos] == '`' {
		backtickCount++
		ip.pos++
	}
	for ip.pos < len(ip.text) {
		if ip.text[ip.pos] == '`' {
			endCount := 0
			endPos := ip.pos
			for endPos < len(ip.text) && ip.text[endPos] == '`' {
				endCount++
				endPos++
			}
			if endCount == backtickCount {
				code := ip.text[start+backtickCount : ip.pos]
				if len(code) > 0 && code[0] == ' ' && code[len(code)-1] == ' ' && len(code) > 2 {
					code = code[1 : len(code)-1]
				}
				node := NewNode(NodeCodeSpan)
				node.Text = code
				parent.Append(node)
				ip.pos = endPos
				return
			}
			ip.pos = endPos
		} else {
			ip.pos++
		}
	}
	ip.pos = start
	ip.emitText(parent, "`")
}

// ── Emphasis / Strong ───────────────────────────────────────────────────────

func (ip *inlineParser) parseEmphasis(parent *Node, delim byte) {
	start := ip.pos
	count := 0
	for ip.pos < len(ip.text) && ip.text[ip.pos] == delim {
		count++
		ip.pos++
	}
	if count > 2 {
		ip.pos = start + 1
		node := NewNode(NodeText)
		node.Text = string(delim)
		parent.Append(node)
		return
	}
	delimStr := strings.Repeat(string(delim), count)

	searchPos := ip.pos
	found := false
	for searchPos < len(ip.text) {
		if ip.text[searchPos] == delim {
			endCount := 0
			endStart := searchPos
			for searchPos < len(ip.text) && ip.text[searchPos] == delim {
				endCount++
				searchPos++
			}
			if endCount >= count {
				innerText := ip.text[ip.pos:endStart]
				if innerText != "" {
					var node *Node
					if count == 2 {
						node = NewNode(NodeStrong)
					} else {
						node = NewNode(NodeEmphasis)
					}
					parseInline(node, innerText)
					parent.Append(node)
				}
				ip.pos = searchPos
				found = true
				break
			}
		} else {
			searchPos++
		}
	}
	if !found {
		node := NewNode(NodeText)
		node.Text = delimStr
		parent.Append(node)
	}
}

// ── Strikethrough ───────────────────────────────────────────────────────────

func (ip *inlineParser) parseStrikethrough(parent *Node) {
	ip.pos += 2
	idx := strings.Index(ip.text[ip.pos:], "~~")
	if idx < 0 {
		node := NewNode(NodeText)
		node.Text = "~~"
		parent.Append(node)
		return
	}
	innerText := ip.text[ip.pos : ip.pos+idx]
	node := NewNode(NodeStrikethrough)
	parseInline(node, innerText)
	parent.Append(node)
	ip.pos += idx + 2
}

// ── Link / Image ────────────────────────────────────────────────────────────

func (ip *inlineParser) parseLinkOrImage(parent *Node) {
	innerStart := ip.pos + 1
	closeBracket := findClosingBracket(ip.text, ip.pos)
	if closeBracket < 0 {
		ip.emitText(parent, "[")
		return
	}
	if closeBracket+1 < len(ip.text) && ip.text[closeBracket+1] == '(' {
		urlStart := closeBracket + 2
		urlEnd := strings.IndexByte(ip.text[urlStart:], ')')
		if urlEnd < 0 {
			ip.emitText(parent, "[")
			return
		}
		urlEnd += urlStart
		url := ip.text[urlStart:urlEnd]
		title := ""
		if spaceIdx := strings.LastIndex(url, ` "`); spaceIdx >= 0 && strings.HasSuffix(url, `"`) {
			title = url[spaceIdx+2 : len(url)-1]
			url = url[:spaceIdx]
		}
		node := NewNode(NodeLink)
		node.URL = url
		node.Title = title
		parseInline(node, ip.text[innerStart:closeBracket])
		parent.Append(node)
		ip.pos = urlEnd + 1
		return
	}
	ip.emitText(parent, "[")
}

func (ip *inlineParser) parseImage(parent *Node) {
	innerStart := ip.pos + 2
	closeBracket := -1
	depth := 0
	for i := innerStart; i < len(ip.text); i++ {
		if ip.text[i] == '[' {
			depth++
		} else if ip.text[i] == ']' {
			if depth == 0 {
				closeBracket = i
				break
			}
			depth--
		}
	}
	if closeBracket < 0 {
		ip.emitText(parent, "!")
		return
	}
	if closeBracket+1 < len(ip.text) && ip.text[closeBracket+1] == '(' {
		urlStart := closeBracket + 2
		urlEnd := strings.IndexByte(ip.text[urlStart:], ')')
		if urlEnd < 0 {
			ip.emitText(parent, "!")
			return
		}
		urlEnd += urlStart
		url := ip.text[urlStart:urlEnd]
		title := ""
		if spaceIdx := strings.LastIndex(url, ` "`); spaceIdx >= 0 && strings.HasSuffix(url, `"`) {
			title = url[spaceIdx+2 : len(url)-1]
			url = url[:spaceIdx]
		}
		node := NewNode(NodeImage)
		node.URL = url
		node.Title = title
		altNode := NewNode(NodeText)
		altNode.Text = ip.text[innerStart:closeBracket]
		node.Append(altNode)
		parent.Append(node)
		ip.pos = urlEnd + 1
		return
	}
	ip.emitText(parent, "!")
}

func findClosingBracket(text string, openPos int) int {
	depth := 0
	for i := openPos; i < len(text); i++ {
		if text[i] == '[' {
			depth++
		} else if text[i] == ']' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// ── Autolink ────────────────────────────────────────────────────────────────

func (ip *inlineParser) tryParseAutolink(parent *Node) bool {
	if ip.text[ip.pos] != '<' {
		return false
	}
	end := strings.IndexByte(ip.text[ip.pos+1:], '>')
	if end < 0 {
		return false
	}
	end += ip.pos + 1
	inner := ip.text[ip.pos+1 : end]
	if isURL(inner) || isEmail(inner) {
		node := NewNode(NodeAutolink)
		node.URL = inner
		node.Text = inner
		parent.Append(node)
		ip.pos = end + 1
		return true
	}
	return false
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "ftp://")
}

func isEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func removeTrailingSpaces(parent *Node) {
	if len(parent.Children) == 0 {
		return
	}
	last := parent.Children[len(parent.Children)-1]
	if last.Type == NodeText {
		last.Text = strings.TrimRight(last.Text, " ")
		if last.Text == "" {
			parent.Children = parent.Children[:len(parent.Children)-1]
		}
	}
}

// isSpace returns true if r is a space or tab.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// trimLeftSpaces trims leading spaces from a string.
func trimLeftSpaces(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}
