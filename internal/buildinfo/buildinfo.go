package buildinfo

import (
	"fmt"
	"runtime/debug"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
	Dirty   = "unknown"
	BuiltBy = "source"
)

type Info struct {
	Version string
	Commit  string
	Date    string
	Dirty   string
	BuiltBy string
}

func Current() Info {
	version := Version
	commit := Commit
	date := Date
	dirty := Dirty

	if info, ok := debug.ReadBuildInfo(); ok {
		if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if commit == "unknown" && setting.Value != "" {
					commit = setting.Value
				}
			case "vcs.time":
				if date == "unknown" && setting.Value != "" {
					date = setting.Value
				}
			case "vcs.modified":
				if dirty == "unknown" && setting.Value != "" {
					dirty = setting.Value
				}
			}
		}
	}

	return Info{
		Version: version,
		Commit:  commit,
		Date:    date,
		Dirty:   dirty,
		BuiltBy: BuiltBy,
	}
}

func String() string {
	info := Current()
	return fmt.Sprintf("%s commit=%s date=%s dirty=%s builtBy=%s", info.Version, info.Commit, info.Date, info.Dirty, info.BuiltBy)
}
