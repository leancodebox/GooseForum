package forum

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"gorm.io/gorm"
)

func TestProfilePrivacyBlocksDirectURLsWithoutExtraQueries(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&users.EntityComplete{}, &topics.Entity{}, &userActivities.Entity{}, &userFollow.Entity{}, &userStatistics.Entity{}); err != nil {
		t.Fatal(err)
	}
	user := users.EntityComplete{Username: "privacy-profile"}
	if err := users.Create(&user); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&user); db.Delete(&userActivities.Entity{}, "user_id = ?", user.Id) })
	if err := userActivities.Record(user.Id, userActivities.ActionFollow, userActivities.SubjectUser, 123, "follow"); err != nil {
		t.Fatal(err)
	}
	ensureForumTestAccessCategory(t, 989111)
	snapshot, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatal(err)
	}
	topic := topics.Entity{Title: "profile privacy topic", UserId: user.Id, MainCategoryId: 989111, CategoryIds: []uint64{989111}, Status: 1, CreatedAt: time.Now()}
	if err := db.Create(&topic).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&topic) })
	if err := userActivities.RecordForTopic(user.Id, userActivities.ActionPost, userActivities.SubjectTopic, topic.Id, topic.Id, "topic"); err != nil {
		t.Fatal(err)
	}
	userservice.RefreshUserCaches(&user)
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest("GET", "/u/1", nil)
	queries := 0
	callback := "test:profile-privacy-query-count"
	if err := db.Callback().Query().After("gorm:query").Register(callback, func(*gorm.DB) { queries++ }); err != nil {
		t.Fatal(err)
	}
	defer db.Callback().Query().Remove(callback)
	build := func(viewer uint64, section, tab string) (UserProfileProps, int) {
		c.Set("userId", viewer)
		queries = 0
		result := buildUserProfileProps(c, snapshot, user, section, tab)
		return result, queries
	}
	// Warm existing user/category caches before comparing request query counts.
	build(0, userProfileSectionSummary, "")
	public, publicQueries := build(0, userProfileSectionSummary, "")
	if len(public.Activities) != 2 || len(public.Topics) != 1 {
		t.Fatalf("public activity = %+v", public.Activities)
	}
	user.HideTopics, user.HideFollowing = true, true
	private, privateQueries := build(0, userProfileSectionSummary, "")
	if len(private.Topics) != 0 || len(private.Activities) != 0 {
		t.Fatal("hidden content leaked in summary")
	}
	if privateQueries > publicQueries {
		t.Fatalf("queries increased: public=%d private=%d", publicQueries, privateQueries)
	}
	for _, tab := range []string{userProfileActivityTopics, userProfileActivityFollowing, userProfileActivityFollowers} {
		hidden, hiddenQueries := build(0, userProfileSectionActivity, tab)
		if len(hidden.Topics)+len(hidden.Following)+len(hidden.Followers) != 0 {
			t.Fatalf("direct URL leaked %s", tab)
		}
		user.HideTopics, user.HideFollowing = false, false
		_, visibleQueries := build(0, userProfileSectionActivity, tab)
		user.HideTopics, user.HideFollowing = true, true
		if hiddenQueries > visibleQueries {
			t.Fatalf("%s queries increased", tab)
		}
	}
	// Hiding the whole timeline also hides likes/replies, while topics can remain public.
	user.HideTopics, user.HideFollowing = false, false
	user.HideActivity = true
	hiddenSummary, hiddenSummaryQueries := build(0, userProfileSectionSummary, "")
	if len(hiddenSummary.Activities) != 0 || len(hiddenSummary.Topics) != 1 || hiddenSummaryQueries > publicQueries {
		t.Fatal("timeline privacy did not apply to summary")
	}
	hiddenTimeline, _ := build(0, userProfileSectionActivity, userProfileActivityTimeline)
	if len(hiddenTimeline.Activities) != 0 || hiddenTimeline.Pagination.HasNext {
		t.Fatal("direct timeline URL bypasses privacy")
	}
	hiddenLikes, _ := build(0, userProfileSectionActivity, userProfileActivityLikes)
	if len(hiddenLikes.Likes) != 0 || hiddenLikes.Pagination.HasNext {
		t.Fatal("likes URL bypasses activity privacy")
	}
	ownTimeline, _ := build(user.Id, userProfileSectionActivity, userProfileActivityTimeline)
	if len(ownTimeline.Activities) != 2 {
		t.Fatal("owner cannot see hidden timeline")
	}
	own, _ := build(user.Id, userProfileSectionSummary, "")
	if len(own.Activities) != 2 || len(own.Topics) != 1 {
		t.Fatal("owner cannot see own follow activity")
	}
}
