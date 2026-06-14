#!/bin/bash
# 基本使用示例
# 使用方法: ./examples/basic-usage.sh

set -e

BINARY="./GoStreamingMarkdown"

# 检查二进制文件是否存在
if [ ! -f "$BINARY" ]; then
    echo "正在构建 GoStreamingMarkdown..."
    go build -o GoStreamingMarkdown .
fi

echo "=== 基本使用示例 ==="
echo ""

# 示例 1: 渲染文件
echo "1. 渲染示例 Markdown 文件:"
echo "   $ GoStreamingMarkdown examples/sample.md"
echo ""
$BINARY examples/sample.md
echo ""

echo "=========================================="
echo ""

# 示例 2: 从管道读取
echo "2. 从管道读取 Markdown:"
echo "   $ echo '# Hello **world**' | GoStreamingMarkdown"
echo ""
echo '# Hello **world**' | $BINARY
echo ""

echo "=========================================="
echo ""

# 示例 3: 使用浅色主题
echo "3. 使用浅色主题:"
echo "   $ echo '# 浅色主题' | GoStreamingMarkdown -t light"
echo ""
echo '# 浅色主题示例' | $BINARY -t light
echo ""

echo "=========================================="
echo ""

# 示例 4: 指定宽度
echo "4. 指定终端宽度 (60列):"
echo "   $ GoStreamingMarkdown -w 60 examples/sample.md"
echo ""
$BINARY -w 60 examples/sample.md
