package npmctl

import "fmt"

var (
	Version   = "0.0.1"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func versionSummary() string {
	return fmt.Sprintf("npmctl %s", Version)
}

func fullVersion() string {
	return fmt.Sprintf("%s\ncommit: %s\nbuilt: %s", versionSummary(), Commit, BuildDate)
}
