package markdown2html

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/mermaid"
	"log/slog"
	"strings"
	"unicode/utf8"
)

func GetVersion() uint32 {
	return 2
}

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

// GetParser 获取 goldmark 解析器实例
func GetParser() goldmark.Markdown {
	return md
}

// ExtractDescription 从markdown内容中智能提取描述
func ExtractDescription(content string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 200 // 默认最大长度
	}

	// 使用 goldmark 解析 markdown
	reader := text.NewReader([]byte(content))
	doc := GetParser().Parser().Parse(reader)

	// 提取纯文本
	var textParts []string
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node := n.(type) {
			case *ast.Text:
				// 提取文本节点内容
				text := string(node.Segment.Value(reader.Source()))
				text = strings.TrimSpace(text)
				if text != "" && len(text) > 3 {
					textParts = append(textParts, text)
				}
			case *ast.CodeBlock, *ast.FencedCodeBlock:
				// 跳过代码块
				return ast.WalkSkipChildren, nil
			case *ast.Image:
				// 跳过图片
				return ast.WalkSkipChildren, nil
			}
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		// 如果解析失败，回退到简单的文本清理
		return fallbackExtractDescription(content, maxLength)
	}

	// 合并文本并清理
	description := strings.Join(textParts, " ")
	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.ReplaceAll(description, "\t", " ")
	// 清理多余空格
	for strings.Contains(description, "  ") {
		description = strings.ReplaceAll(description, "  ", " ")
	}
	description = strings.TrimSpace(description)

	// 截断到指定长度，确保不会截断中文字符
	if utf8.RuneCountInString(description) > maxLength {
		runes := []rune(description)
		if len(runes) > maxLength {
			description = string(runes[:maxLength]) + "..."
		}
	}

	return description
}

// fallbackExtractDescription 简单的回退文本提取方法
func fallbackExtractDescription(content string, maxLength int) string {
	// 简单清理：移除常见的 markdown 标记
	lines := strings.Split(content, "\n")
	var textLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 跳过代码块标记
		if strings.HasPrefix(line, "```") {
			continue
		}

		// 跳过图片
		if strings.Contains(line, "![]") || (strings.Contains(line, "![") && strings.Contains(line, "](") && strings.Contains(line, ")")) {
			continue
		}

		// 移除标题标记
		if strings.HasPrefix(line, "#") {
			line = strings.TrimLeft(line, "# ")
		}

		// 移除列表标记
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
			line = line[2:]
		}

		if len(line) > 10 {
			textLines = append(textLines, line)
		}
	}

	description := strings.Join(textLines, " ")

	// 截断到指定长度
	if utf8.RuneCountInString(description) > maxLength {
		runes := []rune(description)
		if len(runes) > maxLength {
			description = string(runes[:maxLength]) + "..."
		}
	}

	return description
}
