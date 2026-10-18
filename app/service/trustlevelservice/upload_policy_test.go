package trustlevelservice

import (
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestPendingDirectUploadAndAvatarFileCount(t *testing.T) {
	db := db4fileconnect.Connect()
	if err := db.AutoMigrate(&filedata.Entity{}); err != nil {
		t.Fatal(err)
	}
	const id uint64 = 98999101
	row := filedata.Entity{Name: "quota-pending-direct", UserId: id, StorageStatus: filedata.StorageStatusPending}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Delete(&row) })
	p := UserPolicy{Account: users.AccountState{Id: id, CreatedAt: time.Now().Add(-48 * time.Hour)}, Limits: pageConfig.TrustLimits{AllowImageUpload: true, MaxDailyUploads: new(1)}}
	if err := p.ValidateExistingUpload(); err != nil {
		t.Fatalf("pending upload charged twice: %v", err)
	}
	if err := p.ValidateUpload(1, false); err == nil {
		t.Fatal("new upload exceeded quota")
	}
	p.Limits.MaxDailyUploads = new(3)
	if err := p.ValidateUpload(3, true); err == nil {
		t.Fatal("avatar variants not counted")
	}
	if err := p.ValidateUpload(2, true); err != nil {
		t.Fatal("avatar at exact quota denied", err)
	}
}

func TestRoleExemptionAndRegistrationAgeSemantics(t *testing.T) {
	p := UserPolicy{Account: users.AccountState{RoleId: 1, CreatedAt: time.Now()}, Limits: pageConfig.TrustLimits{AllowImageUpload: true, MinUploadAccountAgeMinutes: 60, MaxDailyUploads: new(0), MinPostAccountAgeMinutes: 60}}
	if err := p.ValidateUpload(1, false); err != nil {
		t.Fatalf("community role exemption lost: %v", err)
	}
	var policy *PolicyError
	if err := p.ValidateUpload(1, true); !errors.As(err, &policy) || policy.Code != "upload.cooldown" {
		t.Fatalf("avatar bypassed age: %v", err)
	}
	if err := p.ValidatePostAge(); !errors.As(err, &policy) || policy.Code != "topic.post.cooldown" {
		t.Fatalf("posting gained role exemption: %v", err)
	}
	p.Account.CreatedAt = time.Now().Add(-2 * time.Hour)
	if err := p.ValidateUpload(3, true); !errors.As(err, &policy) || policy.Code != "upload.dailyLimit.avatar" {
		t.Fatalf("avatar quota bypassed: %v", err)
	}
	p.Limits.MaxDailyUploads = nil
	if err := p.ValidateUpload(3, true); err != nil {
		t.Fatalf("unlimited avatar quota: %v", err)
	}
	if err := p.ValidatePostAge(); err != nil {
		t.Fatal("age treated as posting interval", err)
	}
	p.Account.RoleId = 0
	p.Limits.MaxDailyUploads = new(0)
	if err := p.ValidateExistingUpload(); !errors.As(err, &policy) {
		t.Fatal("existing upload bypassed zero quota")
	}
}
