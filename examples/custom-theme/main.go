// Example: Custom Theme - 自定义主题示例
package main

import (
	"fmt"

	"GoStreamingMarkdown/renderer"
)

func main() {
	markdown := `# 自定义主题演示

这是一个 **粗体** 和 *斜体* 的段落。

` + "```go" + `
func main() {
    fmt.Println("Hello, Custom Theme!")
}
` + "```" + `

> 引用文本，支持 **嵌套格式**

| 功能 | 状态 |
|------|:----:|
| 标题 | ✅ |
| 代码 | ✅ |

---

- 列表项 1
- [x] 已完成
- [ ] 待办
`

	// 使用内置主题
	fmt.Println("=== 默认深色主题 ===")
	fmt.Println(renderer.Render(markdown, 70, renderer.DefaultTheme()))

	fmt.Println("\n=== 浅色主题 ===")
	fmt.Println(renderer.Render(markdown, 70, renderer.LightTheme()))

	// 自定义主题
	fmt.Println("\n=== 自定义霓虹主题 ===")
	neonTheme := &renderer.Theme{
		Heading:         "\033[1;35m", // 紫色粗体
		Heading1:        "\033[1;95m", // 亮紫色
		Heading2:        "\033[1;94m", // 亮蓝色
		Heading3:        "\033[1;96m", // 亮青色
		CodeText:        "\033[93m",   // 亮黄色
		CodeBg:          "\033[48;5;236m",
		CodeLang:        "\033[38;5;208m", // 橙色
		BlockQuote:      "\033[38;5;45m",  // 青蓝色
		BlockQuoteBar:   "\033[38;5;45m",
		Bold:            "\033[1;93m", // 亮黄色粗体
		Italic:          "\033[3;92m", // 亮绿色斜体
		Strike:          "\033[9;90m",
		Code:            "\033[96m",   // 亮青色
		CodeBgInline:    "\033[48;5;236m",
		Link:            "\033[4;94m", // 蓝色下划线
		LinkURL:         "\033[90m",   // 灰色
		ListBullet:      "\033[91m",   // 红色
		ListNumber:      "\033[91m",   // 红色
		TableBorder:     "\033[90m",
		TableHeader:     "\033[1;97m",
		TableHeaderText: "\033[1;97m",
		TableCell:       "\033[97m",
		Horizontal:      "\033[90m",
		TaskChecked:     "\033[92m", // 绿色
		TaskUnchecked:   "\033[90m",
	}
	fmt.Println(renderer.Render(markdown, 70, neonTheme))
}
