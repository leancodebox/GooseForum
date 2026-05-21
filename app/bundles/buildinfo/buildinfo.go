package buildinfo

import (
	"runtime/debug"
	"strings"
)

var (
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	Mode      string `json:"mode"`
}

func Get() Info {
	version := strings.TrimSpace(Version)
	commit := strings.TrimSpace(Commit)
	buildDate := strings.TrimSpace(BuildDate)

	if version == "" || version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if version == "" {
				version = "dev"
			}
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if commit == "" {
						commit = setting.Value
					}
				case "vcs.modified":
					if version == "dev" && setting.Value == "true" {
						version = "dev-dirty"
					}
				}
			}
		}
	}

	if version == "" {
		version = "dev"
	}

	return Info{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		Mode:      buildMode(version),
	}
}

func buildMode(version string) string {
	switch {
	case version == "dev" || version == "dev-dirty":
		return "development"
	case strings.Contains(version, "snapshot"):
		return "snapshot"
	case strings.HasPrefix(version, "v"):
		return "release"
	default:
		return "custom"
	}
}
