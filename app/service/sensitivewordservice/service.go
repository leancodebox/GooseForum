package sensitivewordservice

import (
	"sort"
	"strings"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
)

type Match struct {
	Word        string
	Action      string
	Replacement string
}

type Result struct {
	Passed       bool
	MatchedWords []string
	Reason       string
}

var (
	mu    sync.RWMutex
	words []sensitiveWord.Entity
)

func Refresh() {
	list := sensitiveWord.List(true)
	sort.Slice(list, func(i, j int) bool { return len(list[i].Word) > len(list[j].Word) })
	mu.Lock()
	words = list
	mu.Unlock()
}

func Check(content string) Result {
	mu.RLock()
	list := append([]sensitiveWord.Entity(nil), words...)
	mu.RUnlock()
	if len(list) == 0 {
		Refresh()
		mu.RLock()
		list = append([]sensitiveWord.Entity(nil), words...)
		mu.RUnlock()
	}
	result := Result{Passed: true}
	seen := make(map[string]struct{})
	for _, word := range list {
		if word.Word == "" || !strings.Contains(content, word.Word) {
			continue
		}
		if _, ok := seen[word.Word]; ok {
			continue
		}
		seen[word.Word] = struct{}{}
		result.MatchedWords = append(result.MatchedWords, word.Word)
		if word.Action == sensitiveWord.ActionReject || word.Action == "" {
			result.Passed = false
		}
	}
	if len(result.MatchedWords) > 0 {
		result.Reason = "matched sensitive words: " + strings.Join(result.MatchedWords, ", ")
	}
	return result
}

func Replace(content string) string {
	mu.RLock()
	list := append([]sensitiveWord.Entity(nil), words...)
	mu.RUnlock()
	for _, word := range list {
		if word.Action == sensitiveWord.ActionReplace && word.Word != "" {
			content = strings.ReplaceAll(content, word.Word, word.Replacement)
		}
	}
	return content
}
