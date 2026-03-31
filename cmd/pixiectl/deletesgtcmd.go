package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteSgtCmd = &cobra.Command{
	Use:   "delete-sgt",
	Short: "Delete an SGT by name from the specified device",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, tenant, err := setup()
		if err != nil {
			return err
		}

		deviceName := viper.GetString("delete-sgt.device")
		if deviceName == "" {
			return fmt.Errorf("device name is required (use --device)")
		}
		sgtName := viper.GetString("delete-sgt.name")
		if sgtName == "" {
			return fmt.Errorf("SGT name is required (use --name)")
		}
		foundDevice, err := findDevice(tenant, deviceName)
		if err != nil {
			return fmt.Errorf("device %s not found", viper.GetString("delete-sgt.device"))
		}

		// Get the SGT ID by name (ERS API doesn't support delete by name, only by ID)
		req, _ := http.NewRequest(
			http.MethodGet,
			"/ers/config/sgt/name/"+sgtName,
			strings.NewReader(""),
		)
		req.Header.Add("Accept", "application/json")
		resp, err := foundDevice.Query(req)
		if err != nil {
			return fmt.Errorf("failed to get SGT: %w", err)
		}

		var result struct {
			Sgt struct {
				ID string `json:"id"`
			} `json:"Sgt"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse SGT response: %w", err)
		}
		sgtID := result.Sgt.ID

		// delete SGT
		req, _ = http.NewRequest(
			http.MethodDelete,
			"/ers/config/sgt/"+sgtID,
			strings.NewReader(""),
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		resp, err = foundDevice.Query(req)
		if err != nil {
			logger.Infof("Failed to invoke %s on %s: %v", req, foundDevice.Name(), err)
		} else {
			logger.Infof("Delete SGT response: status=%s", resp.Status)
		}
		return nil
	},
}
