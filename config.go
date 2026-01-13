package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	defaultAPIURL          = "https://api.porkbun.com/api/json/v3"
	defaultCertificatePath = "/certs/{domain}/certificate.pem"
	defaultPrivateKeyPath  = "/certs/{domain}/private_key.pem"
	domainPlaceholder      = "{domain}"
)

type Configuration struct {
	Domain          string `mapstructure:"domain" validate:"required,min=1"`
	APIKey          string `mapstructure:"api_key" validate:"required,min=1"`
	SecretKey       string `mapstructure:"secret_key" validate:"required,min=1"`
	APIURL          string `mapstructure:"api_url" validate:"required,min=1"`
	CertificatePath string `mapstructure:"certificate_path" validate:"required,min=1"`
	PrivateKeyPath  string `mapstructure:"private_key_path" validate:"required,min=1"`
	CronSchedule    string `mapstructure:"cron_schedule" validate:"required,min=1"`
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

	_ = v.BindEnv("domain", "DOMAIN")
	_ = v.BindEnv("api_key", "API_KEY")
	_ = v.BindEnv("secret_key", "SECRET_KEY")
	_ = v.BindEnv("api_url", "API_URL")
	_ = v.BindEnv("certificate_path", "CERTIFICATE_PATH")
	_ = v.BindEnv("private_key_path", "PRIVATE_KEY_PATH")
	_ = v.BindEnv("cron_schedule", "CRON_SCHEDULE")
}

// Validate application configuration.
func (c *Configuration) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
