package parser

import (
	"fmt"
)

// Example demonstrates parsing a Markdown document into an AST.
func Example() {
	doc := Parse("# Hello world", DefaultOption())
	fmt.Println(doc.Type)
	// Output: 0
}

// ExampleParse demonstrates parsing inline formatting elements.
func ExampleParse() {
	doc := Parse("- item 1\n- item 2", DefaultOption())
	fmt.Println(doc.Children[0].Type)
	// Output: 8
}

// ExampleDefaultOption returns sensible default parsing options.
func ExampleDefaultOption() {
	opt := DefaultOption()
	fmt.Println(opt.SpeculativeRewrite)
	// Output: false
}

// ExampleStreamOption returns options optimized for streaming/incremental rendering.
func ExampleStreamOption() {
	opt := StreamOption()
	fmt.Println(opt.SpeculativeRewrite)
	// Output: true
}

// ExampleNode_TextContent demonstrates extracting plain text from an AST node.
func ExampleNode_TextContent() {
	doc := Parse("Hello **world**", DefaultOption())
	fmt.Print(doc.TextContent())
	// Output: Hello world
}

// ExampleNode_Walk demonstrates depth-first traversal of the AST.
func ExampleNode_Walk() {
	doc := Parse("# Title", DefaultOption())
	count := 0
	doc.Walk(func(n *Node) bool {
		count++
		return true
	})
	fmt.Println(count)
	// Output: 3
}

// ExampleComputeIDs demonstrates assigning stable IDs to AST nodes.
func ExampleComputeIDs() {
	doc := Parse("# Title\n\nParagraph.", DefaultOption())
	ComputeIDs(doc)
	fmt.Println(doc.ID)
	// Output: 0
}

// ExampleNewNode creates a new AST node of the given type.
func ExampleNewNode() {
	node := NewNode(NodeHeading)
	fmt.Println(node.Type)
	// Output: 1
}

// ExampleNode_IsBlock demonstrates checking whether a node is block-level.
func ExampleNode_IsBlock() {
	doc := Parse("text", DefaultOption())
	fmt.Println(doc.IsBlock())
	// Output: true
}

// ExampleNode_IsInline demonstrates checking whether a node is inline-level.
func ExampleNode_IsInline() {
	doc := Parse("text", DefaultOption())
	fmt.Println(doc.IsInline())
	// Output: false
}

// ExampleNode_FindChild demonstrates finding the first child of a given type.
func ExampleNode_FindChild() {
	doc := Parse("# Hello", DefaultOption())
	fmt.Println(doc.FindChild(NodeHeading) != nil)
	// Output: true
}

// ExampleNode_IndexInParent demonstrates finding a node's index in its parent.
func ExampleNode_IndexInParent() {
	doc := Parse("# Title\n\nPara.", DefaultOption())
	para := doc.FindChild(NodeParagraph)
	if para != nil {
		fmt.Println(para.IndexInParent())
	}
	// Output: 1
}

// ExampleNode_RightmostDescendant demonstrates finding the deepest leaf node.
func ExampleNode_RightmostDescendant() {
	doc := Parse("# Title", DefaultOption())
	leaf := doc.RightmostDescendant()
	fmt.Println(leaf.Type)
	// Output: 100
}

// ExampleNode_Append demonstrates adding a child node.
func ExampleNode_Append() {
	parent := NewNode(NodeParagraph)
	child := NewNode(NodeText)
	child.Text = "hello"
	parent.Append(child)
	fmt.Println(len(parent.Children))
	// Output: 1
}
