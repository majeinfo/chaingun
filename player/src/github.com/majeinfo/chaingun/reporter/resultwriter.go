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
			log.Fatalf("Cannot create directory %s: %s", locdir, err)
		}
		f, err = os.Create(fileName)
		if err != nil {
			log.Fatalf("Cannot create output file %s: %s", fileName, err)
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
		_, err = w.WriteString("Timestamp,Vid,Type,Title,Status,Size,Latency,FullRequest\n")
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

func WriteResult(sampleResult *SampleReqResult) {
	if output_type == "json" {
		jsonString, err := json.Marshal(sampleResult)

		if err != nil {
			log.Fatal(err)
		}
		_, err = w.WriteString(string(jsonString))
		_, err = w.WriteString("|")
	} else if output_type == "csv" {
		s := fmt.Sprintf("%d,%s,%s,%s,%d,%d,%d,%s\n",
			sampleResult.When, sampleResult.Vid, sampleResult.Type, sampleResult.Title,
			sampleResult.Status, sampleResult.Size, sampleResult.Latency, sampleResult.FullRequest)
		_, err = w.WriteString(s)
	}

	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

}
