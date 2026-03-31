package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cisco-pxgrid/cloud-sdk-go/log"
	"gopkg.in/yaml.v2"

	sdk "github.com/cisco-pxgrid/cloud-sdk-go"
)

var logger *log.DefaultLogger = &log.DefaultLogger{Level: log.LogLevelInfo}

type appConfig struct {
	Id           string `yaml:"id"`
	ApiKey       string `yaml:"apiKey"`
	GlobalFQDN   string `yaml:"globalFQDN"`
	RegionalFQDN string `yaml:"regionalFQDN"`
	ReadStream   string `yaml:"readStream"`
	WriteStream  string `yaml:"writeStream"`
}

type tenantConfig struct {
	Otp   string `yaml:"otp"`
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

type config struct {
	App    appConfig    `yaml:"app"`
	Tenant tenantConfig `yaml:"tenant"`
}

func messageHandler(id string, d *sdk.Device, stream string, p []byte) {
	logger.Infof("Message received. tenant=%s device=%s stream=%s", d.Tenant().Name(), d.Name(), stream)
	SendJsonToInfo("message:", p)
}

func activationHandler(d *sdk.Device) {
	status, _ := d.Status()
	logger.Infof("Device activation: %v", d.Name())
	logger.Infof("  Region : %v", d.Region())
	logger.Infof("  Tenant : %v", d.Tenant().Name())
	logger.Infof("  Type   : %v", d.Type())
	logger.Infof("  Status : %v", status)
}

func deActivationHandler(d *sdk.Device) {
	status, _ := d.Status()
	logger.Infof("Device deactivation: %v", d.Name())
	logger.Infof("  Region : %v", d.Region())
	logger.Infof("  Tenant : %v", d.Tenant().Name())
	logger.Infof("  Type   : %v", d.Type())
	logger.Infof("  Status : %v", status)
}

func loadConfig(file string) (*config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c := config{}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *config) store(file string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

func findDevice(tenant *sdk.Tenant, deviceName string) (*sdk.Device, error) {
	devices, err := tenant.GetDevices()
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		if d.Name() == deviceName {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("device not found: %s", deviceName)
}

func main() {

	// Load config
	configFile := flag.String("config", "", "Configuration yaml file to use (required)")
	deviceName := flag.String("device", "", "Device name")
	debug := flag.Bool("debug", false, "Enable debug output")
	insecure := flag.Bool("insecure", false, "Insecure TLS")
	flag.Parse()
	config, err := loadConfig(*configFile)
	if err != nil {
		panic(err)
	}

	// Set logger
	log.Logger = logger
	if *debug {
		logger.Level = log.LogLevelDebug
	}

	// Log after set logger
	logger.Debugf("Config: %+v", config)

	// check we have a device
	if *deviceName == "" {
		logger.Errorf("Device name is required")
		os.Exit(-1)
	}
	
	// SDK App config
	getCredentials := func() (*sdk.Credentials, error) {
		return &sdk.Credentials{
			ApiKey: []byte(config.App.ApiKey),
		}, nil
	}
	appConfig := sdk.Config{
		ID:                        config.App.Id,
		GetCredentials:            getCredentials,
		GlobalFQDN:                config.App.GlobalFQDN,
		RegionalFQDN:              config.App.RegionalFQDN,
		DeviceActivationHandler:   activationHandler,
		DeviceDeactivationHandler: deActivationHandler,
		DeviceMessageHandler:      messageHandler,
		ReadStreamID:              config.App.ReadStream,
		WriteStreamID:             config.App.WriteStream,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: *insecure,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}

	// SDK App create
	app, err := sdk.New(appConfig)
	if err != nil {
		panic(err)
	}
	logger.Debugf("App config: %+v", appConfig)

	var tc = &config.Tenant
	var tenant *sdk.Tenant
	if tc.Otp != "" {

		// **first time** need to link tenant using OTP in config file
		tenant, err = app.LinkTenant(tc.Otp)
		if err != nil {
			logger.Errorf("Failed to link tenant: %v", err)
			os.Exit(-1)
		}
		tc.Otp = ""
		tc.ID = tenant.ID()
		tc.Name = tenant.Name()
		tc.Token = tenant.ApiToken()
		config.store(*configFile)

	} else {

		// SDK set tenant with existing id, name and token
		tenant, err = app.SetTenant(tc.ID, tc.Name, tc.Token)
		if err != nil {
			logger.Errorf("Failed to set tenant to app: %v", err)
			syscall.Exit(1)
		}
	}
	logger.Infof("Linked with tenant: %s, id=%s", tenant.Name(), tenant.ID())

	// Catch termination signal
	_, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- do other stuff ---

	//but only after 5 seconds...
	logger.Infof("*** pausing for a few seconds...")
	time.Sleep(time.Second * 5)

	// find the device I'm interested in
	foundDevice, err := findDevice(tenant, *deviceName)
	if err != nil {
		logger.Errorf("Failed to find device: %v", err)
		syscall.Exit(1)
	}

	logger.Infof("*** starting SDK-based API calls")
	logger.Infof(
		"invoking API on device = %v, %v, %v, %v",
		foundDevice.Name(), 
		foundDevice.ID(), 
		tenant.ID(),
		tenant.ApiToken(),
	)

	// make an echo request
	// req, _ := http.NewRequest(
	// 	http.MethodPost,
	// 	"/pxgrid/echo/query",
	// 	strings.NewReader(`{"a":"b"}`),
	// )

	// get SGTs
	// req, _ := http.NewRequest(
	// 	http.MethodPost,
	// 	"/pxgrid/trustsec/getSecurityGroups",
	// 	strings.NewReader("{}"),
	// )

	// get SXP local bindings
	// req, _ := http.NewRequest(
	// 	http.MethodGet, 
	// 	"/ers/config/sxplocalbindings/eeffc5e7-75d4-4780-b393-bafb7019f3ad", 
	// 	nil,
	// )
	// req.Header.Add("Accept", "application/json")

	// get deployment info
	// req, _ := http.NewRequest(
	// 	http.MethodGet, 
	// 	"/api/v1/deployment/node", 
	// 	nil,
	// )
	// req.Header.Add("Accept", "application/json")

	// get the first N network devices
	// req, _ := http.NewRequest(
	// 	http.MethodGet,
	// 	"/ers/config/networkdevice",
	// 	nil,
	// )
	// req.Header.Add("Accept", "application/json")


	// ** doesn't work **
	// make an MNT request
	// req, _ := http.NewRequest(
	// 	http.MethodGet,
	// 	"/admin/API/mnt/Session/ActiveCount",
	// 	nil,
	// )
	// req.Header.Add("Accept", "application/xml")


	// get deployment info
	req, _ := http.NewRequest(
		http.MethodGet,
		"/api/v1/deployment/node",
		nil,
	)
	req.Header.Add("Accept", "application/json")

	resp, err := foundDevice.Query(req)
	if err != nil {
		logger.Infof("Failed to invoke %s on %s: %v", req, foundDevice.Name(), err)
	} else {
		bytes, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("query completed: status=%s", resp.Status)
		SendJsonToInfo(msg, bytes)
	}
	logger.Infof("*** finished SDK-based API calls")


	// get deployment info
	req, _ = http.NewRequest(
			http.MethodPost,
			"/pxgrid/trustsec/getBindings",
			strings.NewReader("{}"),
	)
	req.Header.Add("Accept", "application/json")

	resp, err = foundDevice.Query(req)
	if err != nil {
		logger.Infof("Failed to invoke %s on %s: %v", req, foundDevice.Name(), err)
	} else {
		bytes, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("query completed: status=%s", resp.Status)
		SendJsonToInfo(msg, bytes)
	}
	logger.Infof("*** finished SDK-based API calls")


	// 	// and just echo forever...
	// 	go func(device sdk.Device) {
	// 		for {
	// 			time.Sleep(5 * time.Second)
	// 			Echo(device)
	// 		}
	// 	}(d)

	// }

	// make a DIRECT API request, not using the SDK
	logger.Infof("*** starting direct API calls")
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	logger.Infof("*** calling APIs against %s", foundDevice.Name())
	type apiCall struct {
		api    string
		method string
		body   io.Reader
		setHeader func(req *http.Request)
	}
	apisToCall := []apiCall{
		// ** doesn't work **
		// {
		// 	api:    "/admin/API/mnt/Session/ActiveCount",
		// 	method: http.MethodGet,
		// 	body:   nil,
		// 	setHeader: func(req *http.Request) {
		// 		req.Header.Add("Accept", "application/xml")
		// 		req.Header.Add("x-api-key", foundDevice.Tenant().ApiToken())
		// 		req.Header.Add("X-API-PROXY-COMMUNICATION-STYLE", "sync")
		// 	},
		// },
		{
			api:    "/ers/config/sxplocalbindings",
			method: http.MethodGet,
			body:   nil,
		},
		{
			api:    "/ers/config/sgmapping",
			method: http.MethodGet,
			body:   nil,
		},
		// {
		// 	api:    "/ers/config/networkdevice",
		// 	method: http.MethodGet,
		// 	body:   nil,
		// },
		// {
		// 	api:    "/pxgrid/trustsec/getSecurityGroups",
		// 	method: http.MethodPost,
		// 	body:   strings.NewReader("{}"),
		// },
		// {
		// 	api:    "/pxgrid/trustsec/getSessions",
		// 	method: http.MethodPost,
		// 	body:   strings.NewReader("{}"),
		// },
	}
	// formulate a request
	for _, api := range apisToCall {
		logger.Infof("calling %s", api.api)
		req, _ := http.NewRequest(
			api.method,
			fmt.Sprintf(
				"https://%s/api/dxhub/v2/apiproxy/request/%s/direct%s",
				// "pxgridcloud.cisco.com",
				"neoffers.cisco.com",
				foundDevice.ID(),
				api.api,
			),
			api.body)
		if api.setHeader != nil {
			api.setHeader(req)
		} else {
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("x-api-key", foundDevice.Tenant().ApiToken())
			req.Header.Add("X-API-PROXY-COMMUNICATION-STYLE", "sync")
		}		
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		bytes, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("query completed: status=%s", resp.Status)
		SendJsonToInfo(msg, bytes)
	}
	logger.Infof("*** finished direct API calls")

	// --- done with other stuff ---
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	select {
	case <-ctx.Done():
		logger.Infof("Terminating...")
	case err := <-app.Error:
		logger.Errorf("App error: %v", err)
	}
	if err = app.Close(); err != nil {
		panic(err)
	}
}

// Take a byte array that contains JSON, decode it, prettify it and send it
// line-by-line to log.Infof
func SendJsonToInfo(msg string, b []byte) {
	var decoded interface{}
	json.Unmarshal(b, &decoded)
	pretty, _ := json.MarshalIndent(decoded, "", "  ")
	logger.Infof("message=%s", msg)
	for _, line := range strings.Split(string(pretty), "\n") {
		logger.Infof("%s", line)
	}

}

func Echo(device sdk.Device) {
	logger.Infof("[%s] starting echo...", device.Name())

	// Perform echo-query
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"/pxgrid/echo/query",
		strings.NewReader("data to echo"))
	resp, err := device.Query(req)
	if err != nil {
		logger.Infof("[%s] echo request failed: %+v", device.Name(), err)
		return
	}
	defer resp.Body.Close()

	// log reply output
	bytes, _ := io.ReadAll(resp.Body)
	stringResp := string(bytes)
	logger.Infof(
		"[%s] Query completed. status=%s bodyLen=%d, body=%s\n",
		device.Name(),
		resp.Status,
		len(stringResp),
		stringResp)
}
