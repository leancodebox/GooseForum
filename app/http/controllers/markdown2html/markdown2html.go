package markdown2html

import (
	"bytes"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
	"log/slog"
)

func GetVersion() uint32 {
	return 2
}

var a = highlighting.Highlighting
var b = highlighting.NewHighlighting(
	highlighting.WithStyle("monokai"),
	highlighting.WithFormatOptions(
		chromahtml.WithLineNumbers(true),
	),
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		&mermaid.Extender{},
		//b,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// 添加新的服务端渲染的控制器方法
func MarkdownToHTML(markdown string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		slog.Error("转化失败", "err", err)
	}
	return buf.String()
}
