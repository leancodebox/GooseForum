package attachmentservice

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/attachments"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"gorm.io/gorm"
)

type accessStore struct{}

func (accessStore) SystemGroupIDs() (map[string]uint64, error) {
	return map[string]uint64{"everyone": 1, "registered": 2}, nil
}
func (accessStore) ActiveCustomGroupIDs(user uint64) ([]uint64, error) {
	if user == 11 {
		return []uint64{3}, nil
	}
	return nil, nil
}
func (accessStore) EnabledCategoryGrants(group uint64) ([]accesscontrol.CategoryGrant, error) {
	if group == 1 {
		return []accesscontrol.CategoryGrant{{CategoryID: 1, Capability: accesscontrol.CapabilityRead}}, nil
	}
	if group == 3 {
		return []accesscontrol.CategoryGrant{{CategoryID: 2, Capability: accesscontrol.CapabilityRead}}, nil
	}
	return nil, nil
}
func TestFileValidation(t *testing.T) {
	c := defaultconfig.GetDefaultPostingSettingsConfig().Attachments
	for _, tc := range []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"notes.TXT", []byte("hello 世界"), true}, {"report.pdf", []byte("%PDF-1.7\n"), true}, {"archive.zip", []byte("PK\x03\x04data"), true},
		{"../notes.txt", []byte("hi"), false}, {"bad\r.txt", []byte("hi"), false}, {"page.html", []byte("hello"), false}, {"fake.pdf", []byte("html"), false}, {"binary.txt", []byte{0xff}, false}, {"null.txt", []byte{0}, false}, {"empty.txt", nil, false}, {"big.txt", bytes.Repeat([]byte("a"), c.MaxFileSizeKb*1024+1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateFile(tc.name, tc.data, c); (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
	c.Extensions = nil
	if err := ValidateFile("a.txt", []byte("hello"), c); err == nil {
		t.Fatal("empty whitelist allowed upload")
	}
	c.MaxFileSizeKb = 0
	if ValidateConfig(c) == nil {
		t.Fatal("invalid configuration accepted")
	}
}
func TestUploadBindingVisibilityAndLegacyEdits(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&attachments.Entity{}, &posts.Entity{}, &topics.Entity{}, &users.EntityComplete{}, &pageConfig.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := db4fileconnect.Connect().AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	for _, u := range []users.EntityComplete{{Id: 11, Username: "attachment-member", UsernameLower: "attachment-member", TrustLevel: trustlevel.Member, CreatedAt: time.Now().Add(-48 * time.Hour)}, {Id: 12, Username: "attachment-new", UsernameLower: "attachment-new", TrustLevel: trustlevel.Newcomer, CreatedAt: time.Now().Add(-48 * time.Hour)}} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
	}
	old := accesscontrol.Default
	accesscontrol.Default = accesscontrol.NewResolver(accessStore{}, func(id uint64) bool { return id == 99 }, nil)
	t.Cleanup(func() { accesscontrol.Default = old })
	ctx := context.Background()
	previous := pageConfig.GetByPageType(pageConfig.TrustSettings)
	config := pageConfig.DefaultTrustSettings()
	config.LevelLimits[1].AllowAttachmentUpload = false
	raw, _ := json.Marshal(config)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	t.Cleanup(func() {
		if previous.Id != 0 {
			_ = pageConfig.SaveConfig(pageConfig.TrustSettings, previous.Config)
		} else {
			db.Where("page_type = ?", pageConfig.TrustSettings).Delete(&pageConfig.Entity{})
		}
		hotdataserve.ClearTrustSettingsCache()
	})
	if _, err := Policy(11); err == nil {
		t.Fatal("disabled member attachment permission ignored")
	}
	config.LevelLimits[1].AllowAttachmentUpload = true
	config.LevelLimits[0].AllowAttachmentUpload = true
	raw, _ = json.Marshal(config)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	if _, err := Policy(12); err != nil {
		t.Fatalf("enabled newcomer permission ignored: %v", err)
	}
	config.LevelLimits[0].AllowAttachmentUpload = false
	raw, _ = json.Marshal(config)
	if err := pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)); err != nil {
		t.Fatal(err)
	}
	hotdataserve.ClearTrustSettingsCache()
	if _, err := Upload(ctx, 12, "no.txt", bytes.NewBufferString("no")); err == nil {
		t.Fatal("newcomer uploaded")
	}
	accountReads := 0
	callback := "attachment-policy-query-budget"
	if err := db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "users" {
			accountReads++
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Query().Remove(callback) })
	f, err := Upload(ctx, 11, "hello.txt", bytes.NewBufferString("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	if accountReads != 1 {
		t.Fatalf("account reads=%d, want 1", accountReads)
	}
	db.Callback().Query().Remove(callback)
	e, err := attachments.Get(f.ID)
	if err != nil {
		t.Fatal(err)
	}
	if CanRead(0, e) || CanRead(12, e) || !CanRead(11, e) {
		t.Fatal("unbound visibility")
	}
	_, obj, err := Open(ctx, 11, f.ID)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(obj.Body)
	obj.Body.Close()
	if string(data) != "hello world" {
		t.Fatal("download content")
	}
	ids := []uint64{f.ID}
	if _, err := Prepare(12, 0, nil, &ids); err == nil {
		t.Fatal("foreign reference")
	}
	duplicates := []uint64{f.ID, f.ID}
	if _, err := Prepare(11, 0, nil, &duplicates); err == nil {
		t.Fatal("duplicate reference")
	}
	files, err := Prepare(11, 0, nil, &ids)
	if err != nil {
		t.Fatal(err)
	}
	topic := topics.Entity{Id: 21, UserId: 11, Status: 1, MainCategoryId: 1, CategoryIds: []uint64{1}}
	if err := db.Create(&topic).Error; err != nil {
		t.Fatal(err)
	}
	p := posts.Entity{Id: 31, TopicId: 21, PostNo: 2, UserId: 11, Content: "hello", ModerationVersion: 1, Attachments: files}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		return attachments.BindWithDB(tx, 11, p.Id, files)
	}); err != nil {
		t.Fatal(err)
	}
	// SaveReviewed starts from the posts table; binding must explicitly query attachments.
	p.ModerationVersion = 2
	if err := posts.SaveReviewed(&p, 1); err != nil {
		t.Fatalf("retaining attachments on edit: %v", err)
	}
	e, _ = attachments.Get(f.ID)
	if !CanRead(0, e) {
		t.Fatal("public download denied")
	}
	if err := db.Model(&topics.Entity{}).Where("id = ?", topic.Id).Update("process_status", 1).Error; err != nil {
		t.Fatal(err)
	}
	if CanRead(0, e) || CanRead(11, e) || CanRead(12, e) || !CanRead(99, e) {
		t.Fatal("manually blocked topic exposed attachment to non-moderator")
	}
	if err := db.Model(&topics.Entity{}).Where("id = ?", topic.Id).Updates(map[string]any{"process_status": 0, "status": 0}).Error; err != nil {
		t.Fatal(err)
	}
	if CanRead(0, e) || CanRead(12, e) || !CanRead(11, e) || !CanRead(99, e) {
		t.Fatal("draft topic attachment visibility")
	}
	if err := db.Model(&topics.Entity{}).Where("id = ?", topic.Id).Update("status", 1).Error; err != nil {
		t.Fatal(err)
	}
	if err := Cancel(ctx, 11, f.ID); err == nil {
		t.Fatal("cancel removed bound attachment")
	}
	if _, err := Prepare(11, 32, nil, &ids); err == nil {
		t.Fatal("cross-post reference")
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return attachments.BindWithDB(tx, 11, 32, files) }); err == nil {
		t.Fatal("claim moved to another post")
	}
	if err := db.Model(&posts.Entity{}).Where("id = ?", p.Id).Update("process_status", 1).Error; err != nil {
		t.Fatal(err)
	}
	if CanRead(0, e) || CanRead(12, e) || !CanRead(11, e) || !CanRead(99, e) {
		t.Fatal("hidden reply access")
	}
	db.Model(&posts.Entity{}).Where("id = ?", p.Id).Update("process_status", 0)
	db.Model(&topics.Entity{}).Where("id = ?", topic.Id).Updates(map[string]any{"main_category_id": 2})
	if CanRead(0, e) || CanRead(12, e) || !CanRead(11, e) {
		t.Fatal("private category access")
	}
	// Tightened upload policy must not invalidate retained files.
	db.Model(&users.EntityComplete{}).Where("id = ?", 11).Update("trust_level", trustlevel.Newcomer)
	if _, err := Prepare(11, p.Id, files, &ids); err != nil {
		t.Fatalf("retaining old file after demotion: %v", err)
	}
	if _, err := Upload(ctx, 11, "denied.txt", bytes.NewBufferString("hello")); err == nil {
		t.Fatal("demoted user uploaded")
	}
	db.Model(&users.EntityComplete{}).Where("id = ?", 11).Update("trust_level", trustlevel.Member)
	current, err := Prepare(11, p.Id, files, nil)
	if err != nil || len(current) != 1 {
		t.Fatal("legacy request cleared attachments")
	}
	// A failure after one successful claim must roll back both that claim and the post update.
	unbound := attachments.Entity{UserID: 11, ObjectName: "attachment-rollback", Filename: "rollback.txt", Size: 1}
	if err := attachments.Create(&unbound); err != nil {
		t.Fatal(err)
	}
	failed := p
	failed.Content = "must roll back"
	failed.ModerationVersion = 3
	failed.Attachments = append(append([]attachments.Info(nil), files...), info(unbound), attachments.Info{ID: 987654321})
	if err := posts.SaveReviewed(&failed, 2); err == nil {
		t.Fatal("missing attachment did not fail save")
	}
	unboundAfter, err := attachments.Get(unbound.ID)
	if err != nil || unboundAfter.PostID != 0 {
		t.Fatalf("partial attachment claim persisted: %+v, %v", unboundAfter, err)
	}
	unchanged := posts.Get(p.Id)
	if unchanged.Content != "hello" || unchanged.ModerationVersion != 2 || len(unchanged.Attachments) != 1 {
		t.Fatalf("failed binding changed post: %+v", unchanged)
	}
	// A stale edit cannot change either content or references.
	p.ModerationVersion = 3
	p.Attachments = nil
	p.Content = "stale"
	if err := posts.SaveReviewed(&p, 0); err == nil {
		t.Fatal("stale version saved")
	}
	saved := posts.Get(p.Id)
	if len(saved.Attachments) != 1 || saved.Content != "hello" {
		t.Fatal("version conflict was not atomic")
	}
	if err := posts.SaveReviewed(&p, 2); err != nil {
		t.Fatal(err)
	}
	if CanRead(11, e) {
		t.Fatal("removed attachment still downloadable")
	}
	second, err := Upload(ctx, 11, "cancel.txt", bytes.NewBufferString("cancel"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Cancel(ctx, 11, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Open(ctx, 11, second.ID); err == nil {
		t.Fatal("cancelled file readable")
	}
}
