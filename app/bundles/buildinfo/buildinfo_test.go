package buildinfo

import "testing"

func TestBuildMode(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{version: "dev", want: "development"},
		{version: "dev-dirty", want: "development"},
		{version: "v1.2.3-snapshot", want: "snapshot"},
		{version: "v1.2.3", want: "release"},
		{version: "custom-build", want: "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := buildMode(tt.version); got != tt.want {
				t.Fatalf("buildMode(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestGetUsesConfiguredBuildInfo(t *testing.T) {
	oldVersion, oldCommit, oldBuildDate := Version, Commit, BuildDate
	t.Cleanup(func() {
		Version, Commit, BuildDate = oldVersion, oldCommit, oldBuildDate
	})

	Version = "v1.2.3"
	Commit = "abc123"
	BuildDate = "2026-06-05T00:00:00Z"

	info := Get()
	if info.Version != Version {
		t.Fatalf("version = %q, want %q", info.Version, Version)
	}
	if info.Commit != Commit {
		t.Fatalf("commit = %q, want %q", info.Commit, Commit)
	}
	if info.BuildDate != BuildDate {
		t.Fatalf("buildDate = %q, want %q", info.BuildDate, BuildDate)
	}
	if info.Mode != "release" {
		t.Fatalf("mode = %q, want release", info.Mode)
	}
}
