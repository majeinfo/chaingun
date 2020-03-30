package main

import (
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Check if the requirements are satisfied
func checkNofileLimit() {
	var rlim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		log.Fatalf("syscall.Getrlimit() failed: %s", err)
	}
	log.Infof("Maximum number of open file descriptors: %d", rlim.Cur)
	if rlim.Cur < 4096 {
		log.Warning("You should increase this value to a higher value")
	}
}
