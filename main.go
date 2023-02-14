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

type UnifiHealth struct {
	// Meta map[string]map[string]string `json:"meta"`
	Data []SubsystemHealth `json:"data"`
}

type SubsystemHealth struct {
	SubSystem string `json:"subsystem"`
	WanIP     string `json:"wan_ip"`
}

type UnifiCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CloudFlareCreds struct {
	Token string `json:"token"`
}

type Credentials struct {
	UnifiCreds
	CloudFlareCreds

}

var UnifiLoginURL string = "https://unifi.b6f.net/api/login"
var UnifiHealthURL string = "https://unifi.b6f.net/api/s/default/stat/health"
var CloudFlareZonesURL string = "https://api.cloudflare.com/client/v4/zones"


// Read credentials from kubernetes secrets files as json
func getSecrets() (Credentials, error) {
	var creds Credentials
	unifiCredsFile, err := os.ReadFile("post.json")
	if err != nil {
		return creds, err
	}

	cloudFlareFile, err := os.ReadFile("token.json")
	if err != nil {
		return creds, err
	}

	err = json.Unmarshal(unifiCredsFile, &creds.UnifiCreds)
	err = json.Unmarshal(cloudFlareFile, &creds.CloudFlareCreds)

	return creds, err
}

func getCurrentIP(creds &UnifiCreds) string {

	creds

}

func main() {
	// https://unifi.b6f.net/api/login
	// https://unifi.b6f.net/api/s/default/stat/health

	credsfile, err := os.ReadFile("post.json")
	if err != nil {
		log.Panic(err)
	}

	var creds Credentials
	err = json.Unmarshal(credsfile, &creds)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Found username: %s", creds.Username)

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
