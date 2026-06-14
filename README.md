# GoStreamingMarkdown

**零依赖** 的 Go CLI 流式 Markdown → ANSI 终端渲染器。

从 Swift 项目的 [SwiftStreamingMarkdown](https://github.com/SwiftStreamingMarkdown) 库重构而来，
完整实现了 Preprocess → Parse → Rewrite → Render 四阶段管线。支持流式渲染模式。

## 特性

| 类别 | 支持 |
|------|------|
| **标题** | ATX `# ~ ######`，带颜色分级 |
| **强调** | `*italic*`, `**bold**`, `~~strikethrough~~` |
| **代码** | 行内 `` `code` ``，围栏代码块 ` ```lang ``` `，带语言标签 |
| **列表** | 有序 `1.`，无序 `- * +`，任务列表 `- [x]` |
| **引用** | `> blockquote`，支持嵌套 |
| **表格** | `| a | b |` GFM 表格，带边框 |
| **链接** | `[text](url)`, `<autolink>` |
| **图片** | `![alt](url)` 终端渲染 |
| **分隔线** | `---`, `***`, `___` |
| **LaTeX 数学** | `$$block$$`, `$inline$`, `\(paren\)`，预处理为代码块/行内代码 |
| **流式渲染** | `--stream` 模式，投机性 emphasis 闭合（防抖动） |

## 安装

```bash
cd GoStreamingMarkdown
go build -o GoStreamingMarkdown .
```

**零外部依赖** — 仅使用 Go 标准库。

## 使用

```bash
# 渲染文件
GoStreamingMarkdown README.md

# 从管道读取
cat README.md | GoStreamingMarkdown

# 流式模式（实时渲染，增量更新）
cat stream.txt | GoStreamingMarkdown --stream --delay 100ms

# 指定主题和宽度
GoStreamingMarkdown -t light -w 100 doc.md

# 渲染 Markdown 字符串
echo '# Hello **world**' | GoStreamingMarkdown
```

### 参数

| 参数 | 说明 |
|------|------|
| `-h, --help` | 显示帮助 |
| `-s, --stream` | 流式模式（原地重绘） |
| `-d, --delay <dur>` | 流式更新间隔（如 `50ms`, `1s`） |
| `-t, --theme <name>` | 主题：`dark`（默认）或 `light` |
| `-w, --width <cols>` | 终端宽度（默认 80） |

## 架构

完整复刻 Swift 项目的四阶段管线：

```
输入文本
  │
  ▼
┌─────────────────────┐
│ 1. Preprocess (LaTeX)│  $$...$$ → fenced code block
│                      │  $...$   → inline code
│                      │  \(...\) → inline code
└──────────┬──────────┘
           ▼
┌─────────────────────┐
│ 2. Parse (AST)       │  CommonMark 兼容解析器
│                      │  Block: heading, paragraph, code,
│                      │  blockquote, list, table, hr
│                      │  Inline: emphasis, strong, code,
│                      │  link, image, strikethrough, autolink
└──────────┬──────────┘
           ▼
┌─────────────────────┐
│ 3. Rewrite           │  投机性 emphasis 闭合
│ (speculative)        │  防止流式渲染时文本抖动
└──────────┬──────────┘
           ▼
┌─────────────────────┐
│ 4. Render (ANSI)     │  AST → ANSI 终端输出
│                      │  深色/浅色主题
│                      │  Unicode box-drawing 字符
└─────────────────────┘
```

## 项目结构

```
GoStreamingMarkdown/
├── main.go              # CLI 入口
├── go.mod               # Go 模块定义（零依赖）
├── parser/
│   ├── node.go          # AST 节点类型定义
│   └── parser.go        # Markdown 解析器 + LaTeX 预处理 + 重写器
├── renderer/
│   └── renderer.go      # ANSI 终端渲染器 + 主题系统
└── README.md
```

## 作为库使用

```go
package main

import (
    "fmt"
    "GoStreamingMarkdown/parser"
    "GoStreamingMarkdown/renderer"
)

func main() {
    src := "# Hello **world**"
    
    // 解析
    doc := parser.Parse(src, parser.DefaultOption())
    
    // 渲染
    theme := renderer.DefaultTheme()
    r := renderer.New(theme, 80)
    fmt.Println(r.Render(doc))
    
    // 或一步完成
    fmt.Println(renderer.Render(src, 80, theme))
}
```

### 流式解析

```go
opt := parser.StreamOption() // 启用投机性重写
doc := parser.Parse(partialText, opt)
```

## 从 Swift 到 Go 的映射

| Swift 组件 | Go 对应 |
|-----------|---------|
| `LaTexPreProcessor` | `parser.preprocessLaTeX()` |
| `swift-markdown` (cmark-gfm) | `parser.Parse()` |
| `PartialEmphasisRewriter` | `parser.rewriteSpeculative()` |
| `MarkdownRenderable` | `parser.Node` AST |
| `DocumentView` | `renderer.Render()` |
| `MarkdownRenderConfig` + `Colors` | `renderer.Theme` |
| `StreamedMarkdownSource` | `--stream` CLI 模式 |

## License

MIT
