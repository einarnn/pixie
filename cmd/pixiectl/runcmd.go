package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Connect to pxGrid Cloud and stream messages for a device",
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceName := viper.GetString("run.device")
		if deviceName == "" {
			return fmt.Errorf("device name is required (use --device)")
		}

		app, _, err := setup()
		if err != nil {
			return err
		}

		// Catch termination signal
		_, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// --- do other stuff ---
		logger.Debugf("*** pausing for a few seconds...")
		time.Sleep(time.Second * 5)

		// --- done with other stuff ---
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		select {
		case <-ctx.Done():
			logger.Debugf("Terminating...")
		case err := <-app.Error:
			logger.Errorf("App error: %v", err)
		}
		if err = app.Close(); err != nil {
			return err
		}
		return nil
	},
}
