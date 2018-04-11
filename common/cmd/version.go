package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/spf13/cobra"
	"github.com/ooni/orchestra/common"
)

// VersionCmd is the command used to output the version of orchestra
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of OONI Orchestra",
	Long:  `All software has versions. This is OONI Orchestra'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printOrchestraVersion()
		return nil
	},
}

func printOrchestraVersion() {
	if common.BuildDate == "" {
		setBuildDate()
	} else {
		formatBuildDate()
	}
	if common.CommitHash == "" {
		fmt.Printf("OONI Orchestra v%s %s/%s BuildDate: %s\n", common.CurrentOrchestraVersion, runtime.GOOS, runtime.GOARCH, common.BuildDate)
	} else {
		fmt.Printf("OONI Orchestra v%s-%s %s/%s BuildDate: %s\n", common.CurrentOrchestraVersion, strings.ToUpper(common.CommitHash), runtime.GOOS, runtime.GOARCH, common.BuildDate)
	}
}

func setBuildDate() {
	fname, _ := osext.Executable()
	dir, err := filepath.Abs(filepath.Dir(fname))
	if err != nil {
		fmt.Println("error: failed to get executable date")
		return
	}
	fi, err := os.Lstat(filepath.Join(dir, filepath.Base(fname)))
	if err != nil {
		fmt.Println("error: failed to lstat")
		return
	}
	t := fi.ModTime()
	common.BuildDate = t.Format(time.RFC3339)
}

func formatBuildDate() {
	t, _ := time.Parse("2006-01-02T15:04:05-0700", common.BuildDate)
	common.BuildDate = t.Format(time.RFC3339)
}
