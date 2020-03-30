package main

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/feeder"
)

func cleanSessionMapAndResetUID(UID string, sessionMap map[string]string, playbook *config.TestDef) {
	// Optimization? Delete all entries rather than reallocate map from scratch for each new iteration.
	for k := range sessionMap {
		delete(sessionMap, k)
	}

	// Set permanent variable and variables from playbook
	sessionMap["UID"] = UID
	sessionMap[config.HTTP_RESPONSE] = ""
	sessionMap[config.MONGODB_LAST_INSERT_ID] = ""
	sessionMap[config.SQL_ROW_COUNT] = "0"

	for k, v := range playbook.Variables {
		sessionMap[k] = v
	}
}

func feedSession(playbook *config.TestDef, sessionMap map[string]string) {
	if playbook.DataFeeder.Type != "" {
		go feeder.NextFromFeeder()       // Do async
		feedData := <-feeder.FeedChannel // Will block here until feeder delivers value over the FeedChannel
		for item := range feedData {
			sessionMap[item] = feedData[item]
		}
	}
}
