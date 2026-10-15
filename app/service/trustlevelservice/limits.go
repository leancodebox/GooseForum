package trustlevelservice

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	nethtml "golang.org/x/net/html"
)

type LimitError struct {
	Kind  string
	Limit int
}

const DailyTopicsKind = "每日新建主题数"
const DailyRepliesKind = "每日新建回复数"

func ValidateNewReply(id uint64) error {
	l, enabled, err := Limits(id)
	if err != nil {
		return err
	}
	return validateReplyLimit(id, l, enabled)
}

func validateReplyLimit(id uint64, l pageConfig.TrustLimits, enabled bool) error {
	if !enabled || l.MaxDailyReplies == nil {
		return nil
	}
	if *l.MaxDailyReplies == 0 {
		return &LimitError{DailyRepliesKind, 0}
	}
	count, err := posts.DailyReplyCount(id)
	if err != nil {
		return err
	}
	if count >= int64(*l.MaxDailyReplies) {
		return &LimitError{DailyRepliesKind, *l.MaxDailyReplies}
	}
	return nil
}

var ErrGlobalDailyTopics = errors.New("global daily topic limit reached")

func ValidateNewTopic(id uint64, globalLimit int) error {
	l, enabled, err := Limits(id)
	if err != nil {
		return err
	}
	var levelLimit *int
	if enabled {
		levelLimit = l.MaxDailyTopics
	}
	if levelLimit != nil && *levelLimit == 0 {
		return &LimitError{DailyTopicsKind, 0}
	}
	effective := globalLimit
	if levelLimit != nil && (effective <= 0 || *levelLimit < effective) {
		effective = *levelLimit
	}
	if effective <= 0 {
		return nil
	}
	count, err := topics.DailyCreatedCount(id)
	if err != nil {
		return err
	}
	if count >= int64(effective) {
		if levelLimit != nil && (globalLimit <= 0 || *levelLimit <= globalLimit) {
			return &LimitError{DailyTopicsKind, *levelLimit}
		}
		return ErrGlobalDailyTopics
	}
	return nil
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("当前等级的%s上限为 %d", e.Kind, e.Limit)
}

func Limits(id uint64) (pageConfig.TrustLimits, bool, error) {
	c, err := hotdataserve.GetTrustSettings()
	if err != nil {
		return pageConfig.TrustLimits{}, false, err
	}
	if !c.RestrictionsEnabled {
		return pageConfig.TrustLimits{}, false, nil
	}
	s, err := users.GetAccountState(id)
	if err != nil {
		return pageConfig.TrustLimits{}, false, err
	}
	if !trustlevel.Valid(s.TrustLevel) {
		return pageConfig.TrustLimits{}, false, ErrInvalid
	}
	for _, l := range c.LevelLimits {
		if l.Level == s.TrustLevel {
			return l, true, nil
		}
	}
	return pageConfig.TrustLimits{}, false, ErrInvalid
}

type Counts struct {
	Images, Links, Mentions int
	mentionIDs              []uint64
}

func CountContent(content string, version uint8) Counts {
	// Count the same normalized HTML that the forum renders, including extensions.
	raw := markdown2html.PostMarkdownToHTMLVersion(content, version)
	return countHTML(raw)
}

func countHTML(raw string) Counts {
	root, err := nethtml.Parse(strings.NewReader(raw))
	if err != nil {
		return Counts{}
	}
	result := Counts{}
	mentions := map[uint64]bool{}
	var walk func(*nethtml.Node)
	walk = func(n *nethtml.Node) {
		if n.Type == nethtml.ElementNode {
			if n.Data == "img" {
				result.Images++
			}
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "rel" && strings.Contains(a.Val, "ugc") {
						result.Links++
					}
					if a.Key == "data-mention-user" {
						if id, err := strconv.ParseUint(a.Val, 10, 64); err == nil && id > 0 {
							mentions[id] = true
						}
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	result.Mentions = len(mentions)
	for id := range mentions {
		result.mentionIDs = append(result.mentionIDs, id)
	}
	return result
}

func ValidateContent(id uint64, content string, version uint8) error {
	l, enabled, err := Limits(id)
	if err != nil || !enabled {
		return err
	}
	if !hasContentLimits(l) {
		return nil
	}
	return validateCounts(l, CountContent(content, version))
}

func hasContentLimits(l pageConfig.TrustLimits) bool {
	return l.MaxImagesPerPost != nil || l.MaxExternalLinksPerPost != nil || l.MaxMentionsPerPost != nil
}

// AnalyzePost loads policy once and reuses the HTML needed by the write itself.
func AnalyzePost(id uint64, content string, version uint8, creating bool) (markdown2html.ContentAnalysis, error) {
	l, enabled, err := Limits(id)
	if err != nil {
		return markdown2html.ContentAnalysis{}, err
	}
	if creating {
		if err := validateReplyLimit(id, l, enabled); err != nil {
			return markdown2html.ContentAnalysis{}, err
		}
	}
	analysis := markdown2html.AnalyzePostContentVersion(content, version)
	if enabled && hasContentLimits(l) {
		if err := validateCounts(l, countHTML(analysis.RenderedHTML)); err != nil {
			return markdown2html.ContentAnalysis{}, err
		}
	}
	return analysis, nil
}

func validateCounts(l pageConfig.TrustLimits, n Counts) error {
	if l.MaxMentionsPerPost != nil && len(n.mentionIDs) > 0 {
		identities, err := users.MentionIdentities(n.mentionIDs, nil)
		if err != nil {
			return err
		}
		n.Mentions = len(identities)
	}
	for _, v := range []struct {
		kind  string
		count int
		limit *int
	}{{"图片数量", n.Images, l.MaxImagesPerPost}, {"外链数量", n.Links, l.MaxExternalLinksPerPost}, {"提及人数", n.Mentions, l.MaxMentionsPerPost}} {
		if v.limit != nil && v.count > *v.limit {
			return &LimitError{v.kind, *v.limit}
		}
	}
	return nil
}
