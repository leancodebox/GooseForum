package queryopt

import (
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type testTask struct {
	ID     uint64 `gorm:"primaryKey"`
	Status uint8
}

func TestComparisonFragments(t *testing.T) {
	tests := []struct {
		name      string
		gotQuery  string
		gotValue  any
		wantQuery string
		wantValue any
	}{
		{name: "gt", gotQuery: first(Gt("score", 10)), gotValue: second(Gt("score", 10)), wantQuery: "score > ?", wantValue: 10},
		{name: "ge", gotQuery: first(Ge("score", 10)), gotValue: second(Ge("score", 10)), wantQuery: "score >= ?", wantValue: 10},
		{name: "lt", gotQuery: first(Lt("score", 10)), gotValue: second(Lt("score", 10)), wantQuery: "score < ?", wantValue: 10},
		{name: "le", gotQuery: first(Le("score", 10)), gotValue: second(Le("score", 10)), wantQuery: "score <= ?", wantValue: 10},
		{name: "eq", gotQuery: first(Eq("score", 10)), gotValue: second(Eq("score", 10)), wantQuery: "score = ?", wantValue: 10},
		{name: "ne", gotQuery: first(Ne("score", 10)), gotValue: second(Ne("score", 10)), wantQuery: "score <> ?", wantValue: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gotQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", tt.gotQuery, tt.wantQuery)
			}
			if !reflect.DeepEqual(tt.gotValue, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", tt.gotValue, tt.wantValue)
			}
		})
	}
}

func TestLikeAndOrderFragments(t *testing.T) {
	if query, value := Like("title", "goose"); query != "title like ?" || value != "%goose%" {
		t.Fatalf("Like = (%q, %q)", query, value)
	}
	if query, value := LeftLike("title", "goose"); query != "title like ?" || value != "%goose" {
		t.Fatalf("LeftLike = (%q, %q)", query, value)
	}
	if query, value := RightLike("title", "goose"); query != "title like ?" || value != "goose%" {
		t.Fatalf("RightLike = (%q, %q)", query, value)
	}
	if got := Desc("created_at"); got != "created_at desc" {
		t.Fatalf("Desc = %q", got)
	}
	if got := Asc("created_at"); got != "created_at asc" {
		t.Fatalf("Asc = %q", got)
	}
}

func TestNullFragments(t *testing.T) {
	if got := IsNull("deleted_at"); got != "deleted_at IS NULL" {
		t.Fatalf("IsNull = %q", got)
	}
	if got := IsNotNull("deleted_at"); got != "deleted_at IS NOT NULL" {
		t.Fatalf("IsNotNull = %q", got)
	}
}

func first(query string, _ any) string {
	return query
}

func second(_ string, value any) any {
	return value
}

func TestInWithIntSlice(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&testTask{}); err != nil {
		t.Fatal(err)
	}
	for _, status := range []uint8{0, 3, 4} {
		if err := db.Create(&testTask{Status: status}).Error; err != nil {
			t.Fatal(err)
		}
	}

	var tasks []testTask
	if err := db.Where(In("status", []int{0, 4})).Order("id asc").Find(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].Status != 0 || tasks[1].Status != 4 {
		t.Fatalf("statuses = [%d %d], want [0 4]", tasks[0].Status, tasks[1].Status)
	}
}

func TestNotInWithIntSlice(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&testTask{}); err != nil {
		t.Fatal(err)
	}
	for _, status := range []uint8{0, 3, 4} {
		if err := db.Create(&testTask{Status: status}).Error; err != nil {
			t.Fatal(err)
		}
	}

	var tasks []testTask
	if err := db.Where(NotIn("status", []int{0, 4})).Find(&tasks).Error; err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != 3 {
		t.Fatalf("tasks = %+v, want one status=3 task", tasks)
	}
}
