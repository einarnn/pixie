package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listDevicesCmd = &cobra.Command{
	Use:   "list-devices",
	Short: "List all devices for the configured tenant",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, tenant, err := setup()
		if err != nil {
			return err
		}

		devices, err := tenant.GetDevices()
		if err != nil {
			return fmt.Errorf("failed to get devices: %w", err)
		}

		fmt.Printf("Devices for tenant %s:\n", tenant.Name())
		for _, d := range devices {
			status, _ := d.Status()
			fmt.Printf("name=%s type=%s status=%s\n", d.Name(), d.Type(), status)
		}
		return nil
	},
}
