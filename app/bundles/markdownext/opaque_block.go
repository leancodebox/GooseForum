package markdownext

import (
	"html"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindOpaqueBlock = ast.NewNodeKind("ForumOpaqueBlock")

// OpaqueBlock keeps future extension syntax literal, including its contents.
type OpaqueBlock struct {
	ast.BaseBlock
	stop int
}

func (node *OpaqueBlock) Kind() ast.NodeKind { return KindOpaqueBlock }
func (node *OpaqueBlock) IsRaw() bool        { return true }
func (node *OpaqueBlock) Dump(source []byte, level int) {
	ast.DumpHelper(node, source, level, nil, nil)
}
func (node *OpaqueBlock) Text(source []byte) []byte {
	return node.Lines().Value(source)
}

type opaqueBlockParser struct{}

var opaqueBlockPairsKey = parser.NewContextKey()

func opaqueBlockPairs(source []byte, context parser.Context) map[int]int {
	if pairs, ok := context.Get(opaqueBlockPairsKey).(map[int]int); ok {
		return pairs
	}
	literals, _ := context.Get(literalRangesKey).([]opaqueRange)
	pairs := map[int]int{}
	type openTag struct {
		name  string
		start int
	}
	var stack []openTag
	for offset := 0; offset < len(source); {
		end := offset
		for end < len(source) && source[end] != '\n' {
			end++
		}
		start := offset
		for start < end && (source[start] == ' ' || source[start] == '\t' || source[start] == '>') {
			start++
		}
		line := strings.TrimSpace(string(source[start:end]))
		if start < end && !inRange(literals, start) {
			if strings.HasPrefix(line, "[/") && strings.HasSuffix(line, "]") {
				name := line[2 : len(line)-1]
				if len(stack) > 0 && stack[len(stack)-1].name == name {
					open := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					pairs[open.start] = start + len(line)
				}
			} else if name, _, stop, valid := ParseTag(line); valid && name != "mention" && len(stack) < 8 && strings.TrimSpace(line[stop:]) == "" {
				stack = append(stack, openTag{name, start})
			}
		}
		offset = end + 1
	}
	context.Set(opaqueBlockPairsKey, pairs)
	return pairs
}

func (opaqueBlockParser) Trigger() []byte { return []byte{'['} }
func (opaqueBlockParser) Open(parent ast.Node, reader text.Reader, context parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	offset := context.BlockOffset()
	if offset < 0 || offset >= len(line) {
		return nil, parser.NoChildren
	}
	name, _, end, valid := ParseTag(string(line[offset:]))
	if !valid || name == "mention" || !util.IsBlank(line[offset+end:]) {
		return nil, parser.NoChildren
	}
	start := segment.Start + offset
	protectedByUnknown(reader.Source(), start, context)
	pairs := opaqueBlockPairs(reader.Source(), context)
	stop := pairs[start]
	if stop <= segment.Stop {
		return nil, parser.NoChildren
	}
	source := reader.Source()
	closeStart := stop - len(name) - 3
	closeLineStart := closeStart
	for closeLineStart > 0 && source[closeLineStart-1] != '\n' {
		closeLineStart--
	}
	closeLineStop := stop
	for closeLineStop < len(source) && source[closeLineStop] != '\n' {
		closeLineStop++
	}
	// A block delimiter occupies its own line. Quoted containers include their
	// prefix in source offsets; Goldmark strips that prefix from stored segments.
	prefix := strings.TrimSpace(string(source[closeLineStart:closeStart]))
	if strings.Trim(prefix, "> ") != "" || strings.TrimSpace(string(source[stop:closeLineStop])) != "" {
		return nil, parser.NoChildren
	}
	node := &OpaqueBlock{stop: stop}
	segment.Start = start
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return node, parser.NoChildren
}

func (opaqueBlockParser) Continue(block ast.Node, reader text.Reader, context parser.Context) parser.State {
	node := block.(*OpaqueBlock)
	_, segment := reader.PeekLine()
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	if segment.Stop >= node.stop {
		return parser.Close
	}
	return parser.Continue | parser.NoChildren
}
func (opaqueBlockParser) Close(ast.Node, text.Reader, parser.Context) {}
func (opaqueBlockParser) CanInterruptParagraph() bool                 { return false }
func (opaqueBlockParser) CanAcceptIndentedLine() bool                 { return false }

type opaqueBlockRenderer struct{}

func (opaqueBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindOpaqueBlock, func(w util.BufWriter, source []byte, block ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			_, _ = w.WriteString("<p>" + html.EscapeString(string(block.Text(source))) + "</p>\n")
		}
		return ast.WalkSkipChildren, nil
	})
}
