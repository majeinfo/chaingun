package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Handle the Ctrl-C and forces the VU to finish ASAP but still conserve the results
func shutdownHandler() {
	// signChan channel is used to transmit signal notifications.
	signChan := make(chan os.Signal, 1)
	// Catch and relay certain signal(s) to signChan channel.
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	// Blocking until a signal is sent over signChan channel. Progress to
	// next line after signal
	sig := <-signChan
	log.Infoln("Cleanup started with", sig, "signal")

	lock_emergency_stop.Lock()
	gp_emergency_stop = true
	lock_emergency_stop.Unlock()
	/*
	os.Exit(1)
	*/
}
