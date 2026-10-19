package topicwriteservice

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"gorm.io/gorm"
)

func TestTopicPreparationPolicyQueryBudget(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &category.Entity{}, &pageConfig.Entity{}, &topics.Entity{}, &posts.Entity{}); err != nil {
		t.Fatal(err)
	}
	u := users.EntityComplete{Username: "topic-policy-budget", UsernameLower: "topic-policy-budget", CreatedAt: time.Now().Add(-72 * time.Hour)}
	cat := category.Entity{Name: "Topic policy budget"}
	for _, row := range []any{&u, &cat} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { db.Unscoped().Delete(&u); db.Delete(&cat) })
	old := pageConfig.GetByPageType(pageConfig.TrustSettings)
	cfg := pageConfig.DefaultTrustSettings()
	raw, _ := json.Marshal(cfg)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	t.Cleanup(func() {
		if old.Id > 0 {
			_ = pageConfig.SaveConfig(pageConfig.TrustSettings, old.Config)
		} else {
			db.Where("page_type = ?", pageConfig.TrustSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearTrustSettingsCache()
	})
	previous := accesscontrol.Default
	accesscontrol.Default = accesscontrol.NewResolver(legacyTopicAccessStore{cat.Id}, nil, nil)
	t.Cleanup(func() { accesscontrol.Default = previous })
	var reads atomic.Int32
	callback := "topic-preparation-policy-budget"
	if err := db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "users" {
			reads.Add(1)
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove(callback) })
	for _, tc := range []struct {
		name      string
		existing  bool
		published bool
		status    int8
	}{
		{"new-published", false, false, 1}, {"new-draft", false, false, 0}, {"publish-draft", true, false, 1}, {"edit-published", true, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := WriteInput{UserID: u.Id, Title: "Budget topic", Content: "Description ![image](/budget.png)", CategoryIDs: []uint64{cat.Id}, Status: tc.status}
			if tc.existing {
				topic := topics.Entity{UserId: u.Id, Title: input.Title, CategoryIds: input.CategoryIDs, MainCategoryId: cat.Id}
				if tc.published {
					topic.Status = 1
					at := time.Now()
					topic.PublishedAt = &at
				}
				if err := db.Create(&topic).Error; err != nil {
					t.Fatal(err)
				}
				post := posts.Entity{UserId: u.Id, TopicId: topic.Id, PostNo: 1, Content: "Before edit"}
				if err := db.Create(&post).Error; err != nil {
					t.Fatal(err)
				}
				if err := db.Model(&topic).Update("first_post_id", post.Id).Error; err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { db.Unscoped().Delete(&post); db.Unscoped().Delete(&topic) })
				input.TopicID = topic.Id
			}
			reads.Store(0)
			state, err := loadWriteState(input)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := prepareWrite(state, input); err != nil {
				t.Fatal(err)
			}
			if reads.Load() != 1 {
				t.Fatalf("policy reads=%d; want 1", reads.Load())
			}
			if state.topic.Excerpt == "" || state.topic.FirstImageURL != "/budget.png" {
				t.Fatalf("metadata lost: %+v", state.topic)
			}
			if state.firstPost.RenderedHTML == "" {
				t.Fatalf("prepared HTML was not retained: %q", state.firstPost.RenderedHTML)
			}
		})
	}
}
