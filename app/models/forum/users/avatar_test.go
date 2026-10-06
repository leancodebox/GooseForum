package users

import "testing"

func TestExistingPresetAvatarGetsCurrentVersion(t *testing.T) {
	user := EntityComplete{AvatarUrl: "/static/pic/7.webp"}
	if got := user.GetWebAvatarUrl(); got != "/static/pic/7.webp?t=1791244800" {
		t.Fatalf("GetWebAvatarUrl() = %q", got)
	}
	if user.AvatarUrl != "/static/pic/7.webp" {
		t.Fatal("display URL must not modify the stored avatar")
	}
}
