package common

import (
	"fmt"
)

var (
	// CommitHash contains the current Git revision. Use make to build to make
	// sure this gets set.
	CommitHash string
	// BuildDate contains the date of the current build.
	BuildDate string
)

var orchestraInfo *OrchestraInfo

// OrchestraVersion contains the version information for orchestra
type OrchestraVersion struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

func (v OrchestraVersion) String() string {
	return orchestraVersion(v.Major, v.Minor, v.Patch, v.Suffix)
}

func orchestraVersion(major int, minor int, patchVersion int, suffix string) string {
	return fmt.Sprintf("%d.%d.%d%s", major, minor, patchVersion, suffix)
}

// CurrentOrchestraVersion is the current version of orchestra. Remember to change this before making a release
var CurrentOrchestraVersion = OrchestraVersion{
	Major:  0,
	Minor:  2,
	Patch:  5,
	Suffix: "",
}

// OrchestraInfo contains information for the current orchestra build
type OrchestraInfo struct {
	Version    string
	CommitHash string
	BuildDate  string
}

func init() {
	orchestraInfo = &OrchestraInfo{
		Version:    CurrentOrchestraVersion.String(),
		CommitHash: CommitHash,
		BuildDate:  BuildDate,
	}
}
