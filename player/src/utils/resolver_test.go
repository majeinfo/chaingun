package utils

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestGoodNames(t *testing.T) {
	domains := []string{
		"www.google.com",
		"www.google.com:80",
	}

	for _, domain := range domains {
		if _, found := GetServerAddress(domain); !found {
			t.Errorf("Could not get IP address of %s", domain)
		}
	}
}

func TestCheckCache(t *testing.T) {
	var ipaddr string
	var found bool

	/* Insert a value in the cache and check it can be get back */
	host := "www.example.com"
	addrs4 := make([]string, 0)
	addrs4 = append(addrs4, "1.2.3.4")
	addrs4 = append(addrs4, "5.6.7.8")
	addrCacheSynced.Store(host, addrs4)

	if ipaddr, found = GetServerAddress(host); !found {
		t.Errorf("Could not get IP address of %s", host)
		return
	}
	if ipaddr != "1.2.3.4" && ipaddr != "5.6.7.8" {
		t.Errorf("Bad IP address: should be 1.2.3.4, received '%s'", ipaddr)
		return
	}
	ipaddr, _ = GetServerAddress(host)
	if ipaddr != "1.2.3.4" && ipaddr != "5.6.7.8" {
		t.Errorf("Bad IP address: should be 5.6.7.8, received '%s'", ipaddr)
		return
	}
	ipaddr, _ = GetServerAddress(host)
	if ipaddr != "1.2.3.4" && ipaddr != "5.6.7.8" {
		t.Errorf("Bad IP address: should be 1.2.3.4, received '%s'", ipaddr)
	}
}

func TestBadName(t *testing.T){
	var buf bytes.Buffer
	log.SetOutput(&buf)

	if _, found := GetServerAddress("xyz"); found {
		t.Error("Should not find a value for un unknown domain !")
	}

	log.SetOutput(os.Stderr)
}