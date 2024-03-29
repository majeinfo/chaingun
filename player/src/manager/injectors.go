package manager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	"github.com/majeinfo/chaingun/viewer"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	ERR = "ERR"
)

var (
	injectorClients = make(map[string]*websocket.Conn)
	targetDir       = "/tmp/results"
	pre_tasks_lock  = sync.Mutex{}
	pre_tasks_cond  = sync.NewCond(&pre_tasks_lock)
	post_tasks_lock = sync.Mutex{}
	post_tasks_cond = sync.NewCond(&post_tasks_lock)
)

// Start the Batch mode
func StartBatch(reposdir string, prelaunched_injectors string, script_file string) error {
	// Build the action from playbook
	var playbook config.TestDef
	var pre_actions []action.FullAction
	var actions []action.FullAction
	var post_actions []action.FullAction

	data, err := ioutil.ReadFile(script_file)
	if err != nil {
		log.Fatal(err)
	}
	if !action.CreatePlaybook(script_file, []byte(data), &playbook, &pre_actions, &actions, &post_actions) {
		log.Fatalf("Error while processing the Script File")
	}

	// Get the list of embedded files
	log.Infof("Embedded files: %v", action.GetEmbeddedFilenames())

	// Build the Injector list
	if len(prelaunched_injectors) > 0 {
		injectors = strings.Split(prelaunched_injectors, ",")
	} else {
		injectors = make([]string, 0)
	}
	targetDir = reposdir
	nu_injectors := 0

	// Creates the repository directory if needed
	_, err = os.Stat(targetDir)
	log.Debugf("Repository Directory is '%s'", targetDir)

	if os.IsNotExist(err) {
		log.Debugf("Must create the Repository Directory")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			log.Errorf("Could not create the directory to store the results")
			return err
		}
	}

	// Connect to each Injector - stop in case of error or if no Injector found
	for _, injector := range injectors {
		/*
			conn, err := net.Dial("tcp", injector)
			if err != nil {
				log.Fatalf("Cannot connect to Injector %s", injector)
			}
		*/

		u := url.URL{Scheme: "ws", Host: injector, Path: "/upgrade"}
		log.Printf("connecting to %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatalf("Error when dialing Injector %s: %s", injector, err)
		}
		injectorClients[injector] = c
		err = c.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"status\" }"))
		if err != nil {
			log.Fatalf("Error when writing to Injector %s: %s", injector, err)
		}
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatalf("Could not get answer from Injector %s: %s", injector, err)
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		level, msg, detail := decodeInjectorStatus(injector, message)
		log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)

		nu_injectors++
	}

	if nu_injectors < 1 {
		log.Fatal("Stopping... no valid injector found !")
	}

	// Run the script !
	runScript(script_file)

	return nil
}

func runScript(script_file string) {
	// Read all the data files and compute their MD5sum
	var encoded_files = make(map[string]string, len(action.GetEmbeddedFilenames()))
	var md5sums = make(map[string]string, len(action.GetEmbeddedFilenames()))

	for _, fname := range action.GetEmbeddedFilenames() {
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Fatalf("Cannot read the data file %s: %s", fname, err)
		}
		encoded_files[fname] = base64.StdEncoding.EncodeToString(data)
		md5sums[fname], err = utils.Hash_file_md5(fname)
		if err != nil {
			log.Fatalf("Cannot compute MD5sum of file %s: %s", fname, err)
		}
		log.Debugf("MD5sum for file %s is %s", fname, md5sums[fname])
	}

	// Read the scenario file and convert it in Base64
	data, err := ioutil.ReadFile(script_file)
	if err != nil {
		log.Fatalf("Cannot read the script file %s: %s", script_file, err)
	}

	encoded_data := base64.StdEncoding.EncodeToString(data)

	wg := sync.WaitGroup{}
	pre_tasks_lock.Lock()
	post_tasks_lock.Lock()
	first_injector := true
	for injector, conn := range injectorClients {
		wg.Add(1)
		go runScriptOnInjector(first_injector, injector, conn, &script_file, encoded_data, encoded_files, md5sums, &wg)
		first_injector = false
	}
	log.Info("Waiting for the Injectors to complete their job...")
	wg.Wait()
	log.Info("Jobs completed")
	log.Info("Merge the results")
	err = _mergeResults(targetDir)
	if err != nil {
		log.Fatalf("Error while merging result files: %s", err)
	}

	// Create metadata files
	scriptnames := []string{script_file}
	if err := reporter.WriteMetadata(time.Now(), time.Now(), targetDir, scriptnames); err != nil {
		log.Fatalf("Error while writing metedata file: %s", err.Error())
	}

	// Build graphs
	if err := viewer.BuildGraphs(targetDir+"/merged.csv", script_file, targetDir); err != nil {
		log.Fatalf("BuildGraphs failed: %s", err)
	}
}

func runScriptOnInjector(first_injector bool, injector string, conn *websocket.Conn, script_file *string, encoded_data string,
	encoded_files map[string]string, md5sums map[string]string, wg *sync.WaitGroup) error {
	defer wg.Done()

	// TODO: file names are computed from script file location or they can be given using S3 URL ?

	// Send the data and feeder files
	for _, fname := range action.GetEmbeddedFilenames() {
		sendDataFile(injector, conn, fname, encoded_files[fname], md5sums[fname])
	}

	// Send the script
	err := sendScript(injector, conn, script_file, encoded_data)
	if err != nil {
		return err
	}

	// sync mechanism so that the other go routines (injectors) do
	// not start playing the script until the pre-tasks are made

	// Start the pre-actions only on first Injector
	if first_injector {
		err = preStartScript(injector, conn)
		pre_tasks_cond.Broadcast()
		if err != nil {
			return err
		}
	} else {
		pre_tasks_cond.Wait()
	}

	// Start the script
	err = startScript(injector, conn)
	if err != nil {
		return err
	}

	// Start the post-actions only on first Injector
	if first_injector {
		err = postStartScript(injector, conn)
		post_tasks_cond.Broadcast()
		if err != nil {
			return err
		}
	} else {
		post_tasks_cond.Wait()
	}

	// Get the results
	err = getResults(injector, conn)
	if err != nil {
		return err
	}

	return nil
}

func sendScript(injector string, conn *websocket.Conn, script_file *string, encoded_data string) error {
	log.Infof("Send script %s to Injector %s", *script_file, injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"script\", \"moreinfo\": \""+*script_file+"\", \"value\": \""+encoded_data+"\" }"))
	if err != nil {
		log.Errorf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	level, msg, detail := decodeInjectorStatus(injector, message)
	log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
	if level == ERR {
		err = fmt.Errorf("Injector returned error: %v", msg)
		log.Errorf("%v", err)
		return err
	}

	return nil
}

func sendDataFile(injector string, conn *websocket.Conn, fname string, encoded_data string, md5sum string) error {
	log.Infof("Send data file %s (%d) to Injector %s", fname, len(encoded_data), injector)

	// Get the remote MD5 value
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"getmd5\", \"moreinfo\": \"\", \"value\": \""+fname+"\" }"))
	if err != nil {
		log.Errorf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	level, msg, detail := decodeInjectorStatus(injector, message)
	log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
	if level == ERR {
		err = fmt.Errorf("Injector returned error: %v", message)
		log.Errorf("%v", err)
		return err
	}

	// If MD5 sum match then do not send the file content !
	var stat PlayerStatus
	err = json.Unmarshal(message, &stat)
	if err != nil {
		log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
		return err
	}
	log.Debugf("Injector %s returned MD5sum value %s for file %s", injector, stat.Msg, fname)
	if stat.Msg == md5sum {
		log.Debugf("Returned MD5sum matches => file %s not transfered", fname)
		return nil
	} else {
		log.Debugf("Returned MD5sum does not match => file %s (re)transferred", fname)
	}

	// Data to be sent may be huge, so we must send them by chunks...
	const CHUNKSIZE = 30_000

	if len(encoded_data) < CHUNKSIZE {
		// No need to split the data in chunks
		err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"datafile\", \"moreinfo\": \""+fname+"\", \"value\": \""+encoded_data+"\" }"))
		if err != nil {
			log.Errorf("Error when writing to Injector %s: %s", injector, err)
			return err
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not get answer from Injector %s: %s", injector, err)
			return err
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		level, msg, detail := decodeInjectorStatus(injector, message)
		log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
		if level == ERR {
			err = fmt.Errorf("Injector returned error: %v", message)
			log.Errorf("%v", err)
			return err
		}
	} else {
		// Send first chunk
		err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"datafile\", \"moreinfo\": \""+fname+"\", \"value\": \""+encoded_data[:CHUNKSIZE]+"\" }"))
		if err != nil {
			log.Errorf("Error when writing to Injector %s: %s", injector, err)
			return err
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not get answer from Injector %s: %s", injector, err)
			return err
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		level, msg, detail := decodeInjectorStatus(injector, message)
		log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
		if level == ERR {
			err = fmt.Errorf("Injector returned error: %v", message)
			log.Errorf("%v", err)
			return err
		}

		encoded_data = encoded_data[CHUNKSIZE:]
		for len(encoded_data) > CHUNKSIZE {
			// send next chunk
			log.Infof("Send data chunk %s to Injector %s", fname, injector)
			err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"nextchunk\", \"moreinfo\": \""+fname+"\", \"value\": \""+encoded_data[:CHUNKSIZE]+"\" }"))
			if err != nil {
				log.Errorf("Error when writing to Injector %s: %s", injector, err)
				return err
			}
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Errorf("Could not get answer from Injector %s: %s", injector, err)
				return err
			}
			//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			level, msg, detail := decodeInjectorStatus(injector, message)
			log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
			if level == ERR {
				err = fmt.Errorf("Injector returned error: %v", message)
				log.Errorf("%v", err)
				return err
			}

			encoded_data = encoded_data[CHUNKSIZE:]
		}

		// last chunk
		log.Infof("Send last data chunk %s to Injector %s", fname, injector)
		err = conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"nextchunk\", \"moreinfo\": \""+fname+"\", \"value\": \""+encoded_data+"\" }"))
		if err != nil {
			log.Errorf("Error when writing to Injector %s: %s", injector, err)
			return err
		}
		_, message, err = conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not get answer from Injector %s: %s", injector, err)
			return err
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		level, msg, detail = decodeInjectorStatus(injector, message)
		log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
		if level == ERR {
			err = fmt.Errorf("Injector returned error: %v", message)
			log.Errorf("%v", err)
			return err
		}
	}

	return nil
}

func preStartScript(injector string, conn *websocket.Conn) error {
	log.Infof("Start pre-actions on Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"pre_start\" }"))
	if err != nil {
		log.Errorf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	level, msg, detail := decodeInjectorStatus(injector, message)
	log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
	if level == ERR {
		err = fmt.Errorf("Injector returned error: %v", message)
		log.Errorf("%v", err)
		return err
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not get answer from Injector %s: %s", injector, err)
			return err
		}
		log.Debugf("Injector %s sent: %s", injector, message)
		var stat reporter.StatFrame
		err = json.Unmarshal(message, &stat)
		if err != nil {
			log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
			return err
		}

		if stat.Type == "status" {
			var status PlayerStatus
			err = json.Unmarshal(message, &status)
			if err != nil {
				log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
				return err
			}
			log.Debug("status rcvd")
			break
		} else {
			log.Debug("rt frame rcvd")
		}
	}

	return nil
}

func postStartScript(injector string, conn *websocket.Conn) error {
	log.Infof("Start post-actions on Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"post_start\" }"))
	if err != nil {
		log.Errorf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	level, msg, detail := decodeInjectorStatus(injector, message)
	log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
	if level == ERR {
		err = fmt.Errorf("Injector returned error: %v", message)
		log.Errorf("%v", err)
		return err
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not get answer from Injector %s: %s", injector, err)
			return err
		}
		log.Debugf("Injector %s sent: %s", injector, message)
		var stat reporter.StatFrame
		err = json.Unmarshal(message, &stat)
		if err != nil {
			log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
			return err
		}

		if stat.Type == "status" {
			var status PlayerStatus
			err = json.Unmarshal(message, &status)
			if err != nil {
				log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
				return err
			}
			log.Debug("status rcvd")
			break
		} else {
			log.Debug("rt frame rcvd")
		}
	}

	return nil
}

func startScript(injector string, conn *websocket.Conn) error {
	log.Infof("Start script on Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"start\" }"))
	if err != nil {
		log.Errorf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	level, msg, detail := decodeInjectorStatus(injector, message)
	log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
	if level == ERR {
		err = fmt.Errorf("Injector returned error: %v", message)
		log.Errorf("%v", err)
		return err
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Could not get answer from Injector %s: %s", injector, err)
			return err
		}
		level, msg, detail := decodeInjectorStatus(injector, message)
		log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
		if level == ERR {
			err = fmt.Errorf("Injector returned error: %v", message)
			log.Errorf("%v", err)
			return err
		}

		var stat reporter.StatFrame
		err = json.Unmarshal(message, &stat)
		if err != nil {
			log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
			return err
		}

		if stat.Type == "status" {
			var status PlayerStatus
			err = json.Unmarshal(message, &status)
			if err != nil {
				log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
				return err
			}
			log.Debug("status rcvd")
			break
		} else {
			log.Debug("rt frame rcvd")
		}
	}

	return nil
}

func getResults(injector string, conn *websocket.Conn) error {
	// Get the results from the Injector
	log.Infof("Get results from Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"get_results\" }"))
	if err != nil {
		log.Errorf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	level, msg, detail := decodeInjectorStatus(injector, message)
	log.Infof("Injector %s answers: level=%s, message=%s, detail=%s", injector, level, msg, detail)
	if level == ERR {
		err = fmt.Errorf("Injector returned error: %v", message)
		log.Errorf("%v", err)
		return err
	}

	_, message, err = conn.ReadMessage()
	if err != nil {
		log.Errorf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	log.Debugf("Injector %s sent: %s", injector, message)
	var results PlayerResults
	err = json.Unmarshal(message, &results)
	if err != nil {
		log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
		return err
	}
	log.Debugf("Injector %s answers: %s", injector, results)

	return _storeResults(injector, results.Msg)
}

func _storeResults(injector string, results string) error {
	// Write the results
	fname := targetDir + "/" + injector + fname_suffix
	log.Debugf("Write results in file %s", fname)
	file, err := os.Create(fname)
	defer file.Close()
	if err != nil {
		log.Errorf("Could not create file %s: %s", fname, err)
		return err
	}

	if _, err := file.Write([]byte(results)); err != nil {
		log.Errorf("Could not write into file %s: %s", fname, err)
		return err
	}

	return nil
}

func decodeInjectorStatus(injector string, msg []byte) (string, string, string) {
	// Decode JSON message
	var status PlayerStatus
	err := json.Unmarshal(msg, &status)
	if err != nil {
		log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
		return "", "", ""
	}

	return status.Level, status.Msg, status.Detail
}
