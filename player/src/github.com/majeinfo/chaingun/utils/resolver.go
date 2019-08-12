package utils

import (
	"math/rand"
	"net"
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
		addrCache[serverName] = addrs
		return getAddr(addrs), true
	}

	log.Errorf("Could not resolve the Server Name: %s", serverName)
	return "", false
}

func getAddr(addrs []string) string {
	if len(addrs) == 1 {
		return addrs[0]
	}

	idx := rand.Intn(len(addrs))
	return addrs[idx]
}
