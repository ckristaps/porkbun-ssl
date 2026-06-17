package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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

	if a.config.CombinedCertPath != "" {
		combinedPath := strings.ReplaceAll(a.config.CombinedCertPath, domainPlaceholder, a.config.Domain)
		combinedContent := data.CertificateChain + "\n" + data.PrivateKey

		incomingCert, err := parseCertificate(data.CertificateChain)
		if err != nil {
			return fmt.Errorf("failed to parse incoming certificate: %w", err)
		}

		log.Printf("[INFO] incoming certificate valid from: %s, valid until: %s",
			incomingCert.NotBefore.Format(time.RFC3339), incomingCert.NotAfter.Format(time.RFC3339))

		if _, err := os.Stat(combinedPath); err == nil {
			existingContent, err := os.ReadFile(combinedPath)
			if err == nil {
				existingCert, err := parseCertificate(string(existingContent))
				if err == nil {
					log.Printf("[INFO] existing certificate valid from: %s, valid until: %s",
						existingCert.NotBefore.Format(time.RFC3339), existingCert.NotAfter.Format(time.RFC3339))

					if !incomingCert.NotBefore.After(existingCert.NotBefore) {
						log.Printf("[INFO] existing certificate is same or newer, skipping save")
						return nil
					}
					log.Printf("[INFO] incoming certificate is newer, updating")
				}
			}
		}

		log.Printf("[INFO] saving combined certificate to %s", combinedPath)
		if err := saveFile(combinedPath, combinedContent); err != nil {
			return fmt.Errorf("failed to save combined certificate: %w", err)
		}

		return nil
	}

	certPath := strings.ReplaceAll(a.config.CertificatePath, domainPlaceholder, a.config.Domain)

	incomingCert, err := parseCertificate(data.CertificateChain)
	if err != nil {
		return fmt.Errorf("failed to parse incoming certificate: %w", err)
	}

	log.Printf("[INFO] incoming certificate valid from: %s, valid until: %s",
		incomingCert.NotBefore.Format(time.RFC3339), incomingCert.NotAfter.Format(time.RFC3339))

	if _, err := os.Stat(certPath); err == nil {
		existingContent, err := os.ReadFile(certPath)
		if err == nil {
			existingCert, err := parseCertificate(string(existingContent))
			if err == nil {
				log.Printf("[INFO] existing certificate valid from: %s, valid until: %s",
					existingCert.NotBefore.Format(time.RFC3339), existingCert.NotAfter.Format(time.RFC3339))

				if !incomingCert.NotBefore.After(existingCert.NotBefore) {
					log.Printf("[INFO] existing certificate is same or newer, skipping save")
					return nil
				}
				log.Printf("[INFO] incoming certificate is newer, updating")
			}
		}
	}

	log.Printf("[INFO] saving certificate to %s", certPath)
	if err := saveFile(certPath, data.CertificateChain); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

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

func parseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}
