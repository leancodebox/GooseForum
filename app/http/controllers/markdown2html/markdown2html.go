package markdown2html

import (
	"bytes"
	"log/slog"
	"strings"
	"unicode/utf8"

	headingid "github.com/jkboxomine/goldmark-headingid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
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
		extension.Typographer,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
	),
)

// MarkdownToHTML 添加新的服务端渲染的控制器方法
func MarkdownToHTML(markdown string) string {
	var buf bytes.Buffer
	// 创建带有改进的标题 ID 生成的上下文
	ctx := parser.NewContext(parser.WithIDs(headingid.NewIDs()))
	if err := md.Convert([]byte(markdown), &buf, parser.WithContext(ctx)); err != nil {
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

// ExtractSearchContent 为全文搜索优化的提取器
func ExtractSearchContent(content string) string {
	reader := text.NewReader([]byte(content))
	doc := GetParser().Parser().Parse(reader)

	var searchBuf strings.Builder
	stack := make([]ast.Node, 0, 8) // 节点栈用于上下文感知

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		// 上下行关系处理
		if entering {
			// 识别上下文分隔点（标题/分隔线）
			if n.Kind() == ast.KindHeading || n.Kind() == ast.KindThematicBreak {
				if searchBuf.Len() > 0 {
					searchBuf.WriteByte('\n')
				}
			}
			stack = append(stack, n)
		} else {
			stack = stack[:len(stack)-1] // 弹出当前节点
		}

		// 内容提取策略
		switch node := n.(type) {
		case *ast.Text:
			if entering {
				// 智能空格插入（考虑中文无空格特性）
				if searchBuf.Len() > 0 && shouldInsertSpace(stack) {
					searchBuf.WriteByte(' ')
				}
				segment := node.Segment
				searchBuf.Write(segment.Value(reader.Source()))
			}

		case *ast.Link:
			// 保留链接文本+URL（提升搜索命中率）
			if entering {
				searchBuf.WriteString("[")
			} else {
				dest := node.Destination
				searchBuf.WriteString("](")
				searchBuf.Write(dest)
				searchBuf.WriteString(")")
			}

		case *ast.CodeSpan:
			if entering {
				searchBuf.WriteString("`")
			} else {
				searchBuf.WriteString("`")
			}

		case *ast.Image:
			// 提取alt文本 + URL（可搜索图片内容）
			if entering {
				searchBuf.WriteString("![")
			} else {
				dest := node.Destination
				searchBuf.WriteString("](")
				searchBuf.Write(dest)
				searchBuf.WriteString(")")
			}

		case *ast.CodeBlock, *ast.FencedCodeBlock:
			return ast.WalkSkipChildren, nil

		case *ast.HTMLBlock:
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	// 压缩冗余空白（保留换行结构）
	return compactWhitespace(searchBuf.String())
}

// 智能空格决策（中文不插入空格）
func shouldInsertSpace(stack []ast.Node) bool {
	if len(stack) == 0 {
		return false
	}

	// 检查上层节点是否中文上下文
	for i := len(stack) - 1; i >= 0; i-- {
		switch stack[i].Kind() {
		case ast.KindParagraph:
			return true // 英文段落需要空格
		case ast.KindHeading:
			// 根据语言检测决定（此处简化实现）
			return true
		}
	}
	return false
}

// 保留段落分隔的空白压缩
func compactWhitespace(s string) string {
	var buf strings.Builder
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(trimmed)

		// 保留标题后的空行
		if i < len(lines)-1 && strings.HasPrefix(trimmed, "#") {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
