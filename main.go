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

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
