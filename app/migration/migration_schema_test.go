package migration

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationUsesCleanTopicEdgeModels(t *testing.T) {
	source, err := os.ReadFile("migration.go")
	if err != nil {
		t.Fatalf("read migration.go: %v", err)
	}

	text := string(source)
	for _, oldModel := range []string{
		"models/forum/articleCategory",
		"models/forum/articleCategoryRs",
		"models/forum/articleUserAction",
		"models/forum/articlesUserStat",
		"models/forum/articles",
		"models/forum/reply",
	} {
		if strings.Contains(text, oldModel) {
			t.Fatalf("migration still imports old edge model %q", oldModel)
		}
	}

	for _, cleanModel := range []string{
		"models/forum/category",
		"models/forum/migrationMapping",
		"models/forum/topicCategoryIndex",
		"models/forum/topicUserAction",
		"models/forum/topicUserStat",
	} {
		if !strings.Contains(text, cleanModel) {
			t.Fatalf("migration does not import clean edge model %q", cleanModel)
		}
	}
}
