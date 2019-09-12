package manager

import (
	"encoding/base64"
	"encoding/json"
	"github.com/majeinfo/chaingun/reporter"
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

var (
	injectorClients = make(map[string]*websocket.Conn)
	targetDir       = "/tmp/results"
)

// Start the Batch mode
func StartBatch(mgrAddr *string, reposdir *string, prelaunched_injectors *string, script_file *string) error {
	if len(*prelaunched_injectors) > 0 {
		injectors = strings.Split(*prelaunched_injectors, ",")
	} else {
		injectors = make([]string, 0)
	}
	targetDir = *reposdir
	nu_injectors := 0

	// Creates the repository directory if needed
	_, err := os.Stat(targetDir)
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
		log.Debugf("Injector %s answers: %s", injector, message)
		log.Infof("Injector %s answers: %s", injector, decodeInjectorStatus(injector, message))

		nu_injectors++
	}

	if nu_injectors < 1 {
		log.Fatal("Stopping... no valid injector found !")
	}

	// Run the script !
	runScript(script_file)

	return nil
}

func runScript(script_file *string) {
	// Read the file and convert it in Base64
	// Read the scenario from file
	data, err := ioutil.ReadFile(*script_file)
	if err != nil {
		log.Fatalf("Cannot read the script file %s: %s", *script_file, err)
	}

	encoded_data := base64.StdEncoding.EncodeToString(data)
	if err != nil {
		log.Fatalf("Cannot encode the script file %s: %s", *script_file, err)
	}

	wg := sync.WaitGroup{}
	for injector, conn := range injectorClients {
		wg.Add(1)
		go runScriptOnInjector(injector, conn, script_file, encoded_data, &wg)
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
	scriptnames := []string{*script_file}
	if err := reporter.WriteMetadata(time.Now(), time.Now(), targetDir, scriptnames); err != nil {
		log.Fatalf("Error while writing metedata file: %s", err.Error())
	}

	// Build graphs
	if err := viewer.BuildGraphs(targetDir+"/merged.csv", *script_file, targetDir); err != nil {
		log.Fatalf("BuildGraphs failed: %s", err)
	}
}

func runScriptOnInjector(injector string, conn *websocket.Conn, script_file *string, encoded_data string, wg *sync.WaitGroup) error {
	defer wg.Done()

	err := sendScript(injector, conn, script_file, encoded_data)
	if err != nil {
		return err
	}

	// TODO: should send the feeder data !
	// TODO: should send the template data file !
	// TODO: file names are computed from script file locate or they can be given using S3 URL ?

	err = startScript(injector, conn)
	if err != nil {
		return err
	}

	err = getResults(injector, conn)
	if err != nil {
		return err
	}

	return nil
}

func sendScript(injector string, conn *websocket.Conn, script_file *string, encoded_data string) error {
	log.Infof("Send script to Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"script\", \"moreinfo\": \""+*script_file+"\", \"value\": \""+encoded_data+"\" }"))
	if err != nil {
		log.Fatalf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	log.Debugf("Injector %s answers: %s", injector, message)
	log.Infof("Injector %s answers: %s", injector, decodeInjectorStatus(injector, message))

	return nil
}

func startScript(injector string, conn *websocket.Conn) error {
	log.Infof("Start script on Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"start\" }"))
	if err != nil {
		log.Fatalf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	log.Debugf("Injector %s answers: %s", injector, message)
	log.Infof("Injector %s answers: %s", injector, decodeInjectorStatus(injector, message))

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("Could not get answer from Injector %s: %s", injector, err)
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

func getResults(injector string, conn *websocket.Conn) error {
	// Get the results from the Injector
	log.Infof("Get results from Injector %s", injector)
	err := conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"get_results\" }"))
	if err != nil {
		log.Fatalf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Could not get answer from Injector %s: %s", injector, err)
		return err
	}
	log.Debugf("Injector %s answers: %s", injector, message)
	log.Infof("Injector %s answers: %s", injector, decodeInjectorStatus(injector, message))

	_, message, err = conn.ReadMessage()
	if err != nil {
		log.Fatalf("Could not get answer from Injector %s: %s", injector, err)
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

func decodeInjectorStatus(injector string, msg []byte) string {
	// Decode JSON message
	var status PlayerStatus
	err := json.Unmarshal(msg, &status)
	if err != nil {
		log.Errorf("Message from Injector %s could not be decoded as JSON", injector)
		return ""
	}

	return status.Level
}
