package cloudflare

var CloudFlareZonesURL string = "https://api.cloudflare.com/client/v4/zones"


type CloudflareCreds struct {
	Token string `json:"token"`
}


type CloudflareZoneResult struct {
	Result []CloudflareZone `json:"result"`
}

type CloudflareZone struct {
Name string `json:"name"`
ID string `json:"id"`

}
