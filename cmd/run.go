package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/digital-mob-filecoin/filstats-client/core"
	"github.com/digital-mob-filecoin/filstats-client/node/lotus"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Extract data from filecoin nodes and send to filstats",
	Run: func(cmd *cobra.Command, args []string) {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT)
		signal.Notify(stopChan, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			select {
			case <-stopChan:
				log.Info("Got stop signal. Finishing work.")
				// close whatever there is to close
				cancel()

				log.Info("Work done. Goodbye!")
			}
		}()

	runLoop:
		for {
			node := lotus.New(lotus.Config{
				Url:   viper.GetString("node.addr"),
				Token: viper.GetString("node.auth-token"),
			})

			c, err := core.New(core.Config{
				Filstats: core.FilstatsConfig{
					ServerAddr: viper.GetString("filstats.addr"),
					TLS:        viper.GetBool("filstats.tls"),
					ClientName: viper.GetString("filstats.client-name"),
				},
				DataFolder: viper.GetString("data-folder"),
			}, node)
			if err != nil {
				log.Panic(err)
			}

			err = c.Run(ctx)
			if err != nil {
				log.Error(err)
			}

			select {
			case <-ctx.Done():
				break runLoop
			default:
			}

			log.Info("restarting run loop")
			time.Sleep(1 * time.Second)
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

	runCmd.Flags().String("node.type", "lotus", "The type of Filecoin node we'll extract data from")
	viper.BindPFlag("node.type", runCmd.Flag("node.type"))

	runCmd.Flags().String("node.addr", "http://localhost:1234/rpc/v0", "The address of the node's RPC api")
	viper.BindPFlag("node.addr", runCmd.Flag("node.addr"))

	runCmd.Flags().String("node.auth-token", "", "Authorization Bearer token to authenticate RPC requests to the node")
	viper.BindPFlag("node.auth-token", runCmd.Flag("node.auth-token"))

	runCmd.Flags().String("data-folder", ".", "The folder where filstats-client will persist information. Used mostly to persist the auth token.")
	viper.BindPFlag("data-folder", runCmd.Flag("data-folder"))
}
