package main

import (
	"fmt"
	"os"

	sdk "github.com/cisco-pxgrid/cloud-sdk-go"
)

// activationHandler is invoked when the app gets activated for a new device for a tenant
func activationHandler(device *sdk.Device) {
	fmt.Printf("New device activated: %s\n", device)
}

// deactivationHandler is invoked when the app gets deactivated for an existing device for a tenant
func deactivationHandler(device *sdk.Device) {
	fmt.Printf("Device deactivated: %s\n", device)
}

// messageHandler is invoked when there's a new message received by the app for a device
func messageHandler(id string, device *sdk.Device, stream string, payload []byte) {
	fmt.Printf("Received new message (%s) from %s\n", id, device)
	fmt.Printf("Message stream: %s\n", stream)
	fmt.Printf("Message payload: %s\n", payload)
}

// getCredentials is invoked whenever the app needs to retrieve the app credentials
func getCredentials() (*sdk.Credentials, error) {
	return &sdk.Credentials{
		ApiKey: []byte("it1t7jJsYZiDb6YQycwBH7wHJMn424lOBKf1aZVOj0dNtkeNXDSW6RVCcYbbYALl"),
	}, nil
}

func main() {
	// create app configuration required for creating a new app instance
	config := sdk.Config{
		ID:                        "einar-demo-app",
		GetCredentials:            getCredentials,
		RegionalFQDN:              "us-west-2.cloud.cisco.com",
		GlobalFQDN:                "global.cloud.cisco.com",
		DeviceActivationHandler:   activationHandler,
		DeviceDeactivationHandler: deactivationHandler,
		DeviceMessageHandler:      messageHandler,
		ReadStreamID:              "app--einar-demo-app-R",
		WriteStreamID:             "app--einar-demo-app-W",
	}

	app, err := sdk.New(config)
	if err != nil {
		fmt.Printf("Failed to create new app: %v", err)
		os.Exit(1)
	}
	// make sure to Close() the app instance to release resources acquired during app creation
	defer func() {
		if err := app.Close(); err != nil {
			fmt.Printf("Disconnected with error: %v", err)
		}
	}()

	// wait for any errors from the app instance created above
	err = <-app.Error
	if err != nil {
		fmt.Printf("%s received an error: %v", app, err)
	}
}
