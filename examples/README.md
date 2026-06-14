# GoStreamingMarkdown 示例

本目录包含 GoStreamingMarkdown 的使用示例。

## 文件说明

- `sample.md` - 示例 Markdown 文档，展示各种语法支持
- `basic-usage.sh` - 基本使用示例脚本
- `stream-demo.sh` - 流式渲染演示脚本

## 运行示例

### 前置条件

确保已编译项目：

```bash
cd ..
go build -o GoStreamingMarkdown .
```

### 运行基本示例

```bash
./basic-usage.sh
```

### 运行流式演示

```bash
./stream-demo.sh
```

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
