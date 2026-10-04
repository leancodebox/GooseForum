package markdownext

import (
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const Version uint8 = 1

var Kind = ast.NewNodeKind("ForumExtension")

type Node struct {
	ast.BaseInline
	Start, End           int
	Raw, Label, Username string
	UserID               uint64
	Invalid              bool
}

func (n *Node) Kind() ast.NodeKind            { return Kind }
func (n *Node) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }
func (n *Node) Text(source []byte) []byte {
	if n.UserID != 0 {
		return []byte(n.Label)
	}
	return []byte(n.Raw)
}

// ParseTag is the common, bounded header grammar for registered extensions.
func ParseTag(value string) (string, map[string]string, int, bool) {
	if len(value) == 0 || value[0] != '[' {
		return "", nil, 0, false
	}
	nameEnd := 1
	for nameEnd < len(value) && (value[nameEnd] >= 'a' && value[nameEnd] <= 'z' || nameEnd > 1 && (value[nameEnd] >= '0' && value[nameEnd] <= '9' || value[nameEnd] == '-')) {
		nameEnd++
	}
	if nameEnd == 1 {
		return "", nil, 0, false
	}
	attrs := map[string]string{}
	i := nameEnd
	for i < len(value) && i < 1024 {
		if value[i] == ']' {
			return value[1:nameEnd], attrs, i + 1, true
		}
		if value[i] != ' ' {
			return "", nil, 0, false
		}
		for i < len(value) && value[i] == ' ' {
			i++
		}
		if i < len(value) && value[i] == ']' {
			continue
		}
		start := i
		for i < len(value) && (value[i] >= 'a' && value[i] <= 'z' || i > start && (value[i] >= '0' && value[i] <= '9' || value[i] == '-')) {
			i++
		}
		if start == i || i+1 >= len(value) || value[i:i+2] != "=\"" {
			return "", nil, 0, false
		}
		key := value[start:i]
		i += 2
		var decoded strings.Builder
		for i < len(value) && i < 1024 && value[i] != '"' {
			if value[i] < 32 || value[i] == 127 {
				return "", nil, 0, false
			}
			if value[i] == '\\' {
				i++
				if i >= len(value) || !strings.ContainsRune("\\\"]", rune(value[i])) {
					return "", nil, 0, false
				}
			}
			decoded.WriteByte(value[i])
			i++
		}
		if i >= len(value) || i >= 1024 {
			return "", nil, 0, false
		}
		if _, exists := attrs[key]; exists {
			return "", nil, 0, false
		}
		attrs[key] = decoded.String()
		i++
	}
	return "", nil, 0, false
}

type mentionParser struct{}

type opaqueRange struct{ start, end int }

var opaqueRangesKey = parser.NewContextKey()
var opaquePairsKey = parser.NewContextKey()
var htmlRangesKey = parser.NewContextKey()
var literalRangesKey = parser.NewContextKey()
var tagMarkers = regexp.MustCompile(`\[(/?)([a-z][a-z0-9-]*)(?: |\])`)
var htmlMarkers = regexp.MustCompile(`</?([A-Za-z][A-Za-z0-9-]*)(?:\s[^<>]*|)>`)
var contextMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Linkify))

func literalRanges(source []byte) []opaqueRange {
	var ranges []opaqueRange
	doc := contextMarkdown.Parser().Parse(text.NewReader(source))
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.CodeBlock, *ast.FencedCodeBlock, *ast.HTMLBlock:
			for i := 0; i < node.Lines().Len(); i++ {
				s := node.Lines().At(i)
				ranges = append(ranges, opaqueRange{s.Start, s.Stop})
			}
			return ast.WalkSkipChildren, nil
		case *ast.CodeSpan, *ast.Link, *ast.Image:
			_ = ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					if t, ok := child.(*ast.Text); ok {
						start := t.Segment.Start
						if (node.Kind() == ast.KindLink || node.Kind() == ast.KindImage) && start > 0 && source[start-1] == '[' {
							start--
						}
						ranges = append(ranges, opaqueRange{start, t.Segment.Stop})
					}
				}
				return ast.WalkContinue, nil
			})
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start < ranges[j].start })
	return ranges
}
func inRange(ranges []opaqueRange, position int) bool {
	i := sort.Search(len(ranges), func(i int) bool { return ranges[i].start > position }) - 1
	return i >= 0 && position < ranges[i].end
}
func escaped(source []byte, position int) bool {
	count := 0
	for position > 0 && source[position-1] == '\\' {
		position--
		count++
	}
	return count%2 == 1
}

func protectedByUnknown(source []byte, position int, context parser.Context) bool {
	ranges, initialized := context.Get(opaqueRangesKey).([]opaqueRange)
	if !initialized {
		literals := literalRanges(source)
		context.Set(literalRangesKey, literals)
		pairs := map[int]int{}
		type openTag struct {
			name  string
			start int
		}
		var stack []openTag
		for _, match := range tagMarkers.FindAllSubmatchIndex(source, -1) {
			if inRange(literals, match[0]) || escaped(source, match[0]) {
				continue
			}
			name := string(source[match[4]:match[5]])
			if name == "mention" {
				continue
			}
			if match[2] != match[3] {
				if len(stack) > 0 && stack[len(stack)-1].name == name {
					last := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					ranges = append(ranges, opaqueRange{last.start, match[1]})
					pairs[last.start] = match[1]
				}
			} else if len(stack) < 8 {
				stack = append(stack, openTag{name, match[0]})
			}
		}
		sort.Slice(ranges, func(i, j int) bool { return ranges[i].start < ranges[j].start })
		merged := make([]opaqueRange, 0, len(ranges))
		for _, interval := range ranges {
			if len(merged) > 0 && interval.start <= merged[len(merged)-1].end {
				merged[len(merged)-1].end = max(merged[len(merged)-1].end, interval.end)
			} else {
				merged = append(merged, interval)
			}
		}
		ranges = merged
		context.Set(opaqueRangesKey, ranges)
		context.Set(opaquePairsKey, pairs)
	}
	i := sort.Search(len(ranges), func(i int) bool { return ranges[i].start >= position }) - 1
	return i >= 0 && position < ranges[i].end
}

func (mentionParser) Trigger() []byte { return []byte{'[', '@'} }
func (mentionParser) Parse(parent ast.Node, reader text.Reader, context parser.Context) ast.Node {
	if context.IsInLinkLabel() {
		return nil
	}
	line, segment := reader.PeekLine()
	if protectedByUnknown(reader.Source(), segment.Start, context) {
		return nil
	}
	// Inline HTML tags do not enclose their text in the Markdown AST.
	if insideInlineHTML(reader.Source(), segment.Start, parent, context) {
		return nil
	}
	value := string(line[:min(len(line), 1162)])
	n := &Node{Start: segment.Start}
	if value[0] == '@' {
		if segment.Start > 0 {
			previous, _ := utf8.DecodeLastRune(reader.Source()[:segment.Start])
			if previous < 128 && (unicode.IsLetter(previous) || unicode.IsDigit(previous) || strings.ContainsRune("_-.@/:", previous)) {
				return nil
			}
		}
		i := 1
		for i < len(value) && usernameByte(value[i]) {
			i++
		}
		if i < 5 || i > 33 {
			return nil
		}
		if i < len(value) && (value[i] == '.' && i+1 < len(value) && usernameByte(value[i+1]) || value[i] == '/' || value[i] == '@') {
			return nil
		}
		n.Username = value[1:i]
		n.Raw = value[:i]
	} else {
		name, attrs, end, ok := ParseTag(value)
		bracketEnd := strings.IndexByte(value, ']') + 1
		if bracketEnd > 0 && bracketEnd < len(value) && (value[bracketEnd] == '(' || value[bracketEnd] == '[') {
			return nil
		}
		if !ok {
			if !strings.HasPrefix(value, "[mention") || len(value) > 8 && usernameByte(value[8]) {
				return nil
			}
			end = strings.IndexByte(value, ']') + 1
			if end <= 0 {
				end = len(strings.TrimSuffix(value, "\n"))
			}
			n.Raw, n.Invalid = value[:end], true
		} else {
			if name != "mention" {
				pairs, _ := context.Get(opaquePairsKey).(map[int]int)
				stop := pairs[segment.Start] - segment.Start
				if stop <= 0 || stop > len(value) {
					return nil
				}
				n.Raw = value[:stop]
				n.End = n.Start + stop
				reader.Advance(stop)
				return n
			}
			closeAt := strings.Index(value[end:min(len(value), end+138)], "[/"+name+"]")
			if closeAt < 0 {
				if name != "mention" {
					return nil
				}
				n.Raw = value[:end]
				n.Invalid = name == "mention"
			} else {
				stop := end + closeAt + len(name) + 3
				n.Raw = value[:stop]
				if name == "mention" {
					id, err := strconv.ParseUint(attrs["user"], 10, 64)
					label := value[end : end+closeAt]
					n.Invalid = err != nil || id == 0 || strconv.FormatUint(id, 10) != attrs["user"] || len(attrs) != 1 || len(label) > 128 || strings.ContainsAny(label, "[]<>\r\n")
					if !n.Invalid {
						n.UserID, n.Label = id, label
					}
				}
			}
		}
	}
	n.End = n.Start + len(n.Raw)
	reader.Advance(len(n.Raw))
	return n
}

func insideInlineHTML(source []byte, position int, parent ast.Node, context parser.Context) bool {
	cache, _ := context.Get(htmlRangesKey).(map[ast.Node][]opaqueRange)
	if cache == nil {
		cache = map[ast.Node][]opaqueRange{}
		context.Set(htmlRangesKey, cache)
	}
	if ranges, ok := cache[parent]; ok {
		return inRange(ranges, position)
	}
	if parent.Lines().Len() == 0 {
		return false
	}
	start, end := parent.Lines().At(0).Start, parent.Lines().At(parent.Lines().Len()-1).Stop
	depth := 0
	open := 0
	var ranges []opaqueRange
	for _, match := range htmlMarkers.FindAllSubmatchIndex(source[start:end], -1) {
		literals, _ := context.Get(literalRangesKey).([]opaqueRange)
		if inRange(literals, start+match[0]) || escaped(source, start+match[0]) {
			continue
		}
		raw := string(source[start+match[0] : start+match[1]])
		name := strings.ToLower(string(source[start+match[2] : start+match[3]]))
		if strings.HasPrefix(raw, "</") {
			if depth > 0 {
				depth--
				if depth == 0 {
					ranges = append(ranges, opaqueRange{open, start + match[1]})
				}
			}
		} else if !strings.HasSuffix(raw, "/>") && !strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+name+" ") {
			if depth == 0 {
				open = start + match[0]
			}
			depth++
		}
	}
	if depth > 0 {
		ranges = append(ranges, opaqueRange{open, end})
	}
	cache[parent] = ranges
	return inRange(ranges, position)
}
func usernameByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
}
func Canonical(id uint64, username string) string {
	return fmt.Sprintf("[mention user=\"%d\"]@%s[/mention]", id, username)
}
func NotificationsAllowed(n ast.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.Kind() {
		case ast.KindBlockquote, ast.KindHeading, ast.KindLink, ast.KindImage:
			return false
		}
	}
	return true
}

type nodeRenderer struct{}

func (nodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(Kind, func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		n := node.(*Node)
		if n.UserID == 0 {
			_, _ = w.WriteString(html.EscapeString(n.Raw))
		} else {
			_, _ = fmt.Fprintf(w, "<a class=\"mention\" data-mention-user=\"%d\" href=\"/u/%d\">%s</a>", n.UserID, n.UserID, html.EscapeString(n.Label))
		}
		return ast.WalkSkipChildren, nil
	})
}

type Extension struct{}

func (Extension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithBlockParsers(util.Prioritized(opaqueBlockParser{}, 850)))
	md.Parser().AddOptions(parser.WithInlineParsers(util.Prioritized(mentionParser{}, 150)))
	md.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(opaqueBlockRenderer{}, 100)))
	md.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(nodeRenderer{}, 100)))
}
