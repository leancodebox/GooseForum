package api

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
)

func TestPostWriteResponseKeepsSharedContract(t *testing.T) {
	entity := posts.Entity{
		Id:                41,
		PostNo:            3,
		Content:           "reply",
		SourceVersion:     1,
		ModerationVersion: 7,
		RenderedHTML:      "<p>reply</p>",
		ModerationStatus:  "approved",
		ProcessStatus:     0,
	}

	got := postWriteResponse(entity)
	if got["id"] != uint64(41) || got["postNo"] != uint64(3) || got["content"] != "reply" || got["sourceVersion"] != uint8(1) {
		t.Fatalf("identity fields = %#v", got)
	}
	if got["version"] != uint64(7) || got["renderedContent"] != "<p>reply</p>" || got["moderationStatus"] != "approved" || got["processStatus"] != int8(0) {
		t.Fatalf("moderation fields = %#v", got)
	}
	if got["updatedAt"] != nil {
		t.Fatalf("shared response unexpectedly contains updatedAt: %#v", got)
	}
}
