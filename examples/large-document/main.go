// Example: Large Document - 处理大型文档示例
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/startvibecoding/GoStreamingMarkdown/gsm"
)

func main() {
	fmt.Println("=== 大型文档渲染演示 ===")
	fmt.Println()

	// 生成一个大型 Markdown 文档
	markdown := generateLargeDocument()

	// 测试渲染性能
	fmt.Printf("文档大小: %d 字符\n", len(markdown))
	fmt.Println()

	// 一次性渲染
	fmt.Println("--- 一次性渲染 ---")
	start := time.Now()
	output := gsm.Render(markdown, 80, nil)
	elapsed := time.Since(start)
	fmt.Printf("渲染耗时: %v\n", elapsed)
	fmt.Printf("输出大小: %d 字符\n", len(output))
	fmt.Println()

	// 流式渲染
	fmt.Println("--- 流式渲染 ---")
	stream := gsm.NewStream(80, nil)

	// 模拟分块接收
	chunks := splitIntoChunks(markdown, 500)
	fmt.Printf("分为 %d 个块\n", len(chunks))

	start = time.Now()
	for i, chunk := range chunks {
		stream.Update(chunk)
		_ = stream.Output() // 在实际应用中会显示输出

		if (i+1)%10 == 0 {
			fmt.Printf("  已处理 %d/%d 块\n", i+1, len(chunks))
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("流式渲染耗时: %v\n", elapsed)
	fmt.Println()

	// 显示部分内容
	fmt.Println("--- 文档预览 (前 2000 字符) ---")
	fmt.Println(gsm.Render(markdown[:2000], 60, nil))
	fmt.Println("...")
}

// generateLargeDocument 生成一个大型 Markdown 文档
func generateLargeDocument() string {
	var sb strings.Builder

	sb.WriteString("# 大型文档测试\n\n")
	sb.WriteString("这是一个用于测试大型文档渲染性能的示例。\n\n")

	// 添加多个章节
	for i := 1; i <= 20; i++ {
		sb.WriteString(fmt.Sprintf("## 第 %d 章\n\n", i))

		// 添加段落
		for j := 1; j <= 5; j++ {
			sb.WriteString(fmt.Sprintf("这是第 %d 章的第 %d 个段落。", i, j))
			sb.WriteString("这里包含一些 **粗体** 和 *斜体* 文本，")
			sb.WriteString("以及 `inline code` 示例。\n\n")
		}

		// 添加代码块
		if i%3 == 0 {
			sb.WriteString("```go\n")
			sb.WriteString(fmt.Sprintf("// Chapter %d example\n", i))
			sb.WriteString("func example() {\n")
			sb.WriteString("    fmt.Println(\"Hello\")\n")
			sb.WriteString("}\n")
			sb.WriteString("```\n\n")
		}

		// 添加列表
		if i%2 == 0 {
			sb.WriteString("- 列表项 1\n")
			sb.WriteString("- 列表项 2\n")
			sb.WriteString("- 列表项 3\n\n")
		}

		// 添加表格
		if i%5 == 0 {
			sb.WriteString("| 列 A | 列 B | 列 C |\n")
			sb.WriteString("|------|------|------|\n")
			sb.WriteString("| 数据 1 | 数据 2 | 数据 3 |\n")
			sb.WriteString("| 数据 4 | 数据 5 | 数据 6 |\n\n")
		}

		// 添加引用
		if i%4 == 0 {
			sb.WriteString("> 这是一段引用文本。\n")
			sb.WriteString("> 支持多行。\n\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// splitIntoChunks 将字符串分成指定大小的块
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for len(s) > chunkSize {
		// 找一个合适的分割点（在换行符处）
		splitAt := chunkSize
		for splitAt > 0 && s[splitAt] != '\n' {
			splitAt--
		}
		if splitAt == 0 {
			splitAt = chunkSize
		}
		chunks = append(chunks, s[:splitAt])
		s = s[splitAt:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}
