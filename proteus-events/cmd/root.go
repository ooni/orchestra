package cmd

import (
	"strings"
	"fmt"
	"os"

	"github.com/thetorproject/proteus/proteus-common/cmd"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	logLevel string
	psqlDBStr string
)

var ctx = log.WithFields(log.Fields{
	"env": "production",
})

var RootCmd = &cobra.Command{
	Use:   "proteus-events",
	Short: "I receive events to propagate to probes",
	Long: `Is responsible for receiving events via the admin interface and triggering notifications via proteus-notify`,
}

func Execute() {
	RootCmd.AddCommand(cmd.VersionCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/proteus/proteus-events.yaml)")
	RootCmd.PersistentFlags().StringP("log-level", "", "info", "Set the log level")
	RootCmd.PersistentFlags().StringP("db-url", "", "", "Set the url of the postgres database (ex. postgres://username:password@host/dbname?sslmode=verify-full)")
	viper.BindPFlag("database.url", RootCmd.PersistentFlags().Lookup("db-url"))
	viper.BindPFlag("core.log-level", RootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetDefault("database.active-probes-table", "active_probes")
	viper.SetDefault("database.probe-updates-table", "probe_updates")
}

func initConfig() {
	viper.SetConfigName("proteus-events")
	viper.AddConfigPath("/etc/proteus/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_") // Allows us to defined keys with -, but set them in via env variables with _
	viper.SetEnvKeyReplacer(replacer)

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err == nil {
		ctx.Infof("using config file: %s", viper.ConfigFileUsed())
	}

	log.SetHandler(cli.Default)
	level, err := log.ParseLevel(viper.GetString("core.log-level"))
	if err != nil {
		fmt.Println("Invalid log level. Must be one of debug, info, warn, error, fatal")
	}
	log.SetLevel(level)
}
