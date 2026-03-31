package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/cisco-pxgrid/cloud-sdk-go/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	sdk "github.com/cisco-pxgrid/cloud-sdk-go"
	pixieutils "github.com/einarnn/pixie/internal/utils"
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
	pixieutils.SendJsonToInfo(logger, "message:", p)
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

// initConfig reads the config file specified by the --config flag.
func initConfig() (*config, error) {
	configFile := viper.GetString("config")
	if configFile == "" {
		return nil, fmt.Errorf("config file is required (use --config)")
	}
	return loadConfig(configFile)
}

// initLogger sets up the logger based on the --debug or --info flags.
func initLogger() {
	log.Logger = logger
	if viper.GetBool("info") {
		logger.Level = log.LogLevelInfo
	} else if viper.GetBool("debug") {
		logger.Level = log.LogLevelDebug
	} else {
		logger.Level = log.LogLevelError
	}
}

// newApp creates an SDK app from the given config.
func newApp(cfg *config) (*sdk.App, error) {
	getCredentials := func() (*sdk.Credentials, error) {
		return &sdk.Credentials{
			ApiKey: []byte(cfg.App.ApiKey),
		}, nil
	}
	appConfig := sdk.Config{
		ID:                        cfg.App.Id,
		GetCredentials:            getCredentials,
		GlobalFQDN:                cfg.App.GlobalFQDN,
		RegionalFQDN:              cfg.App.RegionalFQDN,
		DeviceActivationHandler:   activationHandler,
		DeviceDeactivationHandler: deActivationHandler,
		DeviceMessageHandler:      messageHandler,
		ReadStreamID:              cfg.App.ReadStream,
		WriteStreamID:             cfg.App.WriteStream,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: viper.GetBool("insecure"),
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
	app, err := sdk.New(appConfig)
	if err != nil {
		return nil, err
	}
	logger.Debugf("App config: %+v", appConfig)
	return app, nil
}

// linkOrSetTenant either links a new tenant via OTP or sets an existing one.
func linkOrSetTenant(app *sdk.App, cfg *config) (*sdk.Tenant, error) {
	tc := &cfg.Tenant
	if tc.Otp != "" {
		tenant, err := app.LinkTenant(tc.Otp)
		if err != nil {
			return nil, fmt.Errorf("failed to link tenant: %w", err)
		}
		tc.Otp = ""
		tc.ID = tenant.ID()
		tc.Name = tenant.Name()
		tc.Token = tenant.ApiToken()
		configFile := viper.GetString("config")
		if err := cfg.store(configFile); err != nil {
			return nil, fmt.Errorf("failed to store config: %w", err)
		}
		return tenant, nil
	}
	tenant, err := app.SetTenant(tc.ID, tc.Name, tc.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant: %w", err)
	}
	return tenant, nil
}

// setup initialises logging, loads the config, creates the SDK app,
// and links/sets the tenant. It is the common preamble for all commands.
func setup() (*sdk.App, *sdk.Tenant, error) {
	initLogger()
	cfg, err := initConfig()
	if err != nil {
		return nil, nil, err
	}
	logger.Debugf("Config: %+v", cfg)

	app, err := newApp(cfg)
	if err != nil {
		return nil, nil, err
	}

	tenant, err := linkOrSetTenant(app, cfg)
	if err != nil {
		return nil, nil, err
	}
	logger.Debugf("Linked with tenant: %s, id=%s", tenant.Name(), tenant.ID())
	return app, tenant, nil
}

var rootCmd = &cobra.Command{
	Use:   "pixiectl",
	Short: "CLI for managing pxGrid Cloud devices",
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "Configuration yaml file to use (required)")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")
	rootCmd.PersistentFlags().Bool("info", false, "Enable info output")
	rootCmd.PersistentFlags().Bool("insecure", false, "Insecure TLS")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("info", rootCmd.PersistentFlags().Lookup("info"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))

	runCmd.Flags().String("device", "", "Device name")
	viper.BindPFlag("run.device", runCmd.Flags().Lookup("device"))

	getSgtsCmd.Flags().String("device", "", "Device name")
	viper.BindPFlag("get-sgts.device", getSgtsCmd.Flags().Lookup("device"))

	createSgtCmd.Flags().String("device", "", "Device name")
	createSgtCmd.Flags().String("name", "", "SGT name")
	createSgtCmd.Flags().String("description", "", "SGT description")
	createSgtCmd.Flags().Int("tag", 0, "SGT value")
	viper.BindPFlag("create-sgt.device", createSgtCmd.Flags().Lookup("device"))
	viper.BindPFlag("create-sgt.name", createSgtCmd.Flags().Lookup("name"))
	viper.BindPFlag("create-sgt.description", createSgtCmd.Flags().Lookup("description"))
	viper.BindPFlag("create-sgt.tag", createSgtCmd.Flags().Lookup("tag"))

	deleteSgtCmd.Flags().String("device", "", "Device name")
	deleteSgtCmd.Flags().String("name", "", "SGT name")
	viper.BindPFlag("delete-sgt.device", deleteSgtCmd.Flags().Lookup("device"))
	viper.BindPFlag("delete-sgt.name", deleteSgtCmd.Flags().Lookup("name"))

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listDevicesCmd)
	rootCmd.AddCommand(getSgtsCmd)
	rootCmd.AddCommand(createSgtCmd)
	rootCmd.AddCommand(deleteSgtCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
