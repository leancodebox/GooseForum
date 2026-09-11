package api

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/sensitiveWord"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
)

func TestSensitiveWordEditPersistsOnlySelectedRule(t *testing.T) {
	db := setupAdminTopicTestDB(t)
	if err := db.AutoMigrate(&sensitiveWord.Entity{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sensitivewordservice.Refresh() })
	save := func(word sensitiveWord.Entity) sensitiveWord.Entity {
		t.Helper()
		response := SaveSensitiveWord(component.BetterRequest[SensitiveWordSaveReq]{Params: SensitiveWordSaveReq{Word: word}})
		result, ok := response.Data.Result.(component.DataMap)
		if !ok {
			t.Fatalf("save failed: %+v", response)
		}
		saved, ok := result["word"].(sensitiveWord.Entity)
		if !ok {
			t.Fatalf("missing saved rule: %+v", response)
		}
		return saved
	}
	first := save(sensitiveWord.Entity{Word: "edit-test", Action: sensitiveWord.ActionReject, Enabled: true})
	other := save(sensitiveWord.Entity{Word: "unchanged-test", Action: sensitiveWord.ActionReject, Enabled: true})
	first.Word = "edited-test"
	first.Action = sensitiveWord.ActionReplace
	first.Replacement = "replacement"
	first = save(first)
	stored, err := sensitiveWord.Get(first.Id)
	if err != nil || stored.Word != first.Word || stored.Action != first.Action || stored.Replacement != "replacement" {
		t.Fatalf("edit not persisted: %+v, %v", stored, err)
	}
	first.Replacement = ""
	first.Enabled = false
	save(first)
	stored, err = sensitiveWord.Get(first.Id)
	if err != nil || stored.Enabled || stored.Replacement != "" {
		t.Fatalf("zero values not persisted: %+v, %v", stored, err)
	}
	untouched, err := sensitiveWord.Get(other.Id)
	if err != nil || untouched.Word != other.Word || untouched.Action != other.Action || !untouched.Enabled {
		t.Fatalf("another rule changed: %+v, %v", untouched, err)
	}
}
