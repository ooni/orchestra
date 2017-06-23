package cmd

import (
	"runtime"
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
	"env": viper.GetString("environment"),
})

var RootCmd = &cobra.Command{
	Use:   "proteus-notify",
	Short: "I tell probes what to do",
	Long: ``,
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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/proteus/proteus-notify.toml)")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "info", "Set the log level")
	RootCmd.PersistentFlags().StringP("db-url", "", "", "Set the url of the postgres database (ex. postgres://username:password@host/dbname?sslmode=verify-full)")
	viper.BindPFlag("database.url", RootCmd.PersistentFlags().Lookup("db-url"))
	viper.BindPFlag("core.log-level", RootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetDefault("database.active-probes-table", "active_probes")
	viper.SetDefault("database.probe-updates-table", "probe_updates")
	viper.SetDefault("fcm.max-retries", 5)
	viper.SetDefault("fcm.max-retries", 5)
	viper.SetDefault("core.environment", "production")
	viper.SetDefault("core.worker-num", runtime.NumCPU())
	viper.SetDefault("core.queue-size", 2048)
}

func initConfig() {
	viper.SetConfigName("proteus-notify")
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
	} else {
		ctx.WithError(err).Errorf("failed to read config file: %s", viper.ConfigFileUsed())
	}

	log.SetHandler(cli.Default)
	level, err := log.ParseLevel(viper.GetString("core.log-level"))
	if err != nil {
		ctx.Error("Invalid log level. Must be one of debug, info, warn, error, fatal")
	}
	log.SetLevel(level)
}
