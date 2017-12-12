package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/thetorproject/proteus/proteus-common/cmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	logLevel  string
	psqlDBStr string
)

var ctx = log.WithFields(log.Fields{
	"pkg": "cmd",
	"cmd": "proteus-registry",
})

// RootCmd for proteus-registry
var RootCmd = &cobra.Command{
	Use:   "proteus-registry",
	Short: "I know what probes are out there",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute parse command line arguments and run
func Execute() {
	RootCmd.AddCommand(cmd.VersionCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/proteus/proteus-registry.toml)")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "", "info", "Set the log level")
	RootCmd.PersistentFlags().StringP("db-url", "", "", "Set the url of the postgres database (ex. postgres://username:password@host/dbname?sslmode=verify-full)")
	viper.BindPFlag("database.url", RootCmd.PersistentFlags().Lookup("db-url"))
}

func initConfig() {
	viper.SetConfigName("proteus-registry")
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
		ctx.WithError(err).Errorf("using default configuration")
	}

	log.SetHandler(cli.Default)
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		fmt.Println("Invalid log level. Must be one of debug, info, warn, error, fatal")
	}
	log.SetLevel(level)
}
