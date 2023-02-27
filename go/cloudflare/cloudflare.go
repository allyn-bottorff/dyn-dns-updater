package cloudflare

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var CloudFlareZonesURL string = "https://api.cloudflare.com/client/v4/zones"

type creds struct {
	Token string `json:"token"`
}

type zoneResult struct {
	Result []zone `json:"result"`
}

type zone struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type zoneDNSResponse struct {
	Success bool     `json:"success"`
	Result  []record `json:"result"`
}

type record struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}


// Get all of the DNS zones which are accessible by the token.
func GetZones(token string) []zone {

	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, CloudFlareZonesURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	var zr zoneResult

	err = json.NewDecoder(resp.Body).Decode(&zr)

	if err != nil {
		log.Fatal(err)
	}

	return zr.Result
}


// Get the apex A record of a zone
func getApex(z zone) (string, error) {
	var ipAddr string
	var zoneDNS zoneDNSResponse
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/dns_records", CloudFlareZonesURL, z.ID), nil)
	if err != nil {
		return ipAddr, err
	}
	resp, err := client.Do(request)
	if err != nil {
		return ipAddr, err
	}

	err = json.NewDecoder(resp.Body).Decode(&zoneDNS)
	if err != nil {
		return ipAddr, err
	}

	for _, record := range zoneDNS.Result {
		if record.Name == z.Name {
			if record.Type == "A" {
				ipAddr = record.Content
				break
			}
		}
	}
	return ipAddr, err

}

func patchApex(z zone, r record, ipAddr string) error {
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s/dns_records/%s", CloudFlareZonesURL, z.ID, r.ID), nil)


	client.Do(request)


	return err

}
