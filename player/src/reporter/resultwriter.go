package reporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	_ "strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// flushInterval bounds how much data can be lost if the process is killed
// abruptly: WriteResult() is only flushed to disk at most this often instead
// of on every single result, since fsync-ing per result caps throughput.
const flushInterval = 500 * time.Millisecond

var (
	w   *bufio.Writer
	f   *os.File
	err error
	opened bool = false
	clean_re *regexp.Regexp
	lastFlush time.Time
)

func init() {
	clean_re, err = regexp.Compile(`[",]`)
	if err != nil {
		log.Fatalf("Could not compile regexp ! %v", err)
	}
}

func OpenResultsFile(fileName string) {
	log.Debugf("OpenResultFile: %s", fileName)

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
	// Remove the previous temp file if it exists
	if f != nil {
		os.Remove(f.Name())
	}

	f = tmpfile
	initResultsFile()
}

func initResultsFile() {
	log.Debug("initResultFile()")
	w = bufio.NewWriter(f)
	lastFlush = time.Now()
	if outputType == jsonOutput {
		_, err = w.WriteString(string("var logdata = '"))
	} else if outputType == csvOutput {
		_, err = w.WriteString("Timestamp,Vid,Type,Title,Status,Size,Latency,FullRequest\n")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func CloseResultsFile() {
	if opened {
		if outputType == jsonOutput {
			_, err = w.WriteString(string("';"))
		}
		w.Flush()
		f.Close()
	}
	// Do nothing if not opened
	opened = false
}

func WriteResult(sampleResult *SampleReqResult) {
	if outputType == jsonOutput {
		jsonString, err := json.Marshal(sampleResult)

		if err != nil {
			log.Fatal(err)
		}
		_, err = w.WriteString(string(jsonString))
		_, err = w.WriteString("|")
	} else if outputType == csvOutput {
		// Because of CSV structure, we remove double-quote and commas in the FullRequest
		cleanFullRequest := clean_re.ReplaceAllString(sampleResult.FullRequest, "")
		s := fmt.Sprintf("%d,%s,%s,%s,%d,%d,%d,%s\n",
			sampleResult.When, sampleResult.Vid, sampleResult.Type, sampleResult.Title,
			sampleResult.Status, sampleResult.Size, sampleResult.Latency, cleanFullRequest)
		_, err = w.WriteString(s)
	}

	if err != nil {
		log.Fatal(err)
	}

	// Flush periodically rather than on every result: WriteResult is only ever
	// called from the single AcceptResults goroutine, so this stays race-free.
	if time.Since(lastFlush) >= flushInterval {
		w.Flush()
		lastFlush = time.Now()
	}
}
