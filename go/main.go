package main

import (
	"encoding/json"
	"log"
	"os"

	"gihub.com/allyn-bottorff/dyn-dns-updater/unifi"
)

type Secrets struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type Config struct {
	UnifiDomain    string   `json:"unifiDomain"`
	UnifiSiteName  string   `json:"unifiSiteName"`
	ManagedDomains []string `json:"managedDomains"`
}

// Read credentials from kubernetes secrets files as json
func getSecrets() (Secrets, error) {
	var secrets Secrets
	secretsFile, err := os.ReadFile("secrets.json")
	if err != nil {
		return secrets, err
	}

	err = json.Unmarshal(secretsFile, &secrets)

	return secrets, err
}

func getConfig() (Config, error) {
	var config Config
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configFile, &config)
	if config.UnifiSiteName == "" {
		config.UnifiSiteName = "default"
	}
	return config, err
}

func main() {

	secrets, err := getSecrets()
	if err != nil {
		log.Panicf("Failed to read secrets: %v", err)
	}

	config, err := getConfig()
	if err != nil {
		log.Panicf("Failed to read config: %v", err)
	}

	// Find Public IP
	publicIP := unifi.GetLocalIP(
		secrets.Username,
		secrets.Password,
		config.UnifiDomain,
		config.UnifiSiteName,
	)

	log.Print(publicIP)

}
