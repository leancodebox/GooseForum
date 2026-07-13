package posts

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestPostRepositoryWindows(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate posts: %v", err)
	}
	conn.Where("1 = 1").Delete(&Entity{})

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	conn.Create(&[]Entity{
		{Id: 10, TopicId: 1, PostNo: 1, UserId: 1, Content: "first", CreatedAt: now},
		{Id: 11, TopicId: 1, PostNo: 2, UserId: 2, Content: "second", CreatedAt: now.Add(time.Minute)},
		{Id: 12, TopicId: 1, PostNo: 3, UserId: 3, Content: "third", CreatedAt: now.Add(2 * time.Minute)},
		{Id: 20, TopicId: 2, PostNo: 1, UserId: 4, Content: "other", CreatedAt: now},
	})

	first := GetFirstPageByTopicId(1)
	if len(first) != 3 || first[0].PostNo != 1 || first[2].PostNo != 3 {
		t.Fatalf("GetFirstPageByTopicId() = %#v", postNos(first))
	}

	desc := GetByTopicPostNoDesc(1, 2)
	if len(desc) != 2 || desc[0].PostNo != 2 || desc[1].PostNo != 3 {
		t.Fatalf("GetByTopicPostNoDesc() = %#v, want ascending returned window [2 3]", postNos(desc))
	}

	after := GetByTopicPostNoAfter(1, 1, 10)
	if len(after) != 2 || after[0].PostNo != 2 || after[1].PostNo != 3 {
		t.Fatalf("GetByTopicPostNoAfter() = %#v", postNos(after))
	}

	before := GetByTopicPostNoBefore(1, 3, 10)
	if len(before) != 2 || before[0].PostNo != 1 || before[1].PostNo != 2 {
		t.Fatalf("GetByTopicPostNoBefore() = %#v", postNos(before))
	}

	if got := GetMaxPostNoByTopicId(1); got != 3 {
		t.Fatalf("GetMaxPostNoByTopicId()=%d, want 3", got)
	}

	if err := UpdateProcessStatus(11, 1); err != nil {
		t.Fatalf("UpdateProcessStatus() err=%v", err)
	}
	if got := Get(11); got.ProcessStatus != 1 {
		t.Fatalf("post ProcessStatus=%d, want 1", got.ProcessStatus)
	}
}

func postNos(rows []*Entity) []uint64 {
	res := make([]uint64, 0, len(rows))
	for _, row := range rows {
		res = append(res, row.PostNo)
	}
	return res
}
