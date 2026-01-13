package main

import (
	"log"
)

type App struct {
	config *Configuration
}

type Config struct {
	Domains         []string
	APIKey          string
	SecretKey       string
	APIURL          string
	CertificatePath string
	PrivateKeyPath  string
	CronSchedule    string
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

	config, err := NewConfiguration()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	app := &App{config: config}

	app.runWithSchedule()
}
