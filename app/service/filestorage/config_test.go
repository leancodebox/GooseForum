package filestorage

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

func TestConfigureFromPreferencesSelectsDatabase(t *testing.T) {
	setStoragePreference(t, "storage.driver", DatabaseDriver)
	if err := ConfigureFromPreferences(); err != nil {
		t.Fatalf("configure: %v", err)
	}
	if current().writeStore.Driver() != DatabaseDriver || len(current().stores) != 1 {
		t.Fatalf("configured service = %#v", current().stores)
	}
}

func TestConfigureFromPreferencesSelectsS3WithDatabaseFallback(t *testing.T) {
	setStoragePreference(t, "storage.driver", S3Driver)
	setStoragePreference(t, "storage.s3.endpoint", "https://objects.example.com")
	setStoragePreference(t, "storage.s3.bucket", "forum")
	setStoragePreference(t, "storage.s3.accessKey", "access")
	setStoragePreference(t, "storage.s3.secretKey", "secret")
	setStoragePreference(t, "storage.s3.region", "us-east-1")
	if err := ConfigureFromPreferences(); err != nil {
		t.Fatalf("configure: %v", err)
	}
	service := current()
	if service.writeStore.Driver() != S3Driver || service.stores[S3Driver] == nil || service.stores[DatabaseDriver] == nil {
		t.Fatalf("configured stores = %#v", service.stores)
	}
}

func TestConfigureFromPreferencesRejectsUnknownDriver(t *testing.T) {
	setStoragePreference(t, "storage.driver", "unknown")
	if err := ConfigureFromPreferences(); err == nil {
		t.Fatal("unknown driver configured")
	}
}

func setStoragePreference(t *testing.T, key string, value any) {
	t.Helper()
	old := preferences.GetString(key)
	preferences.Set(key, value)
	t.Cleanup(func() {
		preferences.Set(key, old)
		_ = Configure(NewDatabaseStore())
	})
}
