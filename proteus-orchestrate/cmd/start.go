package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		events.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	startCmd.PersistentFlags().IntP("port", "", 8082, "Which port we should bind to")
	startCmd.PersistentFlags().StringP("address", "", "127.0.0.1", "Which interface we should listen on")
	viper.BindPFlag("api.port", startCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("api.address", startCmd.PersistentFlags().Lookup("address"))
}
