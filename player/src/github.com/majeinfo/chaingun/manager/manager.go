package manager

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/majeinfo/chaingun/viewer"
	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
	_ "statik"
)

var (
	repositoryDir string

	// Script names played by the Players
	scriptNames = make(map[string]string)
	scriptName  string
)

// Start creates the HTTP server and creates the Web Interface
func Start(mgrAddr *string, reposdir *string) error {
	repositoryDir = *reposdir
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/clean_results", cleanResults)
	mux.HandleFunc("/store_results/", storeResults)
	mux.HandleFunc("/merge_results/", mergeResults)
	//mux.HandleFunc("/results/", showResults)
	mux.Handle("/results/", http.FileServer(http.Dir(".")))
	mux.Handle("/", http.FileServer(statikFS))
	//mux.Handle("/index.html", http.FileServer(statikFS))
	//mux.Handle("/static", http.FileServer(statikFS))

	log.Fatal(http.ListenAndServe(*mgrAddr, mux))
	return nil
}

func cleanResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("cleanResults called")
	scriptNames = make(map[string]string)
	scriptName = ""
	sendJSONResponse(w, "OK", "", "")
}

func storeResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("storeResults called, urlpath=%s", r.URL.Path)

	// The Request Path looks like :
	// /store_results/<repository>/<resultname>/<scriptname>
	parts := strings.Split(r.URL.Path[1:], "/")
	log.Debugf("%v %d", parts, len(parts))
	if len(parts) != 4 {
		sendJSONErrorResponse(w, "Error", "Malformed URL Path")
		return
	}

	// Creates the repository directory if needed
	targetDir := repositoryDir + "/results/" + parts[1] // <repository>
	stat, err := os.Stat(targetDir)
	log.Debugf("Repository Directory is '%s'", targetDir)

	if os.IsNotExist(err) {
		log.Debugf("Must create the Repository Directory")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			sendJSONErrorResponse(w, "Error", err.Error())
			return
		}
	} else if stat.Mode().IsRegular() {
		sendJSONErrorResponse(w, "Error", "Datadir already exists as a file!")
		return
	}

	// Write the results
	fname := targetDir + "/" + parts[2] // <resultname>
	log.Debugf("Write results in file %s", fname)
	file, err := os.Create(fname)
	defer file.Close()
	if err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}
	responseBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	if _, err := file.Write(responseBody); err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	scriptNames[parts[2]] = parts[3] // scriptname
	scriptName = parts[3]

	sendJSONResponse(w, "OK", "", "")
}

func mergeResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("mergeResults called, urlpath=%s", r.URL.Path)

	// The Request Path looks like :
	// /merge_results/<repository>
	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		sendJSONErrorResponse(w, "Error", "Malformed URL Path")
		return
	}

	// Creates the merged file
	repo_dir := repositoryDir + "/results/" + parts[1] // <repository>
	merged_name := "merged.csv"
	fname := repo_dir + "/" + merged_name

	log.Debugf("Creates merged file '%s'", fname)
	mergedFile, err := os.Create(fname)
	defer mergedFile.Close()
	if err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	// Loop on repo_dir content
	files, err := ioutil.ReadDir(repo_dir)
	if err != nil {
		sendJSONErrorResponse(w, "Error", err.Error())
		return
	}

	first_file := true
	for _, file := range files {
		log.Debugf("Found file %s", file.Name())
		if file.Name() == "merged.csv" || file.Name() == "." || file.Name() == ".." {
			continue
		}
		first_line := true

		log.Debugf("Open file '%s'", repo_dir+"/"+file.Name())
		file, err := os.Open(repo_dir + "/" + file.Name())
		if err != nil {
			sendJSONErrorResponse(w, "Error", err.Error())
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if first_line && !first_file {
				first_line = false
				continue
			}

			if _, err := mergedFile.WriteString(scanner.Text() + "\n"); err != nil {
				sendJSONErrorResponse(w, "Error", err.Error())
				return
			}
		}
		first_file = false

		if err := scanner.Err(); err != nil {
			sendJSONErrorResponse(w, "Error", err.Error())
			return
		}
	}

	// Build graphs...
	err = viewer.BuildGraphs(fname, parts[1], repo_dir)
	if err != nil {
		log.Errorf("BuildGraphs failed: %s", err)
		sendJSONErrorResponse(w, "Error", err.Error())
	} else {
		sendJSONResponse(w, "OK", "", repo_dir+"/index.html") // TODO: link_url missing
	}
}

func showResults(w http.ResponseWriter, r *http.Request) {
	log.Debugf("showResults called: %s", r.URL.Path)

	// The Request Path looks like :
	// /results/<repository>/<file>....

}

func sendJSONErrorResponse(w http.ResponseWriter, status string, msg string) {
	log.Debugf("sendJSONErrorResponse called")
	sendJSONResponse(w, status, msg, "")
}

func sendJSONResponse(w http.ResponseWriter, status string, msg string, link_url string) {
	log.Debugf("sendJSONResponse called")

	data := map[string]string{
		"status":   status,
		"msg":      msg,
		"link_url": link_url,
	}
	jData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Json Marshaling failed with %v (%s)", data, err)
	}
	log.Debugf(string(jData))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}