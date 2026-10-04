package mentionservice

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/markdownext"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/yuin/goldmark/ast"
	nethtml "golang.org/x/net/html"
)

var ErrInvalid = errors.New("invalid mention syntax or too many recipients (maximum 20)")

func Nodes(content string) []*markdownext.Node {
	_, doc := markdown2html.ParseExtended(content)
	var nodes []*markdownext.Node
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if node, ok := n.(*markdownext.Node); ok {
				nodes = append(nodes, node)
			}
		}
		return ast.WalkContinue, nil
	})
	return nodes
}

func Normalize(content string, version uint8, validate bool) (string, error) {
	if version == 0 {
		return content, nil
	}
	if version != markdownext.Version {
		return "", ErrInvalid
	}
	nodes := Nodes(content)
	if len(nodes) > 100 && validate {
		return "", ErrInvalid
	}
	ids := []uint64{}
	names := []string{}
	for _, n := range nodes {
		if n.Invalid && validate {
			return "", ErrInvalid
		}
		if n.UserID > 0 {
			ids = append(ids, n.UserID)
		}
		if n.Username != "" {
			names = append(names, n.Username)
		}
	}
	identities, err := users.MentionIdentities(ids, names)
	if err != nil {
		return "", err
	}
	byID := map[uint64]users.MentionIdentity{}
	byName := map[string][]users.MentionIdentity{}
	for _, user := range identities {
		byID[user.Id] = user
		key := strings.ToLower(user.Username)
		byName[key] = append(byName[key], user)
	}
	var result strings.Builder
	offset := 0
	recipients := map[uint64]bool{}
	for _, n := range nodes {
		result.WriteString(content[offset:n.Start])
		replacement := n.Raw
		id := n.UserID
		if n.Username != "" {
			matches := byName[strings.ToLower(n.Username)]
			if len(matches) == 1 {
				id = matches[0].Id
				replacement = markdownext.Canonical(id, matches[0].Username)
			}
		}
		if id > 0 {
			if _, exists := byID[id]; exists && markdownext.NotificationsAllowed(n) {
				recipients[id] = true
			}
		}
		result.WriteString(replacement)
		offset = n.End
	}
	result.WriteString(content[offset:])
	if len(recipients) > 20 && validate {
		return "", ErrInvalid
	}
	return result.String(), nil
}

// HydrateHTML refreshes labels by stable identity without changing saved Markdown.
func HydrateHTML(raw string) string {
	return HydrateHTMLs([]string{raw})[0]
}

func HydrateHTMLs(values []string) []string {
	result := append([]string(nil), values...)
	documents := make([][]*nethtml.Node, len(values))
	var links []*nethtml.Node
	var ids []uint64
	var walk func(*nethtml.Node)
	walk = func(n *nethtml.Node) {
		if n.Type == nethtml.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "data-mention-user" {
					id, _ := strconv.ParseUint(a.Val, 10, 64)
					if id > 0 {
						ids = append(ids, id)
						links = append(links, n)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for index, raw := range values {
		if !strings.Contains(raw, "data-mention-user=") {
			continue
		}
		doc, err := nethtml.ParseFragment(strings.NewReader(raw), nil)
		if err != nil {
			continue
		}
		documents[index] = doc
		for _, node := range doc {
			walk(node)
		}
	}
	if len(ids) == 0 {
		return result
	}
	identities, err := users.MentionIdentities(ids, nil)
	if err != nil {
		return result
	}
	byID := map[uint64]string{}
	for _, user := range identities {
		byID[user.Id] = user.Username
	}
	for _, node := range links {
		var id uint64
		for _, a := range node.Attr {
			if a.Key == "data-mention-user" {
				id, _ = strconv.ParseUint(a.Val, 10, 64)
			}
		}
		if name, ok := byID[id]; ok {
			for node.FirstChild != nil {
				node.RemoveChild(node.FirstChild)
			}
			node.AppendChild(&nethtml.Node{Type: nethtml.TextNode, Data: "@" + name})
		} else {
			node.Data = "span"
			node.Attr = nil
		}
	}
	for index, doc := range documents {
		if doc == nil {
			continue
		}
		var output bytes.Buffer
		for _, node := range doc {
			if node.Data == "html" {
				var body *nethtml.Node
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					if c.Data == "body" {
						body = c
					}
				}
				if body != nil {
					for c := body.FirstChild; c != nil; c = c.NextSibling {
						_ = nethtml.Render(&output, c)
					}
				}
			} else {
				_ = nethtml.Render(&output, node)
			}
		}
		result[index] = output.String()
	}
	return result
}
func DedupeKey(postID, userID uint64) string { return fmt.Sprintf("post:%d:user:%d", postID, userID) }
