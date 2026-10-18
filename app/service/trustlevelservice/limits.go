package trustlevelservice

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"

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
	p, err := ResolvePolicy(id)
	if err != nil {
		return err
	}
	if err := p.ValidatePostAge(); err != nil {
		return err
	}
	return validateReplyLimit(id, p.Limits, true)
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

func ValidateNewTopic(id uint64) error {
	p, err := ResolvePolicy(id)
	if err != nil {
		return err
	}
	if err := p.ValidatePostAge(); err != nil {
		return err
	}
	limit := p.Limits.MaxDailyTopics
	if limit == nil {
		return nil
	}
	if *limit == 0 {
		return &LimitError{DailyTopicsKind, 0}
	}
	count, err := topics.DailyCreatedCount(id)
	if err != nil {
		return err
	}
	if count >= int64(*limit) {
		return &LimitError{DailyTopicsKind, *limit}
	}
	return nil
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("当前等级的%s上限为 %d", e.Kind, e.Limit)
}

type UserPolicy struct {
	Account users.AccountState
	Limits  pageConfig.TrustLimits
}

type PolicyError struct {
	Status int
	Code   string
	Params map[string]any
}

func (e *PolicyError) Error() string { return e.Code }

func ResolvePolicy(id uint64) (UserPolicy, error) {
	c, err := hotdataserve.GetTrustSettings()
	if err != nil {
		return UserPolicy{}, err
	}
	account, err := users.GetAccountState(id)
	if err != nil {
		return UserPolicy{}, err
	}
	if !trustlevel.Valid(account.TrustLevel) {
		return UserPolicy{}, ErrInvalid
	}
	for _, l := range c.LevelLimits {
		if l.Level == account.TrustLevel {
			return UserPolicy{account, l}, nil
		}
	}
	return UserPolicy{}, ErrInvalid
}
func Limits(id uint64) (pageConfig.TrustLimits, bool, error) {
	p, err := ResolvePolicy(id)
	return p.Limits, err == nil, err
}

func (p UserPolicy) ValidatePostAge() error {
	minutes := p.Limits.MinPostAccountAgeMinutes
	available := p.Account.CreatedAt.Add(time.Duration(minutes) * time.Minute)
	if minutes > 0 && time.Now().Before(available) {
		return &PolicyError{429, "topic.post.cooldown", map[string]any{"minutes": minutes, "availableAt": available.Format(time.DateTime)}}
	}
	return nil
}

func (p UserPolicy) ValidateUpload(fileCount int, avatar bool) error {
	return p.validateUpload(fileCount, avatar, fileCount > 0)
}

// Pending direct uploads are already included in the daily file count.
func (p UserPolicy) ValidateExistingUpload() error { return p.validateUpload(0, false, true) }

func (p UserPolicy) validateUpload(fileCount int, avatar bool, checkQuota bool) error {
	if !avatar && !p.Limits.AllowImageUpload {
		return &PolicyError{403, "permission.denied", nil}
	}
	// Existing role exemptions apply to community uploads, not avatars.
	if !avatar && p.Account.RoleId > 0 {
		return nil
	}
	minutes := p.Limits.MinUploadAccountAgeMinutes
	available := p.Account.CreatedAt.Add(time.Duration(minutes) * time.Minute)
	if minutes > 0 && time.Now().Before(available) {
		return &PolicyError{400, "upload.cooldown", map[string]any{"minutes": minutes, "availableAt": available.Format(time.DateTime)}}
	}
	if !checkQuota || p.Limits.MaxDailyUploads == nil {
		return nil
	}
	count := int64(0)
	if *p.Limits.MaxDailyUploads > 0 {
		var err error
		count, err = filedata.DailyUploadCount(p.Account.Id)
		if err != nil {
			return err
		}
	}
	if *p.Limits.MaxDailyUploads == 0 || count+int64(fileCount) > int64(*p.Limits.MaxDailyUploads) {
		code := "upload.dailyLimit"
		if avatar {
			code = "upload.dailyLimit.avatar"
		}
		return &PolicyError{400, code, map[string]any{"count": count, "fileCount": fileCount}}
	}
	return nil
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
	p, err := ResolvePolicy(id)
	if err != nil {
		return markdown2html.ContentAnalysis{}, err
	}
	l, enabled := p.Limits, true
	if creating {
		if err := p.ValidatePostAge(); err != nil {
			return markdown2html.ContentAnalysis{}, err
		}
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
