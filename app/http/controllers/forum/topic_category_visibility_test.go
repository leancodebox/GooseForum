package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

// A public main category may carry restricted auxiliary categories, so the topic
// detail payload must not disclose the auxiliary ones to a viewer who cannot read
// them: their name and URL would render on the page, and buildTopicMeta would
// publish the name as an SEO keyword and an article tag.
func TestReadableCategoryPayloadsHidesRestrictedAuxiliaryCategories(t *testing.T) {
	const (
		publicCategoryID     = uint64(994301)
		restrictedCategoryID = uint64(994302)
		restrictedName       = "Restricted auxiliary"
	)

	ensureForumTestAccessCategory(t, publicCategoryID)
	ensureForumTestAccessCategory(t, restrictedCategoryID)
	renameForumTestCategory(t, publicCategoryID, "Public main")
	renameForumTestCategory(t, restrictedCategoryID, restrictedName)
	revokeForumTestEveryoneRead(t, restrictedCategoryID)

	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatalf("resolve guest snapshot: %v", err)
	}
	if guest.CanReadCategory(restrictedCategoryID) {
		t.Fatal("test setup: guests must not be able to read the restricted category")
	}

	payloads := readableCategoryPayloads(guest, []uint64{publicCategoryID, restrictedCategoryID})

	if len(payloads) != 1 {
		t.Fatalf("expected only the readable category, got %#v", payloads)
	}
	if payloads[0].ID != publicCategoryID {
		t.Fatalf("expected the public main category, got %#v", payloads[0])
	}
	for _, payload := range payloads {
		if payload.Name == restrictedName {
			t.Fatalf("restricted auxiliary category name leaked to a guest: %#v", payload)
		}
	}
}

// The main category is exempt from the filter. The handler requires read on it
// before rendering the page (forum.Topic), so it is readable by construction, and
// dropping it would empty the breadcrumb and the page metadata.
func TestReadableCategoryPayloadsKeepsMainCategory(t *testing.T) {
	const (
		restrictedCategoryID = uint64(994311)
		publicCategoryID     = uint64(994312)
	)

	ensureForumTestAccessCategory(t, restrictedCategoryID)
	ensureForumTestAccessCategory(t, publicCategoryID)
	renameForumTestCategory(t, restrictedCategoryID, "Restricted main")
	renameForumTestCategory(t, publicCategoryID, "Public auxiliary")
	revokeForumTestEveryoneRead(t, restrictedCategoryID)

	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatalf("resolve guest snapshot: %v", err)
	}

	payloads := readableCategoryPayloads(guest, []uint64{restrictedCategoryID, publicCategoryID})

	if len(payloads) != 2 || payloads[0].ID != restrictedCategoryID {
		t.Fatalf("expected the main category to be kept first, got %#v", payloads)
	}
}

// A list only contains topics whose main category the viewer can read, but those
// topics may still carry restricted auxiliary categories, so the cards need the
// same filter as the detail page.
func TestApplyCategoryVisibilityHidesRestrictedAuxiliaryCategoriesOnCards(t *testing.T) {
	const (
		publicCategoryID     = uint64(994321)
		restrictedCategoryID = uint64(994322)
	)

	ensureForumTestAccessCategory(t, publicCategoryID)
	ensureForumTestAccessCategory(t, restrictedCategoryID)
	revokeForumTestEveryoneRead(t, restrictedCategoryID)

	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		t.Fatalf("resolve guest snapshot: %v", err)
	}

	payloads := applyCategoryVisibility(guest, []TopicPayload{{
		ID: 1,
		Categories: []TopicCategoryPayload{
			{ID: publicCategoryID, Name: "Public main"},
			{ID: restrictedCategoryID, Name: "Restricted auxiliary"},
		},
	}})

	if len(payloads) != 1 {
		t.Fatalf("expected the card to survive, got %#v", payloads)
	}
	if len(payloads[0].Categories) != 1 || payloads[0].Categories[0].ID != publicCategoryID {
		t.Fatalf("restricted auxiliary category leaked on a list card: %#v", payloads[0].Categories)
	}
}

func renameForumTestCategory(t *testing.T, categoryID uint64, name string) {
	t.Helper()
	conn := dbconnect.Connect()
	if err := conn.Model(&category.Entity{}).Where("id = ?", categoryID).Update("name", name).Error; err != nil {
		t.Fatalf("rename test category: %v", err)
	}
	hotdataserve.ClearCategoryCache()
	t.Cleanup(hotdataserve.ClearCategoryCache)
}
