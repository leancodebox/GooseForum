package sensitivewordservice

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
	"strings"
	"testing"
)

func TestReplacementDoesNotHideRejectOrCascade(t *testing.T) {
	words := []sensitiveWord.Entity{
		{Word: "bad", Action: sensitiveWord.ActionReject},
		{Word: "bad", Action: sensitiveWord.ActionReplace, Replacement: "safe"},
		{Word: "safe", Action: sensitiveWord.ActionReplace, Replacement: "changed"},
	}
	result := check("bad safe", words)
	if result.Passed || result.Content != "safe changed" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRecordAndUnicodeReplacement(t *testing.T) {
	result := check("敏感词 记录", []sensitiveWord.Entity{
		{Word: "敏感词", Action: sensitiveWord.ActionReplace, Replacement: "***"},
		{Word: "记录", Action: sensitiveWord.ActionRecord},
	})
	if !result.Passed || result.Content != "*** 记录" || len(result.MatchedWords) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func BenchmarkDictionary(b *testing.B) {
	words := make([]sensitiveWord.Entity, 10000)
	for i := range words {
		words[i] = sensitiveWord.Entity{Word: fmt.Sprintf("blocked-%05d", i), Action: sensitiveWord.ActionReject}
	}
	content := strings.Repeat("ordinary forum content ", 2200)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check(content, words)
	}
}
