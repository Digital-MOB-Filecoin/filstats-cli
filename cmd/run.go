package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/digital-mob-filecoin/filstats-client/core"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Extract data from filecoin nodes and send to filstats",
	Run: func(cmd *cobra.Command, args []string) {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT)
		signal.Notify(stopChan, syscall.SIGTERM)

		c := core.New(core.Config{
			Filstats: core.FilstatsConfig{
				ServerAddr: viper.GetString("filstats.addr"),
				TLS:        viper.GetBool("filstats.tls"),
				ClientName: viper.GetString("filstats.client-name"),
			},
		})
		go c.Run()

		select {
		case <-stopChan:
			log.Info("Got stop signal. Finishing work.")
			// close whatever there is to close

			log.Info("Work done. Goodbye!")
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().String("filstats.addr", "localhost:3002", "Address of the Filstats server's gRPC api")
	viper.BindPFlag("filstats.addr", runCmd.Flag("filstats.addr"))

	runCmd.Flags().Bool("filstats.tls", false, "Enable/disable the secure connection to Filstats server")
	viper.BindPFlag("filstats.tls", runCmd.Flag("filstats.tls"))

	runCmd.Flags().String("filstats.client-name", "Client", "The name that will be displayed in the Filstats dashboard")
	viper.BindPFlag("filstats.client-name", runCmd.Flag("filstats.client-name"))
}
