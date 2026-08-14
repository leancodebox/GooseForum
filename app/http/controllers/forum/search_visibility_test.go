package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestFilterCurrentSearchTopicsRejectsStaleDocuments(t *testing.T) {
	readableCategoryIDs := map[uint64]struct{}{10: {}}
	topicMap := map[uint64]*topics.Entity{
		1: {Id: 1, Status: 1, ProcessStatus: 0, MainCategoryId: 10},
		2: {Id: 2, Status: 0, ProcessStatus: 0, MainCategoryId: 10},
		3: {Id: 3, Status: 1, ProcessStatus: 1, MainCategoryId: 10},
		4: {Id: 4, Status: 1, ProcessStatus: 0, MainCategoryId: 20},
	}

	got := filterCurrentSearchTopics([]uint64{1, 2, 3, 4}, topicMap, readableCategoryIDs, false)
	if len(got) != 1 || got[0].Id != 1 {
		t.Fatalf("filtered topics = %#v, want only topic 1", got)
	}

	global := filterCurrentSearchTopics([]uint64{1, 2, 3, 4}, topicMap, readableCategoryIDs, true)
	if len(global) != 2 || global[0].Id != 1 || global[1].Id != 4 {
		t.Fatalf("global filtered topics = %#v, want published unprocessed topics 1 and 4", global)
	}
}
