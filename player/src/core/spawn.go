package core

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// Launch VUs
func spawnUsers(playbook *config.TestDef, actions *[]action.FullAction, mode int) {
	log.Info("Launch VUs to play the actions")
	resultsChannel := make(chan reporter.SampleReqResult, 10000)
	go reporter.AcceptResults(resultsChannel, &VU_count, &lock_vu_count, &hub.broadcast, mode == DaemonMode)
	VU_start = time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < playbook.Users; i++ {
		wg.Add(1)
		lock_vu_count.Lock()
		VU_count++
		lock_vu_count.Unlock()
		//UID := strconv.Itoa(rand.Intn(playbook.Users+1) + 10000)
		UID := strconv.Itoa(os.Getpid()*100000 + i)
		go launchActions(playbook, resultsChannel, &wg, actions, UID)
		waitDuration := float32(playbook.Rampup) / float32(playbook.Users)
		time.Sleep(time.Duration(int(1000*waitDuration)) * time.Millisecond)

		// In daemon mode, we may receive an order to stop the load test
		if gp_daemon_status == STOPPING_NOW {
			log.Info("Stop now")
			break
		}

		// In standalone mode, we may receive a Ctrl-C
		//lock_emergency_stop.Lock()
		must_stop := g_emergency_stop
		//lock_emergency_stop.Unlock()
		if must_stop {
			log.Info("Stopping the VU ASAP !")
			break
		}
	}
	log.Info("All VUs started, waiting at WaitGroup")
	wg.Wait()
	reporter.StopResults()
}

// Called once per each VU
func launchActions(playbook *config.TestDef, resultsChannel chan reporter.SampleReqResult, wg *sync.WaitGroup, actions *[]action.FullAction, UID string) {
	log.Debugf("launchActions called (%s)", UID)
	var sessionMap = make(map[string]string)
	var vucontext config.VUContext

	i := 0
	vulog := log.WithFields(log.Fields{"vuid": UID, "iter": i, "action": ""})

actionLoop:
	for (playbook.Iterations == -1) || (i < playbook.Iterations) {
		// In standalone mode, we may receive a Ctrl-C
		//lock_emergency_stop.Lock()
		must_stop := g_emergency_stop
		//lock_emergency_stop.Unlock()
		if must_stop {
			log.Debugf("Stopping VU %s", UID)
			break
		}

		vulog.Data["iter"] = i

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
				if !action.Action.Execute(resultsChannel, sessionMap, &vucontext, vulog, playbook) {
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
						break actionLoop
					case config.ERR_STOP_VU:
						vulog.Info("Stop VU on error")
						break actionLoop
					}
				}
			}
		}

		i++
		if playbook.Iterations == -1 {
			ti := time.Now()
			if ti.Sub(VU_start) > time.Duration(playbook.Duration)*time.Second {
				//log.Info("finished", time.Duration(t.Duration) * time.Second, ti.Sub(VU_start))
				break
			}
		}
	}
	wg.Done()
	lock_vu_count.Lock()
	VU_count--
	lock_vu_count.Unlock()

	if vucontext.CloseFunc != nil {
		log.Debugf("Call CloseFunc")
		vucontext.CloseFunc(&vucontext)
	}

	log.Debugf("exit launchActions (%s)", UID)
}
