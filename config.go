package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	defaultAPIURL           = "https://api.porkbun.com/api/json/v3"
	defaultCertificatePath  = "/certs/{domain}/certificate.pem"
	defaultPrivateKeyPath   = "/certs/{domain}/private_key.pem"
	defaultCronSchedule     = "0 2 * * 1" // Every Monday at 2 AM
	defaultCombinedCertPath = ""          // Empty means no combined file
	domainPlaceholder       = "{domain}"
)

type Configuration struct {
	Domain           string `mapstructure:"domain" validate:"required,min=1"`
	APIKey           string `mapstructure:"api_key" validate:"required,min=1"`
	SecretKey        string `mapstructure:"secret_key" validate:"required,min=1"`
	APIURL           string `mapstructure:"api_url"`
	CertificatePath  string `mapstructure:"certificate_path"`
	PrivateKeyPath   string `mapstructure:"private_key_path"`
	CronSchedule     string `mapstructure:"cron_schedule"`
	CombinedCertPath string `mapstructure:"combined_cert_path"`
}

func NewConfiguration() (*Configuration, error) {
	v := viper.New()
	v.AutomaticEnv()

	config := &Configuration{}
	config.Bind("", v)

	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Bind configuration to viper.
func (c *Configuration) Bind(_ string, v *viper.Viper) {

	v.SetDefault("api_url", defaultAPIURL)
	v.SetDefault("certificate_path", defaultCertificatePath)
	v.SetDefault("private_key_path", defaultPrivateKeyPath)
	v.SetDefault("cron_schedule", defaultCronSchedule)
	v.SetDefault("combined_cert_path", defaultCombinedCertPath)

	_ = v.BindEnv("domain", "DOMAIN")
	_ = v.BindEnv("api_key", "API_KEY")
	_ = v.BindEnv("secret_key", "SECRET_KEY")
	_ = v.BindEnv("api_url", "API_URL")
	_ = v.BindEnv("certificate_path", "CERTIFICATE_PATH")
	_ = v.BindEnv("private_key_path", "PRIVATE_KEY_PATH")
	_ = v.BindEnv("cron_schedule", "CRON_SCHEDULE")
	_ = v.BindEnv("combined_cert_path", "COMBINED_CERT_PATH")
}

// Validate application configuration.
func (c *Configuration) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
