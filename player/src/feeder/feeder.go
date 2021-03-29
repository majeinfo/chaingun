package feeder

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

var data []map[string]string
var index = 0

var lock sync.Mutex

// "public" synchronized channel for delivering feeder data
// must be initialized in case of reference to a Feeder but with a wrong data file
var FeedChannel chan map[string]string = make(chan map[string]string)

// Return the next data line from feeder
func NextFromFeeder() {

	if data != nil && len(data) > 0 {
		// Push data into the FeedChannel
		// log.Debugf("Current index: %d of total size: %d", index, len(data))
		lock.Lock()
		FeedChannel <- data[index]

		if index < len(data)-1 {
			index += 1
		} else {
			index = 0
		}
		lock.Unlock()
	} else {
		log.Error("NextFromFeeder called but no data to feed with")
		dummy := map[string]string{ // avoid a daemon crash
			"": "",
		}
		FeedChannel <- dummy
	}
}

func Csv(feeder config.Feeder, dirname string) bool {
	var filename = feeder.Filename

	filename = utils.ComputeFilename(feeder.Filename, dirname)

	log.Debugf("Read CSV File %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Cannot open CSV file: %s", err)
		return false
	}

	scanner := bufio.NewScanner(file)
	var lines int

	data = make([]map[string]string, 0, 0)

	// Scan the first line, should contain headers.
	scanner.Scan()
	headers := strings.Split(scanner.Text(), feeder.Separator)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), feeder.Separator)
		if len(headers) != len(line) {
			// Ignore line
			log.Infof("Number of columns mismatch with headers: line %d skipped", lines)
			continue
		}
		item := make(map[string]string)
		for n := 0; n < len(headers); n++ {
			item[headers[n]] = line[n]
		}
		data = append(data, item)
		lines++
	}
	log.Debug(data)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	index = 0
	log.Infof("CSV feeder fed with %d lines of data", lines)
	FeedChannel = make(chan map[string]string)

	return true
}

func CsvInline(feeder config.Feeder, full_content string) bool {
	var lines int

	data = make([]map[string]string, 0, 0)

	// Scan the first line, should contain headers.
	all_lines := strings.Split(full_content, "\n")
	var headers []string

	for idx, line := range all_lines {
		log.Debug(idx)
		if idx == 0 {
			headers = strings.Split(line, feeder.Separator)
			log.Debug(headers)
		} else {
			cols := strings.Split(line, feeder.Separator)
			if len(headers) != len(cols) {
				// Ignore line
				log.Infof("Number of columns mismatch with headers: line %d skipped", lines)
				continue
			}
			item := make(map[string]string)
			for n := 0; n < len(headers); n++ {
				item[headers[n]] = cols[n]
			}
			data = append(data, item)
			lines++
		}
	}
	log.Debug(data)

	index = 0
	log.Infof("CSV feeder fed with %d lines of data", lines)
	FeedChannel = make(chan map[string]string)

	return true
}

// EOF
