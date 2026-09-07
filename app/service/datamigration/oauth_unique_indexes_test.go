package datamigration

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"gorm.io/gorm"
)

func TestEnforceOAuthBindingUniqueness(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE user_o_auth (
		id integer primary key, user_id integer not null, provider text not null,
		provider_uid text not null, access_token text not null default '', refresh_token text not null default '',
		token_expiry datetime, scopes text, raw_user_data text, created_at datetime, updated_at datetime
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE INDEX idx_provider_uid ON user_o_auth(provider, provider_uid)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE INDEX idx_user_provider ON user_o_auth(user_id, provider)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO user_o_auth (id,user_id,provider,provider_uid) VALUES
		(1,10,'github','a'), (2,11,'github','a'), (3,10,'github','b'), (4,12,'google','c')`).Error; err != nil {
		t.Fatal(err)
	}

	result := EnforceOAuthBindingUniquenessWithDB(db)
	if result.Failed != 0 || result.DuplicatesRemoved != 2 {
		t.Fatalf("migration result = %#v", result)
	}
	var bindings []userOAuth.Entity
	if err := db.Order("id").Find(&bindings).Error; err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 2 || bindings[0].Id != 1 || bindings[1].Id != 4 {
		t.Fatalf("remaining bindings = %#v", bindings)
	}
	for _, tc := range []userOAuth.Entity{
		{UserId: 99, Provider: "github", ProviderUid: "a"},
		{UserId: 10, Provider: "github", ProviderUid: "different"},
	} {
		if err := db.Create(&tc).Error; err == nil {
			t.Fatalf("expected unique constraint for %#v", tc)
		}
	}
}
