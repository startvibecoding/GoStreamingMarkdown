// Example: Markdown Showcase - 展示所有支持的 Markdown 语法
package main

import (
	"fmt"

	"GoStreamingMarkdown/gsm"
)

func main() {
	markdown := `# GoStreamingMarkdown 语法展示

## 文本格式化

这是普通文本。

**粗体文本** 和 *斜体文本* 以及 ~~删除线~~。

行内代码：` + "`fmt.Println(\"Hello\")`" + `

混合格式：**粗体中的 *斜体*** 和 ~~**删除粗体**~~

## 代码块

### Go 代码

` + "```go" + `
package main

import "fmt"

func main() {
    // 这是一个注释
    message := "Hello, World!"
    fmt.Println(message)
    
    numbers := []int{1, 2, 3, 4, 5}
    for i, n := range numbers {
        fmt.Printf("Index %d: %d\n", i, n)
    }
}
` + "```" + `

### Python 代码

` + "```python" + `
def fibonacci(n):
    """Generate Fibonacci sequence"""
    a, b = 0, 1
    for _ in range(n):
        yield a
        a, b = b, a + b

# 使用示例
for num in fibonacci(10):
    print(num)
` + "```" + `

### 无语言标注

` + "```" + `
这是一个没有语言标注的代码块。
保留原始格式。
` + "```" + `

## 列表

### 无序列表

- 第一项
- 第二项
  - 嵌套项 A
  - 嵌套项 B
- 第三项

### 有序列表

1. 步骤一
2. 步骤二
3. 步骤三

### 任务列表

- [x] 完成项目初始化
- [x] 实现解析器
- [x] 实现渲染器
- [ ] 添加更多主题
- [ ] 发布 v1.0

## 引用

> 这是一段引用文本。
>
> 支持多行引用。

> **嵌套格式** 在引用中也 *有效*。

> ` + "`代码`" + ` 也可以使用。

## 表格

| 功能 | 语法 | 示例 |
|------|------|------|
| 粗体 | ` + "`**text**`" + ` | **粗体** |
| 斜体 | ` + "`*text*`" + ` | *斜体* |
| 代码 | ` + "`` `code` ``" + ` | ` + "`code`" + ` |
| 链接 | ` + "`[text](url)`" + ` | [Go](https://go.dev) |

### 对齐表格

| 左对齐 | 居中 | 右对齐 |
|:-------|:----:|-------:|
| Left | Center | Right |
| a | b | c |

## 链接和图片

访问 [Go 官网](https://go.dev) 了解更多。

图片：![Go Logo](https://go.dev/images/go-logo.png)

自动链接：<https://go.dev>

## 分隔线

---

## 数学公式

行内公式：$E = mc^2$

块级公式：

$$
\sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

另一个公式：

$$
\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$

---

*由 GoStreamingMarkdown 渲染 - 支持完整的 GFM 语法*
`

	// 渲染并输出
	output := gsm.Render(markdown, 80, nil)
	fmt.Println(output)
}
