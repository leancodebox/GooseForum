package hotdataserve

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestTopicListCacheReadsTopics(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &category.Entity{}, &topicCategoryIndex.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate topic list cache tables: %v", err)
	}
	conn.Where("1 = 1").Delete(&topics.Entity{})
	conn.Where("1 = 1").Delete(&category.Entity{})
	conn.Where("1 = 1").Delete(&topicCategoryIndex.Entity{})
	conn.Where("1 = 1").Delete(&users.EntityComplete{})
	ClearArticleCategoryCache()
	ClearTopicListCache()

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	conn.Create(&users.EntityComplete{Id: 1, Username: "author"})
	conn.Create(&category.Entity{Id: 3, Name: "General", Slug: "general"})
	conn.Create(&topics.Entity{
		Id:            10,
		Title:         "topic title",
		Excerpt:       "topic excerpt",
		FirstImageURL: "/a.png",
		CategoryIds:   []uint64{3},
		UserId:        1,
		Status:        1,
		ProcessStatus: 0,
		ReplyCount:    2,
		ViewCount:     9,
		PinWeight:     7,
		Posters:       []topics.Poster{{UserID: 1}},
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	conn.Create(&topicCategoryIndex.Entity{TopicId: 10, CategoryId: 3, Effective: 1})

	page := GetLatestTopicsSimpleVoPaginated(1, "latest")
	if len(page.Topics) != 1 {
		t.Fatalf("latest topics len=%d, want 1", len(page.Topics))
	}
	item := page.Topics[0]
	if item.Id != 10 || item.Title != "topic title" || item.Description != "topic excerpt" || item.Categories[0] != "General" {
		t.Fatalf("topic list item mismatch: %#v", item)
	}

	categoryPage := GetTopicsByCategorySimpleVo(3, "latest", 1)
	if len(categoryPage.Topics) != 1 || categoryPage.Topics[0].Id != 10 {
		t.Fatalf("category topic page=%#v, want topic 10", categoryPage.Topics)
	}
}

func TestCategoryCacheReadsCleanCategories(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&category.Entity{}); err != nil {
		t.Fatalf("migrate category cache tables: %v", err)
	}
	conn.Unscoped().Delete(&category.Entity{}, 980003)
	ClearCategoryCache()

	conn.Create(&category.Entity{Id: 980003, Name: "Clean", Slug: "clean", Sort: 9})
	categories := GetCategory()
	if len(categories) == 0 {
		t.Fatal("categories empty, want clean category")
	}
	if got := CategoryMap()[980003]; got == nil || got.Name != "Clean" {
		t.Fatalf("category map item=%#v, want Clean", got)
	}
}

func TestSiteStatsReadsTopicPostMaxIds(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &users.EntityComplete{}, &dailyStats.Entity{}, &pageConfig.Entity{}); err != nil {
		t.Fatalf("migrate site stats tables: %v", err)
	}
	conn.Unscoped().Delete(&topics.Entity{}, 990010)
	conn.Unscoped().Delete(&posts.Entity{}, 990020)
	conn.Unscoped().Delete(&users.EntityComplete{}, 990001)
	siteStatisticsDataCache.Clear()

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	conn.Create(&users.EntityComplete{Id: 990001, Username: "stat-user"})
	conn.Create(&topics.Entity{Id: 990010, Title: "stats topic", UserId: 990001, Status: 1, CreatedAt: now, UpdatedAt: now})
	conn.Create(&posts.Entity{Id: 990020, TopicId: 990010, PostNo: 1, UserId: 990001, CreatedAt: now, UpdatedAt: now})

	stats := GetSiteStatisticsData()
	if stats.ArticleCount != 990010 || stats.Reply != 990020 {
		t.Fatalf("stats article=%d reply=%d, want topic/post max ids", stats.ArticleCount, stats.Reply)
	}
}
