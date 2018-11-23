package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/ooni/orchestra/common/cmd"
	"github.com/ooni/orchestra/orchestrate/orchestrate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	logLevel string
)

var ctx = log.WithFields(log.Fields{
	"pkg": "cmd",
	"cmd": "ooni-operator",
})

// RootCmd where all commands begin
var RootCmd = &cobra.Command{
	Use:   "ooni-operator",
	Short: "OONI Operator tools",
	Long:  orchestrate.LongDescription,
}

// Execute parse the command arguments
func Execute() {
	RootCmd.AddCommand(cmd.VersionCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.ooni-operator/config.toml)")
	RootCmd.PersistentFlags().StringP("log-level", "", "info", "Set the log level")
	viper.BindPFlag("core.log-level", RootCmd.PersistentFlags().Lookup("log-level"))
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.ooni-operator")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_") // Allows us to defined keys with -, but set them in via env variables with _
	viper.SetEnvKeyReplacer(replacer)

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	log.SetHandler(cli.Default)
	if err := viper.ReadInConfig(); err == nil {
		ctx.Infof("using config file: %s", viper.ConfigFileUsed())
	} else {
		ctx.WithError(err).Errorf("failed to load config file")
	}

	level, err := log.ParseLevel(viper.GetString("core.log-level"))
	if err != nil {
		fmt.Println("Invalid log level. Must be one of debug, info, warn, error, fatal")
	}
	log.SetLevel(level)
}
