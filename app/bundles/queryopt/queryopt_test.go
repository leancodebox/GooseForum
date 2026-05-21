package queryopt

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type testTask struct {
	ID     uint64 `gorm:"primaryKey"`
	Status uint8
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
