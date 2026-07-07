// Package parser implements a zero-dependency CommonMark-compatible Markdown
// parser that produces an AST. The design mirrors the Swift project's pipeline:
//
//	Preprocess (LaTeX) → Parse → Rewrite (speculative emphasis/table) → Render
package parser

import "strings"

// ── AST Node Types ──────────────────────────────────────────────────────────

// NodeType identifies the kind of AST node.
//
// Example:
//
//	doc := parser.Parse("text", parser.DefaultOption())
//	head := doc.FindChild(parser.NodeHeading)
//	fmt.Println(heading != nil) // false
type NodeType int

const (
	// Block nodes
	NodeDocument          NodeType = iota
	NodeHeading                    // # Heading
	NodeParagraph                  // plain text block
	NodeFencedCodeBlock            // ```lang ... ```
	NodeIndentedCodeBlock          // 4-space indented code
	NodeBlockquote                 // > quoted text
	NodeThematicBreak              // --- or *** or ___
	NodeOrderedList                // 1. item
	NodeUnorderedList              // - item or * item
	NodeListItem                   // single list item
	NodeTable                      // | a | b |
	NodeTableRow                   // table row
	NodeTableCell                  // table cell
)

// Inline node types (>= 100)
const (
	NodeText          NodeType = 100 + iota
	NodeEmphasis               // *text* or _text_
	NodeStrong                 // **text** or __text__
	NodeCodeSpan               // `code`
	NodeLink                   // [text](url)
	NodeImage                  // ![alt](url)
	NodeStrikethrough          // ~~text~~
	NodeSoftBreak              // newline inside paragraph
	NodeHardBreak              // trailing \\ or two trailing spaces + newline
	NodeAutolink               // <url>
	NodeLineBreak              // literal newline
)

// Node is a single node in the Markdown AST.
//
// Example:
//
//	node := parser.NewNode(parser.NodeParagraph)
//	fmt.Println(node.Type) // 1
type Node struct {
	Type     NodeType
	Children []*Node
	Parent   *Node

	// Identity (mirrors Swift Markup+ID.swift)
	ID string // stable ID computed from root path (e.g. "2-1")

	// Block-level fields
	Level    int    // Heading level 1-6
	Language string // fenced code block language
	Code     string // raw code content (for code blocks)
	Ordered  bool   // list is ordered
	StartNum int    // start number for ordered lists (mirrors swift-markdown)

	// Inline-level fields
	Text  string // literal text content
	URL   string // link/image destination
	Title string // link/image title

	// Table fields
	IsTableHeader bool // true for header cells/rows

	// List fields
	Checked        bool // task list checkbox state
	IsTaskItem     bool // this is a task list item (has [x] or [ ] checkbox)
	StartsWithBold bool // first child paragraph starts with Strong (mirrors Swift ListItem+.swift)

	// Blockquote fields
	QuoteLevel int // nesting depth for blockquotes (0 = top-level)
}

// ── Constructors ────────────────────────────────────────────────────────────

// NewNode creates a new AST node of the given type.
//
// Example:
//
//	node := parser.NewNode(parser.NodeHeading)
//	fmt.Println(node.Type) // 1
func NewNode(t NodeType) *Node {
	return &Node{Type: t}
}

// Append adds a child node.
//
// Example:
//
//	parent := parser.NewNode(parser.NodeParagraph)
//	child := parser.NewNode(parser.NodeText)
//	child.Text = "hello"
//	parent.Append(child)
//	fmt.Println(len(parent.Children)) // 1
func (n *Node) Append(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// Walk traverses the AST depth-first, calling fn for each node.
// Return false from fn to stop traversal.
//
// Example:
//
//	doc := parser.Parse("# Title", parser.DefaultOption())
//	count := 0
//	doc.Walk(func(n *parser.Node) bool {
//		count++
//		return true
//	})
//	fmt.Println(count) // 3
func (n *Node) Walk(fn func(*Node) bool) {
	if !fn(n) {
		return
	}
	for _, c := range n.Children {
		c.Walk(fn)
	}
}

// ── Stable ID Computation (mirrors Swift Markup+ID.swift) ───────────────────

// ComputeIDs assigns stable IDs to all nodes by computing the path from root.
// Each node's ID is the concatenation of its index-in-parent chain, e.g. "2-1".
//
// Example:
//
//	doc := parser.Parse("# Title\n\nPara.", parser.DefaultOption())
//	parser.ComputeIDs(doc)
//	fmt.Println(doc.ID) // "0"
func ComputeIDs(root *Node) {
	root.ID = "0"
	computeIDsRec(root)
}

func computeIDsRec(n *Node) {
	for i, child := range n.Children {
		child.ID = n.ID + "-" + itoa(i)
		computeIDsRec(child)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

// ── Plain Text Extraction (mirrors Swift String+.swift extractPlainText) ────

// TextContent recursively collects all text content from a node's children.
//
// Example:
//
//	doc := parser.Parse("Hello **world**", parser.DefaultOption())
//	fmt.Println(doc.TextContent()) // "Hello world\n"
func (n *Node) TextContent() string {
	switch n.Type {
	case NodeText:
		return n.Text
	case NodeSoftBreak:
		return " "
	case NodeHardBreak, NodeLineBreak:
		return "\n"
	case NodeCodeSpan:
		return n.Text
	case NodeHeading:
		s := ""
		for _, c := range n.Children {
			s += c.TextContent()
		}
		return s + "\n"
	case NodeParagraph:
		s := ""
		for _, c := range n.Children {
			s += c.TextContent()
		}
		return s + "\n"
	case NodeFencedCodeBlock, NodeIndentedCodeBlock:
		return n.Code + "\n"
	case NodeThematicBreak:
		return "---\n"
	case NodeOrderedList:
		s := ""
		for i, c := range n.Children {
			num := n.StartNum + i
			s += itoa(num) + ". " + c.TextContent()
		}
		return s
	case NodeUnorderedList:
		s := ""
		for _, c := range n.Children {
			s += "• " + c.TextContent()
		}
		return s
	case NodeListItem:
		s := ""
		for _, c := range n.Children {
			s += c.TextContent()
		}
		return s
	case NodeTable:
		s := ""
		for i, row := range n.Children {
			for j, cell := range row.Children {
				s += cell.TextContent()
				if j < len(row.Children)-1 {
					s += "\t"
				}
			}
			if i < len(n.Children)-1 {
				s += "\n"
			}
		}
		return s + "\n"
	case NodeBlockquote:
		s := ""
		for _, c := range n.Children {
			lines := strings.Split(c.TextContent(), "\n")
			for _, line := range lines {
				if line != "" {
					s += "> " + line + "\n"
				}
			}
		}
		return s
	default:
		s := ""
		for _, c := range n.Children {
			s += c.TextContent()
		}
		return s
	}
}

// IsBlock returns true if the node is a block-level element.
//
// Example:
//
//	doc := parser.Parse("text", parser.DefaultOption())
//	fmt.Println(doc.IsBlock()) // true
func (n *Node) IsBlock() bool {
	return n.Type < 100
}

// IsInline returns true if the node is an inline-level element.
//
// Example:
//
//	doc := parser.Parse("text", parser.DefaultOption())
//	fmt.Println(doc.IsInline()) // false
func (n *Node) IsInline() bool {
	return n.Type >= 100
}

// FindChild returns the first child matching the given type, or nil.
//
// Example:
//
//	doc := parser.Parse("# Hello", parser.DefaultOption())
//	head := doc.FindChild(parser.NodeHeading)
//	fmt.Println(heading != nil) // true
func (n *Node) FindChild(t NodeType) *Node {
	for _, c := range n.Children {
		if c.Type == t {
			return c
		}
	}
	return nil
}

// IndexInParent returns this node's index in its parent's Children slice, or -1.
//
// Example:
//
//	doc := parser.Parse("# Title\n\nPara.", parser.DefaultOption())
//	para := doc.FindChild(parser.NodeParagraph)
//	if para != nil {
//		fmt.Println(para.IndexInParent()) // 1
//	}
func (n *Node) IndexInParent() int {
	if n.Parent == nil {
		return -1
	}
	for i, c := range n.Parent.Children {
		if c == n {
			return i
		}
	}
	return -1
}

// RightmostDescendant returns the deepest last-child leaf node.
// Mirrors Swift PartialEmphasisScanner's rightMostDescendant.
//
// Example:
//
//	doc := parser.Parse("# Title", parser.DefaultOption())
//	leaf := doc.RightmostDescendant()
//	fmt.Println(leaf.Type) // 100 (NodeText)
func (n *Node) RightmostDescendant() *Node {
	cur := n
	for len(cur.Children) > 0 {
		cur = cur.Children[len(cur.Children)-1]
	}
	return cur
}
