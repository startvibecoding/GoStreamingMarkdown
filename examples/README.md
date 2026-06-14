# GoStreamingMarkdown 示例

本目录包含 GoStreamingMarkdown 的使用示例。

## 目录结构

```
examples/
├── README.md                # 本文档
├── sample.md                # 示例 Markdown 文档
├── basic-usage.sh           # 基本使用示例脚本
├── stream-demo.sh           # 流式渲染演示脚本
├── library-usage/           # 库使用示例
│   ├── main.go
│   └── README.md
├── custom-theme/            # 自定义主题示例
│   └── main.go
├── streaming-chat/          # 流式聊天演示
│   └── main.go
├── markdown-showcase/       # Markdown 语法展示
│   └── main.go
└── large-document/          # 大型文档处理示例
    └── main.go
```

## 快速开始

### 前置条件

确保已编译项目：

```bash
cd ..
go build -o GoStreamingMarkdown .
```

### 运行命令行示例

```bash
# 基本使用
./basic-usage.sh

# 流式渲染演示
./stream-demo.sh

# 渲染示例文件
../GoStreamingMarkdown sample.md
```

### 运行 Go 示例

```bash
# 库使用示例
cd library-usage && go run main.go

# 自定义主题示例
cd custom-theme && go run main.go

# 流式聊天演示
cd streaming-chat && go run main.go

# Markdown 语法展示
cd markdown-showcase && go run main.go

# 大型文档处理示例
cd large-document && go run main.go
```

## 示例说明

### 1. 基本使用 (basic-usage.sh)

展示命令行基本用法：渲染文件、管道输入、主题切换、宽度设置。

### 2. 流式渲染 (stream-demo.sh)

模拟流式输入，展示 `--stream` 模式的实时渲染效果。

### 3. 库使用 (library-usage/)

展示如何在 Go 项目中使用 GoStreamingMarkdown 作为库：
- 一次性渲染
- 流式渲染
- 主题选择

### 4. 自定义主题 (custom-theme/)

展示如何创建和使用自定义主题，包括霓虹风格主题。

### 5. 流式聊天 (streaming-chat/)

模拟 AI 聊天机器人的流式 Markdown 输出，展示实时渲染效果。

### 6. Markdown 语法展示 (markdown-showcase/)

展示所有支持的 Markdown 语法：标题、代码块、列表、表格、引用、链接、数学公式等。

### 7. 大型文档处理 (large-document/)

展示如何处理大型文档，包括性能测试和分块渲染。

## 命令行用法

```bash
# 渲染文件
../GoStreamingMarkdown sample.md

# 从管道读取
echo '# Hello' | ../GoStreamingMarkdown

# 流式模式
cat sample.md | ../GoStreamingMarkdown --stream --delay 50ms

# 浅色主题
../GoStreamingMarkdown -t light sample.md

# 指定宽度
../GoStreamingMarkdown -w 100 sample.md
```
