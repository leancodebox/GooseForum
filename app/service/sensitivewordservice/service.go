package sensitivewordservice

import (
	"sort"
	"strings"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
)

type Result struct {
	Passed       bool
	Content      string
	MatchedWords []string
	Reason       string
}

// Published dictionaries are immutable; readers share the slice without copying it.
var dictionary struct {
	sync.Mutex
	loaded bool
	words  []sensitiveWord.Entity
}

func Refresh() error {
	dictionary.Lock()
	defer dictionary.Unlock()
	list, err := loadWords()
	dictionary.loaded = err == nil
	if err != nil {
		return err
	}
	dictionary.words = list
	return nil
}

func loadWords() ([]sensitiveWord.Entity, error) {
	list, err := sensitiveWord.LoadEnabled()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(list, func(i, j int) bool { return len(list[i].Word) > len(list[j].Word) })
	return list, nil
}

func Check(content string) Result {
	dictionary.Lock()
	if !dictionary.loaded {
		list, err := loadWords()
		if err != nil {
			dictionary.Unlock()
			return Result{Passed: false, Content: content, Reason: "sensitive word dictionary unavailable"}
		}
		dictionary.words = list
		dictionary.loaded = true
	}
	list := dictionary.words
	dictionary.Unlock()
	return check(content, list)
}

func check(content string, list []sensitiveWord.Entity) Result {
	result := Result{Passed: true, Content: content}
	replacements := make([]string, 0)
	for _, word := range list {
		if word.Word == "" || !strings.Contains(content, word.Word) {
			continue
		}
		result.MatchedWords = append(result.MatchedWords, word.Word)
		switch word.Action {
		case sensitiveWord.ActionReplace:
			replacements = append(replacements, word.Word, word.Replacement)
		case sensitiveWord.ActionRecord:
		default:
			result.Passed = false
		}
	}
	// One pass prevents replacement text from being processed by another rule.
	if len(replacements) > 0 {
		result.Content = strings.NewReplacer(replacements...).Replace(content)
	}
	if len(result.MatchedWords) > 0 {
		result.Reason = "matched sensitive words: " + strings.Join(result.MatchedWords, ", ")
	}
	return result
}
