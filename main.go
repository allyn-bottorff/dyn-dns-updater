package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	// "net/http"
	// "fmt"
	// "net/http"
	"log"
	"os"
)

type Secrets struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
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

func main() {
	// https://unifi.b6f.net/api/login
	// https://unifi.b6f.net/api/s/default/stat/health

	secrets, err := getSecrets()




	log.Println("Logging into unifi")
	resp, err := http.Post("https://unifi.b6f.net/api/login", "application/json", bytes.NewReader(credsfile))

	if err != nil {
		log.Panic(err)
	}
	log.Printf("Login results: %v", resp.Status)

	client := &http.Client{}

	request, err := http.NewRequest(http.MethodGet, "https://unifi.b6f.net/api/s/default/stat/health", nil)
	request.AddCookie(resp.Cookies()[0])

	healthResp, err := client.Do(request)

	var unifiHealth UnifiHealth

	err = json.NewDecoder(healthResp.Body).Decode(&unifiHealth)
	if err != nil {
		log.Panic(err)
	}

	// file, err := os.ReadFile("unifi-site-health.json")
	// if err != nil {
	// 	log.Panic(err)
	// }
	// var unifiHealth UnifiHealth
	// err = json.Unmarshal(file, &unifiHealth)
	// if err != nil {
	// 	log.Panic(err)
	// }

	var ipAddr string

	for _, system := range unifiHealth.Data {
		if system.SubSystem == "wan" {
			ipAddr = system.WanIP
		}
	}

	log.Print(ipAddr)

}
