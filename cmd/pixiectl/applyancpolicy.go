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

var applyAncPolicyCmd = &cobra.Command{
	Use:   "apply-anc-policy",
	Short: "Apply a named ANC policy to a client by MAC address",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, tenant, err := setup()
		if err != nil {
			return err
		}

		deviceName := viper.GetString("apply-anc-policy.device")
		if deviceName == "" {
			return fmt.Errorf("device name is required (use --device)")
		}
		foundDevice, err := findDevice(tenant, deviceName)
		if err != nil {
			return fmt.Errorf("device %s not found", viper.GetString("apply-anc-policy.device"))
		}
		policyName := viper.GetString("apply-anc-policy.name")
		if policyName == "" {
			return fmt.Errorf("ANC policy name is required (use --name)")
		}
		macAddress := viper.GetString("apply-anc-policy.mac")
		if macAddress == "" {
			return fmt.Errorf("MAC address is required (use --mac)")
		}

		var policyTemplate = `{"policyName": "%s","macAddress": "%s"}`
		req, _ := http.NewRequest(
			http.MethodPost,
			"/pxgrid/anc/applyEndpointByMacAddress",
			strings.NewReader(fmt.Sprintf(policyTemplate, policyName, macAddress)),
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		resp, err := foundDevice.Query(req)
		if err != nil {
			logger.Infof("Failed to invoke %s on %s: %v", req, foundDevice.Name(), err)
		} else {
			bytes, _ := io.ReadAll(resp.Body)
			// fmt.Printf("query completed: status=%s\n", resp.Status)
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
