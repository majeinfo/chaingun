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

func GetServerAddress(serverName string) (string, bool) {
	// Check if address is already in cache, otherwise try to fill the cache
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if values, err := addrCache[serverName]; err == true {
		return getAddr(values), true
	}

	if addrs, err := net.LookupHost(serverName); err == nil {
		// Remove IPv6 addresses
		addrs4 := make([]string, 0)
		for _, addr := range addrs {
			if strings.Count(addr, ":") == 0 {
				addrs4 = append(addrs4, addr)
			}
		}
		addrCache[serverName] = addrs4
		return getAddr(addrs4), true
	}

	log.Errorf("Could not resolve the Server Name: %s", serverName)
	return "", false
}

func getAddr(addrs []string) string {
	log.Debugf("getAddr: %v", addrs)
	if len(addrs) == 1 {
		return addrs[0]
	}

	idx := rand.Intn(len(addrs))
	log.Debugf("getAddr returns:", addrs[idx])
	return addrs[idx]
}
