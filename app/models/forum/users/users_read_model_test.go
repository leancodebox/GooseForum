package users

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestIdentityInfersSelectedColumnsFromReadModel(t *testing.T) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	query := conn.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var identity Identity
		return tx.Table(tableName).Model(&EntityComplete{}).Where("id = ?", 1).First(&identity)
	})
	if strings.Contains(query, "SELECT *") {
		t.Fatalf("identity query did not infer columns from read model: %s", query)
	}
	for _, column := range []string{"id", "username"} {
		if !strings.Contains(query, column) {
			t.Fatalf("identity query missing %s: %s", column, query)
		}
	}
}
