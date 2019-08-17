package manager

import (
	"encoding/base64"
	"encoding/json"
	"github.com/majeinfo/chaingun/reporter"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var (
	injectorClients = make(map[string]*websocket.Conn)
)

// Start the Batch mode
func StartBatch(mgrAddr *string, reposdir *string, prelaunched_injectors *string, script_file *string) error {
	if len(*prelaunched_injectors) > 0 {
		injectors = strings.Split(*prelaunched_injectors, ",")
	} else {
		injectors = make([]string, 0)
	}
	repositoryDir = *reposdir
	nu_injectors := 0

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
}

func runScriptOnInjector(injector string, conn *websocket.Conn, script_file *string, encoded_data string, wg *sync.WaitGroup) error {
	defer wg.Done()

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

	log.Infof("Start script on Injector %s", injector)
	err = conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"start\" }"))
	if err != nil {
		log.Fatalf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err = conn.ReadMessage()
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

	// Get the results from the Injector
	log.Infof("Get results from Injector %s", injector)
	err = conn.WriteMessage(websocket.TextMessage, []byte("{ \"cmd\": \"get_results\" }"))
	if err != nil {
		log.Fatalf("Error when writing to Injector %s: %s", injector, err)
		return err
	}
	_, message, err = conn.ReadMessage()
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

	// TODO: Store the results
	// TODO: Merge the results

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
