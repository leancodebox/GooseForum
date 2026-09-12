package migration

import (
	"os"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPostgresSchemaMigration(t *testing.T) {
	dsn := os.Getenv("GOOSEFORUM_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("GOOSEFORUM_POSTGRES_TEST_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	models := append(defaultSchemaModels(), &filedata.Entity{})
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate PostgreSQL schema: %v", err)
	}
}
