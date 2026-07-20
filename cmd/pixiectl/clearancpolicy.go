package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clearAncPolicyCmd = &cobra.Command{
	Use:   "clear-anc-policy",
	Short: "Clear an ANC policy from a client by MAC address",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, tenant, err := setup()
		if err != nil {
			return err
		}

		deviceName := viper.GetString("clear-anc-policy.device")
		if deviceName == "" {
			return fmt.Errorf("device name is required (use --device)")
		}
		foundDevice, err := findDevice(tenant, deviceName)
		if err != nil {
			return fmt.Errorf("device %s not found", viper.GetString("clear-anc-policy.device"))
		}
		macAddress := viper.GetString("clear-anc-policy.mac")
		if macAddress == "" {
			return fmt.Errorf("MAC address is required (use --mac)")
		}

		var policyTemplate = `{"macAddress": "%s"}`
		req, _ := http.NewRequest(
			http.MethodPost,
			"/pxgrid/anc/clearEndpointByMacAddress",
			strings.NewReader(fmt.Sprintf(policyTemplate, macAddress)),
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		resp, err := foundDevice.Query(req)
		if err != nil {
			logger.Infof("Failed to invoke %s on %s: %v", req, foundDevice.Name(), err)
		} else {
			bytes, _ := io.ReadAll(resp.Body)
			var result interface{}
			if err := json.Unmarshal(bytes, &result); err != nil {
				logger.Errorf("Failed to parse JSON: %v", err)
			} else {
				prettyJSON, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(prettyJSON))
			}
		}
		return nil
	},
}
