package reporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

var (
	w   *bufio.Writer
	f   *os.File
	err error
)

var opened bool = false

func OpenResultsFile(fileName string) {
	if !opened {
		opened = true
	} else {
		return
	}
	f, err = os.Create(fileName)
	if err != nil {
		locdir := path.Dir(fileName)
		err := os.MkdirAll(locdir, 0755)
		if err != nil {
			log.Fatal("Cannot create directory %s: %s", locdir, err)
		}
		f, err = os.Create(fileName)
		if err != nil {
			log.Fatal("Cannot create output file %s: %s", fileName, err)
		}
	}
	initResultsFile()
}

func OpenTempResultsFile(tmpfile *os.File) {
	f = tmpfile
	initResultsFile()
}

func initResultsFile() {
	w = bufio.NewWriter(f)
	if output_type == "json" {
		_, err = w.WriteString(string("var logdata = '"))
	} else if output_type == "csv" {
		_, err = w.WriteString("Timestamp,Vid,Type,Title,Status,Size,Latency\n")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func CloseResultsFile() {
	if opened {
		if output_type == "json" {
			_, err = w.WriteString(string("';"))
		}
		w.Flush()
		f.Close()
	}
	// Do nothing if not opened
	opened = false
}

func WriteResult(httpResult *HttpReqResult) {
	if output_type == "json" {
		jsonString, err := json.Marshal(httpResult)

		if err != nil {
			log.Fatal(err)
		}
		_, err = w.WriteString(string(jsonString))
		_, err = w.WriteString("|")
	} else if output_type == "csv" {
		s := fmt.Sprintf("%d,%s,%s,%s,%d,%d,%d\n",
			httpResult.When, httpResult.Vid, httpResult.Type, httpResult.Title,
			httpResult.Status, httpResult.Size, httpResult.Latency)
		_, err = w.WriteString(s)
	}

	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

}
