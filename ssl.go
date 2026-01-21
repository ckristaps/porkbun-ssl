package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) processDomain() error {
	log.Printf("[INFO] downloading SSL bundle for %s", a.config.Domain)

	// Make API request
	data, err := a.fetchSSLBundle(a.config.Domain)
	if err != nil {
		return err
	}

	// Save combined certificate + private key if path is configured
	if a.config.CombinedCertPath != "" {
		combinedPath := strings.ReplaceAll(a.config.CombinedCertPath, domainPlaceholder, a.config.Domain)
		log.Printf("[INFO] saving combined certificate to %s", combinedPath)
		combinedContent := data.CertificateChain + "\n" + data.PrivateKey
		if err := saveFile(combinedPath, combinedContent); err != nil {
			return fmt.Errorf("failed to save combined certificate: %w", err)
		}

		return nil
	}

	// Save certificate
	certPath := strings.ReplaceAll(a.config.CertificatePath, domainPlaceholder, a.config.Domain)
	log.Printf("[INFO] saving certificate to %s", certPath)
	if err := saveFile(certPath, data.CertificateChain); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	// Save private key
	privKeyPath := strings.ReplaceAll(a.config.PrivateKeyPath, domainPlaceholder, a.config.Domain)
	log.Printf("[INFO] saving private key to %s", privKeyPath)
	if err := saveFile(privKeyPath, data.PrivateKey); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	return nil
}

func (a *App) fetchSSLBundle(domain string) (*PorkbunResponse, error) {
	url := fmt.Sprintf("%s/ssl/retrieve/%s", a.config.APIURL, domain)

	reqBody := PorkbunRequest{
		APIKey:       a.config.APIKey,
		SecretAPIKey: a.config.SecretKey,
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
