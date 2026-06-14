package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	version, commit, date := values()
	return fmt.Sprintf("forge %s (%s, %s)", version, commit, date)
}

func values() (string, string, string) {
	version := Version
	commit := Commit
	date := Date
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, date
	}

	if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if commit == "unknown" && len(setting.Value) >= 12 {
				commit = setting.Value[:12]
			} else if commit == "unknown" && setting.Value != "" {
				commit = setting.Value
			}
		case "vcs.time":
			if date == "unknown" && setting.Value != "" {
				date = setting.Value
			}
		}
	}
	return version, commit, date
}
