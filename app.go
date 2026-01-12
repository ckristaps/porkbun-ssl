package porkbunssl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	defaultAPIURL          = "https://api.porkbun.com/api/json/v3"
	defaultCertificatePath = "/certs/{domain}/certificate.pem"
	defaultPrivateKeyPath  = "/certs/{domain}/private_key.pem"
	domainPlaceholder      = "{domain}"
)

type Config struct {
	Domains            []string
	APIKey             string
	SecretKey          string
	APIURL             string
	CertificatePathTpl string
	PrivateKeyPathTpl  string
	CronSchedule       string
}

type PorkbunRequest struct {
	APIKey       string `json:"apikey"`
	SecretAPIKey string `json:"secretapikey"`
}

type PorkbunResponse struct {
	Status           string `json:"status"`
	Message          string `json:"message,omitempty"`
	CertificateChain string `json:"certificatechain,omitempty"`
	PrivateKey       string `json:"privatekey,omitempty"`
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime)

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	if err := validateConfig(config); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	runWithSchedule(config)
}

func runWithSchedule(config *Config) {
	log.Printf("[INFO] starting SSL certificate renewal scheduler with schedule: %s", config.CronSchedule)

	c := cron.New()
	_, err := c.AddFunc(config.CronSchedule, func() {
		runRenewal(config)
	})

	if err != nil {
		log.Fatalf("[ERROR] invalid cron schedule '%s': %v", config.CronSchedule, err)
	}

	c.Start()
	log.Println("[INFO] scheduler started, waiting for scheduled runs...")

	// Run once immediately on startup
	runRenewal(config)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[INFO] shutting down scheduler...")
	c.Stop()
}

func runRenewal(config *Config) {
	log.Println("[INFO] starting certificate renewal...")

	for _, domain := range config.Domains {
		if err := processDomain(domain, config); err != nil {
			log.Printf("[ERROR] failed to process domain %s: %v", domain, err)
			continue
		}
		log.Printf("[INFO] SSL certificate for %s has been renewed", domain)
	}

	log.Println("[INFO] all certificates renewed successfully")
}

func loadConfig() (*Config, error) {
	domains := os.Getenv("DOMAIN")
	if domains == "" {
		return nil, fmt.Errorf("DOMAIN environment variable is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is required")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is required")
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	certPath := os.Getenv("CERTIFICATE_PATH")
	if certPath == "" {
		certPath = defaultCertificatePath
	}

	privKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	if privKeyPath == "" {
		privKeyPath = defaultPrivateKeyPath
	}

	domainList := strings.Split(domains, ",")
	for i, d := range domainList {
		domainList[i] = strings.TrimSpace(d)
	}

	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		return nil, fmt.Errorf("CRON_SCHEDULE environment variable is required")
	}

	return &Config{
		Domains:            domainList,
		APIKey:             apiKey,
		SecretKey:          secretKey,
		APIURL:             apiURL,
		CertificatePathTpl: certPath,
		PrivateKeyPathTpl:  privKeyPath,
		CronSchedule:       cronSchedule,
	}, nil
}

func validateConfig(config *Config) error {
	if len(config.Domains) > 1 {
		if !strings.Contains(config.CertificatePathTpl, domainPlaceholder) {
			return fmt.Errorf("CERTIFICATE_PATH must contain the %s placeholder when multiple domains are specified", domainPlaceholder)
		}
		if !strings.Contains(config.PrivateKeyPathTpl, domainPlaceholder) {
			return fmt.Errorf("PRIVATE_KEY_PATH must contain the %s placeholder when multiple domains are specified", domainPlaceholder)
		}
	}
	return nil
}

func processDomain(domain string, config *Config) error {
	log.Printf("[INFO] downloading SSL bundle for %s", domain)

	// Make API request
	data, err := fetchSSLBundle(domain, config)
	if err != nil {
		return err
	}

	// Save certificate
	certPath := strings.ReplaceAll(config.CertificatePathTpl, domainPlaceholder, domain)
	log.Printf("[INFO] saving certificate to %s", certPath)
	if err := saveFile(certPath, data.CertificateChain); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	// Save private key
	privKeyPath := strings.ReplaceAll(config.PrivateKeyPathTpl, domainPlaceholder, domain)
	log.Printf("[INFO] saving private key to %s", privKeyPath)
	if err := saveFile(privKeyPath, data.PrivateKey); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	return nil
}

func fetchSSLBundle(domain string, config *Config) (*PorkbunResponse, error) {
	url := fmt.Sprintf("%s/ssl/retrieve/%s", config.APIURL, domain)

	reqBody := PorkbunRequest{
		APIKey:       config.APIKey,
		SecretAPIKey: config.SecretKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var porkbunResp PorkbunResponse
	if err := json.Unmarshal(body, &porkbunResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if porkbunResp.Status == "ERROR" {
		return nil, fmt.Errorf("API error: %s", porkbunResp.Message)
	}

	return &porkbunResp, nil
}

func saveFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
