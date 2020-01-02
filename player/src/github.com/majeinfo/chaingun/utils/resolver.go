package utils

import (
	"math/rand"
	"net"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	addrCache = make(map[string][]string, 50)
	cacheLock sync.Mutex
)

// GetServerAddress fills and manages a cache on host/IP address
// to minimize DNS requests
func GetServerAddress(serverName string) (string, bool) {
	// Split the serverName into two parts: the Host and the Port (may be empty)
	parts := strings.Split(serverName, ":")
	host := parts[0]

	// Check if address is already in cache, otherwise try to fill the cache
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if values, err := addrCache[host]; err == true {
		return appendPort(getAddr(values), parts), true
	}

	if addrs, err := net.LookupHost(host); err == nil {
		// Remove IPv6 addresses
		addrs4 := make([]string, 0)
		for _, addr := range addrs {
			if strings.Count(addr, ":") == 0 {
				addrs4 = append(addrs4, addr)
			}
		}
		addrCache[host] = addrs4
		return appendPort(getAddr(addrs4), parts), true
	}

	log.Errorf("Could not resolve the Server Name: %s", host)
	return "", false
}

func getAddr(addrs []string) string {
	log.Debugf("getAddr: %v", addrs)
	if len(addrs) == 1 {
		return addrs[0]
	}

	idx := rand.Intn(len(addrs))
	log.Debugf("getAddr returns: %s", addrs[idx])
	return addrs[idx]
}

func appendPort(host string, parts []string) string {
	// Concatenate the port with the host name
	if len(parts) > 1 {
		return host + ":" + parts[1]
	}
	return host
}
