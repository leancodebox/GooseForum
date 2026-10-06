package hotdataserve

import (
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
)

func TestCategoryOrderUsesDescendingWeightAndRefreshesAfterSave(t *testing.T) {
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&category.Entity{}); err != nil {
		t.Fatalf("migrate categories: %v", err)
	}
	ids := []uint64{981101, 981102, 981103, 981104}
	conn.Delete(&category.Entity{}, ids)
	t.Cleanup(func() {
		conn.Delete(&category.Entity{}, ids)
		ClearCategoryCache()
	})
	items := []category.Entity{
		{Id: ids[0], Name: "Zero", Sort: 0},
		{Id: ids[2], Name: "High second", Sort: 5},
		{Id: ids[1], Name: "High first", Sort: 5},
		{Id: ids[3], Name: "Negative", Sort: -1},
	}
	if err := conn.Create(&items).Error; err != nil {
		t.Fatalf("create categories: %v", err)
	}
	ClearCategoryCache()
	assertOrder := func(want []uint64) {
		t.Helper()
		got := []uint64{}
		for _, item := range GetCategory() {
			if item.Id >= ids[0] && item.Id <= ids[3] {
				got = append(got, item.Id)
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("category order = %v, want %v", got, want)
		}
	}
	assertOrder([]uint64{ids[1], ids[2], ids[0], ids[3]})
	items[0].Sort = 10
	if category.SaveOrCreateById(&items[0]) != 1 {
		t.Fatal("update category weight failed")
	}
	ClearCategoryCache()
	assertOrder([]uint64{ids[0], ids[1], ids[2], ids[3]})
}
