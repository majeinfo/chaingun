package main

import (
	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// Launch Pre-Actions
func playPreActions(playbook *config.TestDef, actions *[]action.FullAction) {
	log.Info("Play pre-actions")
	var sessionMap = make(map[string]string)
	resultsChannel := make(chan reporter.SampleReqResult, 100)

	i := 0
	UID := "preActions"
	vulog := log.WithFields(log.Fields{"vuid": UID, "iter": i, "action": ""})

	// Make sure the sessionMap is cleared before each iteration - except for the UID which stays
	cleanSessionMapAndResetUID(UID, sessionMap, playbook)

	// If we have feeder data, pop an item and push its key-value pairs into the sessionMap
	feedSession(playbook, sessionMap)

	// Iterate over the actions. Note the use of the command-pattern like Execute method on the Action interface
iterLoop:
	for _, action := range *actions {
		if action.Action != nil {
			// Check for a "when" expression
			if action.CompiledWhen != nil {
				vulog.Debugf("Evaluate 'when' expression: %s", action.When)

				// if evaluation is False, skip the action
				result, err := utils.Evaluate(sessionMap, vulog, action.CompiledWhen, action.When)
				skip := false
				if err == nil {
					switch result.(type) {
					case float64:
						skip = result.(float64) == 0
					case string:
						skip = result.(string) == ""
					case bool:
						skip = !result.(bool)
					default:
						vulog.Errorf("Error when evaluating expression: unknown type %v", result)
					}
				}
				if skip {
					vulog.Infof("Action skipped due to its 'when' condition")
					continue
				}
			}
			if !action.Action.Execute(resultsChannel, sessionMap, vulog, playbook) {
				// An error occurred : continue, stop the vu or stop the test ?
				switch playbook.OnError {
				case config.ERR_CONTINUE:
					vulog.Info("Continue on error")
					break
				case config.ERR_STOP_ITERATION:
					vulog.Info("Stop this iteration")
					break iterLoop
				case config.ERR_STOP_TEST:
					vulog.Info("Stop test on error")
					gp_daemon_status = STOPPING_NOW
					break iterLoop
				case config.ERR_STOP_VU:
					vulog.Info("Stop VU on error")
					break iterLoop
				}
			}
		}
	}

	log.Debug("exit playPreActions")
}
