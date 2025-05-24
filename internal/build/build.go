package build

import "runtime/debug"

// Build-time parameters set via -ldflags
var Version = "DEV"

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}
