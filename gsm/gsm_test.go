package gsm

import (
	"strings"
	"testing"
	"unicode/utf8"
	"unicode"

	"github.com/startvibecoding/GoStreamingMarkdown/renderer"
)

func stripANSI(s string) string {
	return renderer.StripANSI(s)
}

func compactWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func chunkByRunes(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 1
	}

	runes := []rune(s)
	chunks := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func TestStreamOutputPreservesUnicodeOrderForSSEText(t *testing.T) {
	const sseText = "已改回 `https://se.lab.bza.edu.cn`,编译通过。\n\n现在 baseURL 默认是 `https://se.lab.bza.edu.cn`,仍保留了 `OSCANNER_BASE_URL` 环境变量覆盖能力(需要时 `OSCANNER_BASE_URL=xxx go run .` 即可换地址)。\n\n需要注意:之前提交检查那步触发了 429 限流(每个 IP 5 小时内最多 3 次)。带上 `x-secret-token` 后仍被限流,说明该 token 在公网 443 入口未豁免限流。等限流窗口过去,或换个出口 IP,再跑 `/tmp/demo` 就能完整走完 1–5 步。要我现在重试一次看是否已解除限流吗?"
	expectedInOrder := []string{
		"已改回",
		"https://se.lab.bza.edu.cn",
		"编译通过。",
		"现在 baseURL 默认是",
		"OSCANNER_BASE_URL",
		"环境变量覆盖能力",
		"需要注意:之前提交检查那步触发了 429 限流",
		"x-secret-token",
		"公网 443 入口未豁免限流",
		"/tmp/demo",
		"完整走完 1–5 步",
		"要我现在重试一次看是否已解除限流吗?",
	}

	stream := NewStream(80, nil)
	var accumulated string

	for _, chunk := range chunkByRunes(sseText, 7) {
		accumulated += chunk
		stream.Update(accumulated)

		out := stripANSI(stream.Output())
		if !utf8.ValidString(out) {
			t.Fatalf("stream output should stay valid UTF-8.\noutput: %q", out)
		}
	}

	final := compactWhitespace(stripANSI(stream.Output()))
	lastIndex := -1
	for _, part := range expectedInOrder {
		idx := strings.Index(final, compactWhitespace(part))
		if idx < 0 {
			t.Fatalf("final output is missing expected fragment %q.\noutput: %q", part, final)
		}
		if idx < lastIndex {
			t.Fatalf("final output reordered fragment %q.\noutput: %q", part, final)
		}
		lastIndex = idx
	}
}
