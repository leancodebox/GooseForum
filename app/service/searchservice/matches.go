package searchservice

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
)

type SearchMatch struct {
	PostID  uint64        `json:"postId"`
	PostNo  uint64        `json:"postNo"`
	Excerpt []ExcerptPart `json:"excerpt"`
}

type ExcerptPart struct {
	Text    string `json:"text"`
	Matched bool   `json:"matched"`
}

type matchPosition struct {
	Start  int `json:"start"`
	Length int `json:"length"`
}

type searchHit struct {
	ID              uint64                     `json:"id"`
	Title           string                     `json:"title"`
	SearchContent   string                     `json:"searchContent"`
	SearchPosts     []SearchPost               `json:"searchPosts"`
	MatchesPosition map[string][]matchPosition `json:"_matchesPosition"`
}

func (hit searchHit) matchedPost() (SearchPost, bool) {
	for _, source := range hit.SearchPosts {
		for _, position := range hit.MatchesPosition["searchContent"] {
			if position.Start >= source.Start && position.Length > 0 && position.Start+position.Length <= source.Start+source.Length {
				return source, source.PostID > 0
			}
		}
	}
	return SearchPost{}, false
}

func buildSearchResults(hits []searchHit, includeMatches bool) ([]SearchResult, error) {
	results := make([]SearchResult, 0, len(hits))
	var ids []uint64
	if includeMatches {
		for _, hit := range hits {
			if source, ok := hit.matchedPost(); ok {
				ids = append(ids, source.PostID)
			}
		}
	}
	rows, err := posts.GetSearchPostReferences(ids)
	if err != nil {
		return nil, fmt.Errorf("read matched search posts: %w", err)
	}
	currentPosts := make(map[uint64]posts.SearchReference, len(rows))
	for _, row := range rows {
		currentPosts[row.Id] = row
	}
	for _, hit := range hits {
		result := SearchResult{ID: hit.ID, Title: hit.Title}
		if source, ok := hit.matchedPost(); ok && includeMatches {
			// An index task may lag behind edits, moderation or deletion. Only
			// expose an excerpt when the source version is still publicly visible.
			if post, exists := currentPosts[source.PostID]; exists && post.TopicId == hit.ID && post.PostNo == source.PostNo && post.ModerationVersion == source.Version && post.UpdatedAt.UnixMilli() == source.UpdatedAt {
				excerpt := buildExcerpt(hit.SearchContent, source, hit.MatchesPosition["searchContent"])
				if len(excerpt) > 0 {
					result.Match = &SearchMatch{PostID: source.PostID, PostNo: source.PostNo, Excerpt: excerpt}
				}
			}
		}
		results = append(results, result)
	}
	return results, nil
}

func buildExcerpt(content string, source SearchPost, positions []matchPosition) []ExcerptPart {
	end := source.Start + source.Length
	if source.Start < 0 || source.Length <= 0 || end > len(content) {
		return nil
	}
	var matches []matchPosition
	for _, position := range positions {
		if position.Start >= source.Start && position.Length > 0 && position.Start+position.Length <= end {
			matches = append(matches, position)
		}
	}
	if len(matches) == 0 {
		return nil
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Start < matches[j].Start })
	start := matches[0].Start
	for n := 0; start > source.Start && n < 20; n++ {
		_, size := utf8.DecodeLastRuneInString(content[source.Start:start])
		start -= size
	}
	stop := matches[0].Start + matches[0].Length
	for n := 0; stop < end && n < 60; n++ {
		_, size := utf8.DecodeRuneInString(content[stop:end])
		stop += size
	}
	parts := make([]ExcerptPart, 0, len(matches)*2+3)
	if start > source.Start {
		parts = append(parts, ExcerptPart{Text: "..."})
	}
	cursor := start
	for _, position := range matches {
		matchEnd := position.Start + position.Length
		if position.Start < cursor || matchEnd > stop {
			continue
		}
		if position.Start > cursor {
			parts = append(parts, ExcerptPart{Text: content[cursor:position.Start]})
		}
		parts = append(parts, ExcerptPart{Text: content[position.Start:matchEnd], Matched: true})
		cursor = matchEnd
	}
	if cursor < stop {
		parts = append(parts, ExcerptPart{Text: content[cursor:stop]})
	}
	if stop < end {
		parts = append(parts, ExcerptPart{Text: "..."})
	}
	return parts
}
