package cmd

import (
	"github.com/ooni/orchestra/orchestrate/orchestrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		orchestrate.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.
	startCmd.PersistentFlags().IntP("port", "", 8082, "Which port we should bind to")
	startCmd.PersistentFlags().StringP("address", "", "127.0.0.1", "Which interface we should listen on")
	viper.BindPFlag("api.port", startCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("api.address", startCmd.PersistentFlags().Lookup("address"))
}
