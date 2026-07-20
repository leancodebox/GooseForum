package markdown2html

import (
	"bytes"
	"log/slog"
	"net/url"
	"strings"
	"unicode/utf8"

	headingid "github.com/jkboxomine/goldmark-headingid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	nethtml "golang.org/x/net/html"
)

func GetPostVersion() uint32 {
	return 4
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

// PostMarkdownToHTML renders public user content and applies UGC link/image policies.
func PostMarkdownToHTML(markdown string) string {
	return normalizePostHTML(MarkdownToHTML(markdown))
}

func normalizePostHTML(raw string) string {
	root, err := nethtml.Parse(strings.NewReader("<div>" + raw + "</div>"))
	if err != nil {
		return raw
	}

	var walk func(*nethtml.Node)
	walk = func(node *nethtml.Node) {
		if node.Type == nethtml.ElementNode {
			switch node.Data {
			case "a":
				if isExternalHTTPLink(getHTMLAttr(node, "href")) {
					setHTMLAttr(node, "target", "_blank")
					setHTMLAttr(node, "rel", "nofollow ugc noopener noreferrer")
				}
			case "img":
				setHTMLAttr(node, "loading", "lazy")
				setHTMLAttr(node, "decoding", "async")
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)

	container := findFirstElement(root, "div")
	if container == nil {
		return raw
	}
	var buf bytes.Buffer
	for child := container.FirstChild; child != nil; child = child.NextSibling {
		if err := nethtml.Render(&buf, child); err != nil {
			return raw
		}
	}
	return buf.String()
}

func getHTMLAttr(node *nethtml.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isExternalHTTPLink(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.IsAbs() && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func findFirstElement(node *nethtml.Node, tag string) *nethtml.Node {
	if node == nil {
		return nil
	}
	if node.Type == nethtml.ElementNode && node.Data == tag {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findFirstElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func setHTMLAttr(node *nethtml.Node, key, value string) {
	for i := range node.Attr {
		if node.Attr[i].Key == key {
			node.Attr[i].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, nethtml.Attribute{Key: key, Val: value})
}

// GetParser returns the shared goldmark parser.
func GetParser() goldmark.Markdown {
	return md
}

// ExtractFirstImageURL returns the first public image destination from Markdown.
func ExtractFirstImageURL(content string) string {
	urls := ExtractImageURLs(content)
	if len(urls) == 0 {
		return ""
	}
	return urls[0]
}

func ExtractImageURLs(content string) []string {
	reader := text.NewReader([]byte(content))
	doc := GetParser().Parser().Parse(reader)

	imageURLs := make([]string, 0)
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		image, ok := n.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}
		imageURL := strings.TrimSpace(string(image.Destination))
		if isPublicImageURL(imageURL) {
			imageURLs = append(imageURLs, imageURL)
		}
		return ast.WalkContinue, nil
	})
	return imageURLs
}

func isPublicImageURL(value string) bool {
	if value == "" || strings.HasPrefix(value, "data:") || strings.HasPrefix(value, "blob:") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if parsed.IsAbs() {
		return parsed.Scheme == "http" || parsed.Scheme == "https"
	}
	return strings.HasPrefix(value, "/")
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
			case *ast.Heading, *ast.Paragraph, *ast.ListItem:
				textContent := extractDescriptionBlockText(node, reader.Source())
				if textContent != "" && utf8.RuneCountInString(textContent) > 3 {
					textParts = append(textParts, textContent)
				}
				return ast.WalkSkipChildren, nil
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

// ExtractPreview converts Markdown into compact readable text for notifications and activity lists.
func ExtractPreview(content string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}

	reader := text.NewReader([]byte(content))
	doc := GetParser().Parser().Parse(reader)
	var builder strings.Builder
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Text:
			builder.Write(node.Segment.Value(reader.Source()))
			if node.SoftLineBreak() || node.HardLineBreak() {
				builder.WriteByte(' ')
			}
		case *ast.Image:
			builder.WriteString("[图片]")
			return ast.WalkSkipChildren, nil
		case *ast.CodeBlock, *ast.FencedCodeBlock, *ast.HTMLBlock:
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph, *ast.Heading, *ast.ListItem:
			if builder.Len() > 0 {
				builder.WriteByte(' ')
			}
		}
		return ast.WalkContinue, nil
	})

	preview := strings.Join(strings.Fields(builder.String()), " ")
	runes := []rune(preview)
	if len(runes) > maxLength {
		preview = string(runes[:maxLength])
	}
	return preview
}

func extractDescriptionBlockText(node ast.Node, source []byte) string {
	var builder strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := n.(type) {
		case *ast.Text:
			builder.Write(typed.Segment.Value(source))
			if typed.SoftLineBreak() || typed.HardLineBreak() {
				builder.WriteByte(' ')
			}
		case *ast.CodeBlock, *ast.FencedCodeBlock, *ast.Image:
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})

	textContent := strings.ReplaceAll(builder.String(), "\n", " ")
	textContent = strings.ReplaceAll(textContent, "\t", " ")
	for strings.Contains(textContent, "  ") {
		textContent = strings.ReplaceAll(textContent, "  ", " ")
	}
	return strings.TrimSpace(textContent)
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
			searchBuf.WriteString("`")

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
