// Example: Streaming Chat - 模拟聊天应用的流式渲染
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/startvibecoding/GoStreamingMarkdown/gsm"
)

func main() {
	fmt.Println("=== 流式聊天渲染演示 ===")
	fmt.Println("(模拟 AI 聊天机器人的 Markdown 流式输出)")
	fmt.Println()

	// 创建流式渲染器
	stream := gsm.NewStream(60, nil)

	// 模拟聊天消息流
	messages := []struct {
		delay   time.Duration
		content string
	}{
		{100 * time.Millisecond, "# "},
		{100 * time.Millisecond, "# 代码"},
		{100 * time.Millisecond, "# 代码示例"},
		{200 * time.Millisecond, "# 代码示例\n\n"},
		{100 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n"},
		{200 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\n"},
		{100 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n"},
		{100 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n"},
		{150 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n"},
		{100 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n"},
		{200 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n}\n```\n"},
		{300 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n}\n```\n\n运行结果：\n"},
		{200 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n}\n```\n\n运行结果：\n\n`Hello!`\n"},
		{200 * time.Millisecond, "# 代码示例\n\n让我为你展示一个简单的 Go 程序：\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n}\n```\n\n运行结果：\n\n`Hello!`\n\n---\n\n**提示**: 这个示例展示了流式渲染效果！"},
	}

	// 清屏函数
	clearScreen := func() {
		fmt.Print("\033[H\033[2J")
	}

	// 渲染流式内容
	for _, msg := range messages {
		stream.Update(msg.content)
		clearScreen()

		// 显示输出
		output := stream.Output()
		fmt.Println(output)

		// 显示光标动画
		fmt.Print("\033[5m▌\033[0m") // 闪烁光标
		time.Sleep(msg.delay)
	}

	// 最终输出（无光标）
	clearScreen()
	fmt.Println(stream.Output())

	// 分隔线
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("演示完成！")
	fmt.Println("\n在实际应用中，你可以：")
	fmt.Println("1. 从 WebSocket 接收 AI 响应")
	fmt.Println("2. 实时更新 stream.Update()")
	fmt.Println("3. 清屏并重绘 stream.Output()")
}
