#!/bin/bash
# 流式渲染演示脚本
# 使用方法: ./examples/stream-demo.sh

set -e

BINARY="./GoStreamingMarkdown"

# 检查二进制文件是否存在
if [ ! -f "$BINARY" ]; then
    echo "正在构建 GoStreamingMarkdown..."
    go build -o GoStreamingMarkdown .
fi

echo "开始流式渲染演示..."
echo ""

# 模拟流式输入
(
    echo "# 流式渲染演示"
    sleep 0.5
    echo ""
    echo "文本会逐行出现..."
    sleep 0.5
    echo ""
    echo "**粗体** 和 *斜体* 文本"
    sleep 0.5
    echo ""
    echo '```go'
    sleep 0.3
    echo 'func main() {'
    sleep 0.3
    echo '    fmt.Println("Hello!")'
    sleep 0.3
    echo '}'
    sleep 0.3
    echo '```'
    sleep 0.5
    echo ""
    echo "> 这是一段引用"
    sleep 0.5
    echo ""
    echo "- 列表项 1"
    sleep 0.3
    echo "- 列表项 2"
    sleep 0.3
    echo "- 列表项 3"
    sleep 0.3
    echo ""
    echo "---"
    sleep 0.3
    echo ""
    echo "演示完成！"
) | $BINARY --stream --delay 100ms
