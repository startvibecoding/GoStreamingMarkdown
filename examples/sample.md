# GoStreamingMarkdown 示例文档

这是一个展示 GoStreamingMarkdown 渲染效果的示例文档。

## 文本格式

这是 **粗体** 文本，这是 *斜体* 文本，这是 ~~删除线~~ 文本。

行内代码：`fmt.Println("Hello")`

## 代码块

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, GoStreamingMarkdown!")
}
```

## 列表

无序列表：
- 第一项
- 第二项
  - 嵌套项
- 第三项

有序列表：
1. 步骤一
2. 步骤二
3. 步骤三

任务列表：
- [x] 已完成的任务
- [ ] 待办任务
- [ ] 另一个待办任务

## 引用

> 这是一段引用文本。
>
> 可以多行引用。

## 表格

| 功能 | 状态 | 说明 |
|------|:----:|------|
| 标题 | ✅ | ATX 风格 |
| 粗体 | ✅ | `**text**` |
| 代码 | ✅ | 行内和块级 |
| 表格 | ✅ | GFM 表格 |

## 链接和图片

访问 [Go 官网](https://go.dev) 了解更多。

图片：![Go Logo](https://go.dev/images/go-logo.png)

## 分隔线

---

## 数学公式

行内公式：$E = mc^2$

块级公式：
$$
\sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

---

*由 GoStreamingMarkdown 渲染*
