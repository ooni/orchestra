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

var proteusInfo *ProteusInfo

type ProteusVersion struct {
	// Major and minor
	Number	float32
	Patch	int
	Suffix	string
}

func (v ProteusVersion) String() string {
	return proteusVersion(v.Number, v.Patch, v.Suffix)
}

func proteusVersion(version float32, patchVersion int, suffix string) string {
	if patchVersion > 0 {
		return fmt.Sprintf("%.2f.%d%s", version, patchVersion, suffix)
	}
	return fmt.Sprintf("%.2f%s", version, suffix)
}

var CurrentProteusVersion = ProteusVersion{
	Number: 0.1,
	Patch: 0,
	Suffix:	"-dev",
}

type ProteusInfo struct {
	Version    string
	CommitHash string
	BuildDate  string
}

func init() {
	proteusInfo = &ProteusInfo{
		Version:    CurrentProteusVersion.String(),
		CommitHash: CommitHash,
		BuildDate:  BuildDate,
	}
}
