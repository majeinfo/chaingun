package viewer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
	_ "statik"
)

const (
	DFLT_CAP = 3000
)

type measure struct {
	timestamp int
	vid       int64
	title     string
	status    int
	recvBytes int
	latency   int
}

func BuildGraphs(datafile, scriptname, outputdir string) error {
	// Creates the outputdir if needed
	stat, err := os.Stat(outputdir)
	if os.IsNotExist(err) {
		log.Debugf("Must create the Output Directory")
		if err := os.MkdirAll(outputdir, 0755); err != nil {
			return fmt.Errorf("Cannot create Output Directory %s: %s", outputdir, err.Error())
		}
	} else if stat.Mode().IsRegular() {
		return fmt.Errorf("Output Directory %s already exists as a file !", outputdir)
	}

	// Copy templates in outputdir
	if err := copyTemplates(outputdir); err != nil {
		return err
	}

	// Read the data file (csv)
	csvfile, err := os.Open(datafile)
	if err != nil {
		return fmt.Errorf("Couldn't open the csv file %s: %s", datafile, err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	// Iterate through the records
	/*
		Timestamp,Vid,Type,Title,Status,Size,Latency,FullRequest
		41353146,1249300000,HTTP,Page 1,200,17,40794348,http://www.delamarche.com/page1.php
		83251860,1249300000,HTTP,Page 2,200,11,41740141,http://www.delamarche.com/page2.php
		249059934,1249300000,HTTP,Page SSL,200,8083,163870870,https://www.delamarche.com:443/
	*/
	colUniqTitle := make(map[string]bool)
	colUniqStatus := make(map[int]bool)
	measures := make([]measure, 0, DFLT_CAP)

	idx := 0
	firstRow := true

	// Read the raw data from CSV File
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Errorf("Error while reading CSV file %s: %s", datafile, err)
			break
		}
		if idx < 10 {
			log.Debugf("%v", record)
		}
		// Skip the first line
		if firstRow {
			firstRow = false
			continue
		}

		curTime, _ := strconv.ParseFloat(record[0], 64)
		curVid, _ := strconv.ParseInt(record[1], 10, 64)
		curStatus, _ := strconv.ParseInt(record[4], 10, 64)
		curRecvBytes, _ := strconv.ParseInt(record[5], 10, 64)
		curLatency, _ := strconv.ParseFloat(record[6], 64)

		m := measure{
			timestamp: int(curTime) / 1000000000,
			vid:       curVid,
			title:     record[3],
			status:    int(curStatus),
			recvBytes: int(curRecvBytes),
			latency:   int(curLatency) / 1000000,
		}
		measures = append(measures, m)

		colUniqTitle[record[3]] = true
		colUniqStatus[int(curStatus)] = true

		idx++
	}

	// Empty file ?
	if idx == 0 {
		return fmt.Errorf("Datafile %s does not contain any data !", datafile)
	}

	// Sort the measures
	sort.Slice(measures, func(i, j int) bool {
		return measures[i].timestamp < measures[j].timestamp
	})

	// Compute stats per time
	total_elapsed_time := measures[idx-1].timestamp - measures[0].timestamp + 1
	total_requests := len(measures)
	total_netErrors := 0
	log.Debugf("Elapsed seconds=%d", total_elapsed_time)

	vus := make([]int, total_elapsed_time)
	vusSet := make(map[int]map[int64]bool)
	nbReq := make([]int, total_elapsed_time)
	meanTime := make([]int, total_elapsed_time)
	meanTimePerReq := make(map[string][]int, total_elapsed_time)
	reqCountPerTime := make(map[string][]int, total_elapsed_time)
	errors := make([]int, total_elapsed_time)
	netErrors := make([]int, total_elapsed_time)
	rcvBytes := make([]int, total_elapsed_time)

	for idx = 0; idx < total_requests; idx++ {
		nuSec := measures[idx].timestamp - measures[0].timestamp

		// With merged file, we should order the lines to compute the real elapsed time, so we must make
		// a consistency check :
		if nuSec >= total_elapsed_time {
			log.Warningf("Result line %d ignored: out of bounds", idx)
			continue
		}

		if measures[idx].status < 0 {
			netErrors[nuSec] += 1
			total_netErrors += 1
			total_requests -= 1
			continue
		}

		nbReq[nuSec] += 1
		if measures[idx].status >= 400 {
			errors[nuSec] += 1
		}
		rcvBytes[nuSec] += measures[idx].recvBytes
		meanTime[nuSec] += measures[idx].latency
		if vusSet[nuSec] == nil {
			vusSet[nuSec] = make(map[int64]bool)
		}
		vusSet[nuSec][measures[idx].vid] = true
	}

	// Compute latency average
	for idx := 0; idx < len(meanTime); idx++ {
		if nbReq[idx] > 0 {
			meanTime[idx] = int(meanTime[idx] / nbReq[idx])
		}
	}

	// Compute VU count per second
	for idx := 0; idx < len(vus); idx++ {
		vus[idx] = len(vusSet[idx])
	}

	// Compute Latency per Title
	for title, _ := range colUniqTitle {
		meanTimePerReq[title] = make([]int, total_elapsed_time)
		reqCountPerTime[title] = make([]int, total_elapsed_time)
	}
	for idx := 0; idx < total_requests; idx++ {
		nuSec := measures[idx].timestamp - measures[0].timestamp
		log.Debugf("idx=%d, colTitle[idx]=%s, nuSec=%d", idx, measures[idx].title, nuSec)
		// With merged file, we should order the lines to compute the real elapsed time, so we must make
		// a consistency check :
		if nuSec >= total_elapsed_time {
			continue
		}

		meanTimePerReq[measures[idx].title][nuSec] += measures[idx].latency
		reqCountPerTime[measures[idx].title][nuSec]++
	}
	for title, _ := range colUniqTitle {
		for idx := 0; idx < total_elapsed_time; idx++ {
			if reqCountPerTime[title][idx] > 0 {
				meanTimePerReq[title][idx] = int(meanTimePerReq[title][idx] / reqCountPerTime[title][idx])
			}
		}
	}

	// Compute error stats for each request
	errorsPerSeconds := make(map[int][]int, total_elapsed_time)
	for errCode, _ := range colUniqStatus {
		errorsPerSeconds[errCode] = make([]int, total_elapsed_time)
	}
	for idx := 0; idx < total_requests; idx++ {
		nuSec := measures[idx].timestamp - measures[0].timestamp
		// With merged file, we should order the lines to compute the real elapsed time, so we must make
		// a consistency check :
		if nuSec >= total_elapsed_time {
			continue
		}

		errorsPerSeconds[measures[idx].status][nuSec]++
	}

	// Compute errors per request/page
	errorsByPage := make(map[string]map[int]int, len(colUniqTitle))
	for title, _ := range colUniqTitle {
		errorsByPage[title] = make(map[int]int, len(colUniqStatus))
	}
	for idx := 0; idx < total_requests; idx++ {
		errorsByPage[measures[idx].title][measures[idx].status]++
	}

	// Create the result file (data.js)
	outputfilename := outputdir + "/data.js"
	output, err := os.Create(outputfilename)
	defer output.Close()
	if err != nil {
		return fmt.Errorf("Could not create result file %s: %s", outputfilename, err.Error())
	}

	fmt.Fprintf(output, "var elapsed_time = %d;\n", total_elapsed_time)
	fmt.Fprintf(output, "var total_requests = %d;\n", total_requests)
	fmt.Fprintf(output, "var total_netErrors = %d;\n", total_netErrors)
	fmt.Fprintf(output, "var playbook_name = \"%s\";\n\n", scriptname)

	fmt.Fprintf(output, "$(function () {\n")

	graph(output,
		total_elapsed_time,
		"overall_stats",
		"Overall Statistics per Second",
		"Elapsed Time (seconds)",
		"",
		map[string][]int{
			"#VU":             vus,
			"#Req":            nbReq,
			"Latency (in ms)": meanTime,
			"#HTTP Errors":    errors,
			"#Net Errors":     netErrors,
			"#Rcv Bytes":      rcvBytes,
		})

	series := make(map[string][]int, len(colUniqTitle))
	for title, _ := range colUniqTitle {
		series[title] = meanTimePerReq[title]
	}
	graph(output,
		total_elapsed_time,
		"stats_per_req",
		"Latency per Request (in ms)",
		"Elapsed Time (seconds)",
		"time(ms)",
		series)

	err_series := make(map[int][]int)
	for errCode, _ := range errorsPerSeconds {
		err_series[errCode] = errorsPerSeconds[errCode]
	}
	graph(output,
		total_elapsed_time,
		"errors_by_code",
		"HTTP return codes per second",
		"Elapsed Time (seconds)",
		"#err",
		series)

	// Output the HTTP Code array
	// First sort the HTTP codes (keys of the colUniqStatus map)
	var keys []int
	for k := range colUniqStatus {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	firstRow = true
	row := ""
	for title, errs := range errorsByPage {
		log.Debugf("errors for page %s: %v", title, errs)
		if firstRow {
			firstRow = false
			row = "<tr><th></th>"
			for _, err := range keys {
				row += "<th>" + strconv.Itoa(err) + "</th>"
			}
			row += "<th>#Req</th></tr>"
			fmt.Fprintf(output, "$('#http_codes > thead').append('"+row+"');\n")
		}

		row = "<tr><td>" + title + "</td>"
		total := 0
		for _, err := range keys {
			row += "<td>"
			if val, ok := errs[err]; ok {
				row += strconv.Itoa(val)
				total += val
			}
			row += "</td>"
		}

		row += "<td>" + strconv.Itoa(total) + "</td></tr>"
		fmt.Fprintf(output, "$('#http_codes > tbody:last-child').append('"+row+"');\n")
	}

	fmt.Fprintf(output, "});\n")

	return nil
}

func graph(w *os.File, totalTime int, name, title, xtitle, ytitle string, series map[string][]int) {
	fmt.Fprintf(w, "%s_options = {\n", name)
	fmt.Fprintf(w, "chart: { zoomType: 'x', panning: true, panKey: 'shift', renderTo: '%s' },\n", name)
	fmt.Fprintf(w, "title: { text: '%s'	},\n", title)
	fmt.Fprintf(w, "legend: { layout: 'horizontal',	align: 'center', verticalAlign: 'bottom', borderWidth: 0 },\n")
	fmt.Fprintf(w, "xAxis: { categories: [")
	for idx := 0; idx < totalTime; idx++ {
		fmt.Fprintf(w, "%d, ", idx)
	}
	fmt.Fprintf(w, "], title: { text: '%s' }, },\n", title)
	fmt.Fprintf(w, "yAxis: { title: { text: '%s' }, },\n", ytitle)
	fmt.Fprintf(w, "series: [\n")
	for k, v := range series {
		fmt.Fprintf(w, "{ name: '%s', data: [", k)
		for idx := 0; idx < len(v); idx++ {
			fmt.Fprintf(w, "%d, ", v[idx])
		}
		fmt.Fprintf(w, "]},\n")
	}
	fmt.Fprintf(w, "]\n")
	fmt.Fprintf(w, "};\n")
	fmt.Fprintf(w, "var %s = new Highcharts.chart(%s_options);\n", name, name)
}

// Copy templates and js in output directory
func copyTemplates(outputdir string) error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}
	fs.Walk(statikFS, "/graphs", func(path string, info os.FileInfo, err error) error {
		log.Debugf("walkFn processing: %s", path)
		dest := outputdir + "/" + path[len("/graphs"):]

		// Regular file or directory ?
		if info.IsDir() {
			_, err := os.Stat(dest)
			if os.IsNotExist(err) {
				if err := os.Mkdir(dest, 0755); err != nil {
					log.Errorf("copyTemplates could not create directory %s (%s)", dest, err)
				}
			}
		} else {
			source, err := statikFS.Open(path)
			if err != nil {
				log.Errorf("copyTemplates could not open file %s for reading (%s)", path, err)
				return err
			}
			defer source.Close()

			destination, err := os.Create(dest)
			if err != nil {
				log.Errorf("copyTemplates could not open file %s for writing (%s)", dest, err)
				return err
			}
			defer destination.Close()
			_, err = io.Copy(destination, source)
			return err
		}
		return nil
	})

	return nil
}
