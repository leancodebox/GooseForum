package datamigration

import (
	"encoding/json"
	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"gorm.io/gorm"
	"testing"
)

func TestAttachmentSettingsMigrationPreservesImagePolicy(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	row := pageConfig.Entity{PageType: pageConfig.PostingSettings, Config: `{"uploadControl":{"allowAttachments":false,"maxAttachmentSizeKb":123,"authorizedExtensions":[".png"]}}`}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	if err := MigrateAttachmentsWithDB(db); err != nil {
		t.Fatal(err)
	}
	db.First(&row, row.Id)
	var c pageConfig.PostingContent
	if err := json.Unmarshal([]byte(row.Config), &c); err != nil {
		t.Fatal(err)
	}
	if c.UploadControl.AllowAttachments || c.UploadControl.MaxAttachmentSizeKb != 123 || !c.Attachments.Enabled {
		t.Fatalf("bad migration: %+v", c)
	}
	before := row.Config
	if err := MigrateAttachmentsWithDB(db); err != nil {
		t.Fatal(err)
	}
	db.First(&row, row.Id)
	if row.Config != before {
		t.Fatal("migration not idempotent")
	}
}
