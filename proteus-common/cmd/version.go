package cmd

import (
	"os"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/thetorproject/proteus/proteus-common"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Proteus",
	Long:  `All software has versions. This is Proteus'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printProteusVersion()
		return nil
	},
}

func printProteusVersion() {
	if common.BuildDate == "" {
		setBuildDate()
	} else {
		formatBuildDate()
	}
	if common.CommitHash == "" {
		fmt.Printf("Proteus v%s %s/%s BuildDate: %s\n", common.CurrentProteusVersion, runtime.GOOS, runtime.GOARCH, common.BuildDate)
	} else {
		fmt.Printf("Proteus v%s-%s %s/%s BuildDate: %s\n", common.CurrentProteusVersion, strings.ToUpper(common.CommitHash), runtime.GOOS, runtime.GOARCH, common.BuildDate)
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
