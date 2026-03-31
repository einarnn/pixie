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

// a template for creating an SGT
var templateSgt = `{
  "Sgt" : {
	"name" : "%s",
	"description" : "%s",
	"value" : %d
  }
}`

var createSgtCmd = &cobra.Command{
	Use:   "create-sgt",
	Short: "Create an SGT on the specified device",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, tenant, err := setup()
		if err != nil {
			return err
		}

		deviceName := viper.GetString("create-sgt.device")
		if deviceName == "" {
			return fmt.Errorf("device name is required (use --device)")
		}
		sgtName := viper.GetString("create-sgt.name")
		if sgtName == "" {
			return fmt.Errorf("SGT name is required (use --name)")
		}
		sgtDescription := viper.GetString("create-sgt.description")
		if sgtDescription == "" {
			return fmt.Errorf("SGT description is required (use --description)")
		}
		sgtValue := viper.GetInt("create-sgt.tag")
		if sgtValue == 0 {
			return fmt.Errorf("SGT tag is required and must be non-zero (use --tag)")
		}
		foundDevice, err := findDevice(tenant, deviceName)
		if err != nil {
			return fmt.Errorf("device %s not found", viper.GetString("create-sgt.device"))
		}

		// create SGT
		payload := fmt.Sprintf(templateSgt, sgtName, sgtDescription, sgtValue)
		req, _ := http.NewRequest(
			http.MethodPost,
			"/ers/config/sgt",
			strings.NewReader(payload),
		)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		resp, err := foundDevice.Query(req)
		if err != nil {
			logger.Infof("Failed to invoke %s on %s: %v", req, foundDevice.Name(), err)
		} else {
			bytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("query completed: status=%s\n", resp.Status)
			var result any
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
