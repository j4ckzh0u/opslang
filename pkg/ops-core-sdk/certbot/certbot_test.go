package certbot

import (
	"testing"
)

func TestCertificates(t *testing.T) {
	// May fail if certbot not installed
	_, _ = Certificates()
}

func TestObtainNoDomains(t *testing.T) {
	_, err := Obtain([]string{}, "", "", true)
	if err == nil {
		t.Fatal("Obtain with no domains should return error")
	}
}

func TestObtainNoAuthMethod(t *testing.T) {
	_, err := Obtain([]string{"example.com"}, "", "", false)
	if err == nil {
		t.Fatal("Obtain without standalone or webroot should return error")
	}
}

func TestDeleteEmptyDomain(t *testing.T) {
	_, err := Delete("")
	if err == nil {
		t.Fatal("Delete with empty domain should return error")
	}
}
