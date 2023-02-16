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

func GetZones(token string) []zone{

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
