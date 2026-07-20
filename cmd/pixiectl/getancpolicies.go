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

var getAncPoliciesCmd = &cobra.Command{
	Use:   "get-anc-policies",
	Short: "List all established sessions for the specified device",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, tenant, err := setup()
		if err != nil {
			return err
		}

		deviceName := viper.GetString("get-anc-policies.device")
		if deviceName == "" {
			return fmt.Errorf("device name is required (use --device)")
		}
		foundDevice, err := findDevice(tenant, deviceName)
		if err != nil {
			return fmt.Errorf("device %s not found", viper.GetString("get-anc-policies.device"))
		}

		// get SGTs
		// fmt.Printf("SGTs for device %s:\n", deviceName)
		req, _ := http.NewRequest(
			http.MethodPost,
			"/pxgrid/anc/getPolicies",
			strings.NewReader("{}"),
		)
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
