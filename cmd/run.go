package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/digital-mob-filecoin/filstats-client/core"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Extract data from filecoin nodes and send to filstats",
	Run: func(cmd *cobra.Command, args []string) {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT)
		signal.Notify(stopChan, syscall.SIGTERM)

		c := core.New(core.Config{})
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
}
