package cmd

import (
	"github.com/ooni/orchestra/registry/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `This command starts the OONI registry service.`,
	Run: func(cmd *cobra.Command, args []string) {
		registry.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	startCmd.PersistentFlags().IntP("port", "", 8080, "Which port we should bind to")
	startCmd.PersistentFlags().StringP("address", "", "127.0.0.1", "Which interface we should listen on")
	viper.BindPFlag("api.port", startCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("api.address", startCmd.PersistentFlags().Lookup("address"))
}
