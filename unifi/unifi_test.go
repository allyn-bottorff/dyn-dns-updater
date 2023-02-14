package unifi

import (
	"testing"
)

func TestJson(t *testing.T) {
	username := "blah"
	password := "blahpass"

	credsJson := makeCredsJson(username, password)
	correctJson := "{\"username\": \"blah\", \"password\": \"blahpass\"}"

	if credsJson != correctJson {
		t.Fatalf("%s != %s", credsJson, correctJson)
	}

	


}
