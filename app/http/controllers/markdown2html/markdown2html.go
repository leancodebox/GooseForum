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
	"github.com/yuin/goldmark/text"
)

func GetVersion() uint32 {
	return 3
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
)

// MarkdownToHTML renders Markdown to HTML with the shared server parser.
func MarkdownToHTML(markdown string) string {
	var buf bytes.Buffer
	ctx := parser.NewContext(parser.WithIDs(headingid.NewIDs()))
	if err := md.Convert([]byte(markdown), &buf, parser.WithContext(ctx)); err != nil {
		slog.Error("转化失败", "err", err)
	}
	return buf.String()
}

// GetParser returns the shared goldmark parser.
func GetParser() goldmark.Markdown {
	return md
}

// ExtractDescription extracts readable summary text from Markdown.
func ExtractDescription(content string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 200
	}

	reader := text.NewReader([]byte(content))
	doc := GetParser().Parser().Parse(reader)

	var textParts []string
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node := n.(type) {
			case *ast.Text:
				textContent := string(node.Segment.Value(reader.Source()))
				textContent = strings.TrimSpace(textContent)
				if textContent != "" && len(textContent) > 3 {
					textParts = append(textParts, textContent)
				}
			case *ast.CodeBlock, *ast.FencedCodeBlock:
				return ast.WalkSkipChildren, nil
			case *ast.Image:
				return ast.WalkSkipChildren, nil
			}
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return fallbackExtractDescription(content, maxLength)
	}

	description := strings.Join(textParts, " ")
	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.ReplaceAll(description, "\t", " ")
	for strings.Contains(description, "  ") {
		description = strings.ReplaceAll(description, "  ", " ")
	}
	description = strings.TrimSpace(description)

	if utf8.RuneCountInString(description) > maxLength {
		runes := []rune(description)
		if len(runes) > maxLength {
			description = string(runes[:maxLength]) + "..."
		}
	}

	return description
}

// fallbackExtractDescription strips common Markdown markers without parsing.
func fallbackExtractDescription(content string, maxLength int) string {
	lines := strings.Split(content, "\n")
	var textLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "```") {
			continue
		}

		if strings.Contains(line, "![]") || (strings.Contains(line, "![") && strings.Contains(line, "](") && strings.Contains(line, ")")) {
			continue
		}

		if strings.HasPrefix(line, "#") {
			line = strings.TrimLeft(line, "# ")
		}

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
			line = line[2:]
		}

		if len(line) > 10 {
			textLines = append(textLines, line)
		}
	}

	description := strings.Join(textLines, " ")

	if utf8.RuneCountInString(description) > maxLength {
		runes := []rune(description)
		if len(runes) > maxLength {
			description = string(runes[:maxLength]) + "..."
		}
	}

	return description
}

// ExtractSearchContent extracts searchable text while preserving useful Markdown context.
func ExtractSearchContent(content string) string {
	reader := text.NewReader([]byte(content))
	doc := GetParser().Parser().Parse(reader)

	var searchBuf strings.Builder
	stack := make([]ast.Node, 0, 8)

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if n.Kind() == ast.KindHeading || n.Kind() == ast.KindThematicBreak {
				if searchBuf.Len() > 0 {
					searchBuf.WriteByte('\n')
				}
			}
			stack = append(stack, n)
		} else {
			stack = stack[:len(stack)-1]
		}

		switch node := n.(type) {
		case *ast.Text:
			if entering {
				if searchBuf.Len() > 0 && shouldInsertSpace(stack) {
					searchBuf.WriteByte(' ')
				}
				segment := node.Segment
				searchBuf.Write(segment.Value(reader.Source()))
			}

		case *ast.Link:
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

	return compactWhitespace(searchBuf.String())
}

// shouldInsertSpace decides whether adjacent text nodes need a separator.
func shouldInsertSpace(stack []ast.Node) bool {
	if len(stack) == 0 {
		return false
	}

	for i := len(stack) - 1; i >= 0; i-- {
		switch stack[i].Kind() {
		case ast.KindParagraph:
			return true
		case ast.KindHeading:
			return true
		}
	}
	return false
}

// compactWhitespace collapses repeated whitespace while preserving paragraph breaks.
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

		if i < len(lines)-1 && strings.HasPrefix(trimmed, "#") {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
