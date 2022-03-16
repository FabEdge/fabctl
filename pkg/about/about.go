package about

import (
	"fmt"
)

var (
	version   = "0.0.0"                // semantic version X.Y.Z
	gitCommit = "00000000"             // sha1 from git
	buildTime = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

func DisplayVersion() {
	fmt.Printf("Version: %s\nBuildTime: %s\nGitCommit: %s\n", version, buildTime, gitCommit)
}
