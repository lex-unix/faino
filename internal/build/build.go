package build

import "runtime/debug"

// Version is a build-time parameter set via -ldflags
var Version = "DEV"

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}
