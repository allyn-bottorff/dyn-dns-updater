package unifi

import (
	"bytes"
	"net/http"
	"encoding/json"
	"log"
)

var UnifiLoginURL string = "https://unifi.b6f.net/api/login"
var UnifiHealthURL string = "https://unifi.b6f.net/api/s/default/stat/health"

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

func (u UnifiCreds) JsonBytes() []byte {
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		log.Panicf("Failed to marshal Unifi credentials JSON: %v",err)
	}
	return jsonBytes
}





func getLocalIP(username string, password string) string {
	creds := UnifiCreds {
		Username: username,
		Password: password,
	}

	// Log into Unifi Controller
	loginResp, err := http.Post(UnifiLoginURL, "application/json", bytes.NewReader(creds.JsonBytes()))
	if err != nil {
		log.Panicf("Failed to log into Unifi: %v", err)
	}
	if loginResp.StatusCode != 200 {
		log.Panicf("Failed to log into Unifi. Unifi response: %v", loginResp.Status)
	}

	// Get health of default site from Unifi Controller
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, UnifiHealthURL, nil)
	request.AddCookie(loginResp.Cookies()[0])

	healthResp, err := client.Do(request)
	if err != nil {
		log.Panicf("Failed to get health of default zone from Unifi Controller: %v", err)
	}
	if healthResp.StatusCode != 200 {
		log.Panicf("Failed to get health of default zone from Unifi Controller. Unifi response: %v", healthResp.Status)
	}

	var health UnifiHealth

	err = json.NewDecoder(healthResp.Body).Decode(&health)
	if err != nil {
		log.Panicf("Failed to unmarshal health repsonse from Unifi: %v", err)
	}

	var ipAddr string

	for _, system := range health.Data {
		if system.SubSystem == "wan" {
			ipAddr = system.WanIP
		}
	}

	return ipAddr
}
