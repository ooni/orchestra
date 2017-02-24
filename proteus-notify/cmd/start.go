package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/hellais/proteus/proteus-notify/notify"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		notify.StartServer()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().IntP("server-port", "", 8080, "Which port we should bind to")
	startCmd.PersistentFlags().StringP("server-address", "", "127.0.0.1", "Which interface we should listen on")
	viper.BindPFlag("server-port", startCmd.PersistentFlags().Lookup("server-port"))
	viper.BindPFlag("server-address", startCmd.PersistentFlags().Lookup("server-address"))
}
