package viewer

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strconv"

	"github.com/bmizerany/perks/quantile"
	log "github.com/sirupsen/logrus"
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

var (
	webUrl string = "/graphs/"

	//go:embed graphs/*
	content embed.FS
)

func BuildGraphs(datafile, scriptname, outputdir string) error {
	log.Debugf("BuildGraphs in output directory: %s", outputdir)
	// Creates the outputdir if needed
	stat, err := os.Stat(outputdir)
	if os.IsNotExist(err) {
		log.Debugf("Must create the Output Directory %s", outputdir)
		if err := os.MkdirAll(outputdir, 0755); err != nil {
			return fmt.Errorf("Cannot create Output Directory %s: %s", outputdir, err.Error())
		}
	} else if !stat.Mode().IsDir() {
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

	// Iterate through the records. Records which Type is "GLOBAL" are not samples
	/*
		Timestamp,Vid,Type,Title,Status,Size,Latency,FullRequest
		41353146,1249300000,HTTP,Page 1,200,17,40794348,http://www.delamarche.com/page1.php
		83251860,1249300000,HTTP,Page 2,200,11,41740141,http://www.delamarche.com/page2.php
		249059934,1249300000,HTTP,Page SSL,200,8083,163870870,https://www.delamarche.com:443/
	*/
	colUniqTitle := make(map[string]bool)
	uniqTitleCount := make(map[string]int)
	uniqTitleLatency := make(map[string]int)
	uniqTitleRcvBytes := make(map[string]int)
	colUniqStatus := make(map[int]bool)
	measures := make([]measure, 0, DFLT_CAP)
	internalVus := make(map[int]int)
	quantilesByPage := make(map[string]*quantile.Stream)

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
		curType := record[2]
		curVid, _ := strconv.ParseInt(record[1], 10, 64)
		curStatus, _ := strconv.ParseInt(record[4], 10, 64)
		curRecvBytes, _ := strconv.ParseInt(record[5], 10, 64)
		curLatency, _ := strconv.ParseFloat(record[6], 64)

		if curType != "GLOBAL" {
			title := record[3]
			m := measure{
				timestamp: int(curTime) / 1000000000,
				vid:       curVid,
				title:     title,
				status:    int(curStatus),
				recvBytes: int(curRecvBytes),
				latency:   int(curLatency) / 1000000,
			}
			measures = append(measures, m)

			colUniqTitle[title] = true
			uniqTitleCount[title]++
			uniqTitleLatency[title] += m.latency
			uniqTitleRcvBytes[title] += m.recvBytes
			colUniqStatus[int(curStatus)] = true
			if quantilesByPage[title] == nil {
				quantilesByPage[title] = quantile.NewTargeted(0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.99)
			}
			quantilesByPage[title].Insert(float64(m.latency))
			idx++
		} else {
			// The count of internal VU is stored in the Size field
			internalVus[int(curTime)/1000000000] += int(curRecvBytes)
		}
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

	// Compute VU count per second (not used anymore...)
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

	// Display the Quantile
	for title, _ := range quantilesByPage {
		log.Debug("quantiles for ", title)
		log.Debug("10% ", quantilesByPage[title].Query(0.1))
		log.Debug("20% ", quantilesByPage[title].Query(0.2))
		log.Debug("30% ", quantilesByPage[title].Query(0.3))
		log.Debug("40% ", quantilesByPage[title].Query(0.4))
		log.Debug("50% ", quantilesByPage[title].Query(0.5))
		log.Debug("60% ", quantilesByPage[title].Query(0.6))
		log.Debug("70% ", quantilesByPage[title].Query(0.7))
		log.Debug("80% ", quantilesByPage[title].Query(0.8))
		log.Debug("90% ", quantilesByPage[title].Query(0.9))
		log.Debug("99% ", quantilesByPage[title].Query(0.99))
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

	// Bug fixed: internalVus may miss some seconds if no answer rcvd for one second !
	// vus = make([]int, len(internalVus))
	vus = make([]int, total_elapsed_time)
	for k, v := range internalVus {
		if len(vus) > k-1 {
			vus[k-1] = v
		}
	}

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
			"#Appl Errors":    errors,
			"#Net Errors":     netErrors,
			"#Rcv Bytes":      rcvBytes,
		}, false)

	series := make(map[string][]int, len(colUniqTitle))
	for title, _ := range colUniqTitle {
		series[title] = meanTimePerReq[title]
	}
	graph(output,
		total_elapsed_time,
		"stats_per_req",
		"Latency per Request",
		"Elapsed Time (seconds)",
		"time(ms)",
		series,
		false)

	err_series := make(map[string][]int)
	for errCode, _ := range errorsPerSeconds {
		if errCode != -1 {
			err_series[strconv.Itoa(errCode)] = errorsPerSeconds[errCode]
		} else {
			err_series["Error"] = errorsPerSeconds[errCode]
		}
	}
	graph(output,
		total_elapsed_time,
		"errors_by_code",
		"Returned codes per second",
		"Elapsed Time (seconds)",
		"#err",
		err_series,
		false)

	// Compute stats with  #VU as x-values
	// Find the higher #VU and stops once it is reached
	max_idx := 0
	max_vus := vus[0]
	for second, value := range vus {
		if value > max_vus {
			max_vus = value
			max_idx = second
		}
	}
	log.Debugf("Maximum number of VU found is %d on second #%d", max_vus, max_idx)

	// Now we can build the new series
	latency_per_vu_series := make(map[string][]int, len(colUniqTitle))
	for title, _ := range colUniqTitle {
		latency_per_vu_series[title] = make([]int, max_vus+1)
		for second, vu := range vus {
			if second > max_idx {
				break
			}
			latency_per_vu_series[title][vu] = meanTimePerReq[title][second]
		}
	}
	graph(output,
		total_elapsed_time,
		"latency_per_vu",
		"Latency per VU",
		"#VU",
		"time(ms)",
		latency_per_vu_series,
		true) // compute regression

	// Output the quantile by page with bar graph
	bar_graph(output,
		"quantiles_per_page",
		"Deciles by Page",
		quantilesByPage)

	// We want the page sorted by title
	page_titles := make([]string, 0, len(colUniqTitle))
	for title := range colUniqTitle {
		page_titles = append(page_titles, title)
	}
	sort.Strings(page_titles)

	// Output the average response time per page
	firstRow = true
	row := ""
	//for title, count := range uniqTitleCount {
	for _, title := range page_titles {
		count := uniqTitleCount[title]
		log.Debugf("Page %s has %d count and %d total latency", title, count, uniqTitleLatency[title])
		if firstRow {
			firstRow = false
			row = "<tr><th>Page Title</th><th>#Req</th><th>Avg Response Time (in ms)</th><th>Avg Response Size (in Bytes)</th></tr>"
			fmt.Fprintf(output, "$('#avg_resp_by_page > thead').append('"+row+"');\n")
		}

		row = "<tr><td>" + title + "</td>"
		row += "<td>" + strconv.Itoa(count) + "</td>"
		row += "<td>" + strconv.Itoa(uniqTitleLatency[title]/count) + "</td>"
		row += "<td>" + IntComma(int(uniqTitleRcvBytes[title]/count)) + "</td>"
		row += "</tr>"
		fmt.Fprintf(output, "$('#avg_resp_by_page > tbody:last-child').append('"+row+"');\n")
	}

	// Output the HTTP Code array
	// First sort the HTTP codes (keys of the colUniqStatus map)
	var keys []int
	for k := range colUniqStatus {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	firstRow = true
	//for title, errs := range errorsByPage {
	for _, title := range page_titles {
		errs := errorsByPage[title]
		log.Debugf("errors for page %s: %v", title, errs)
		if firstRow {
			firstRow = false
			row = "<tr><th>Page Title</th>"
			for _, err := range keys {
				if err != -1 {
					row += "<th>" + strconv.Itoa(err) + "</th>"
				} else {
					row += "<th>Timeout & Network Error</th>"
				}
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

func graph(w *os.File, totalTime int, name, title, xtitle, ytitle string, series map[string][]int, regress bool) {
	fmt.Fprintf(w, "%s_options = {\n", name)
	fmt.Fprintf(w, "chart: {\n")
	if regress {
		fmt.Fprint(w, `events: {
			load: function() { 
				var theChart = this;
				var theSeries = this.series;
				for (var key in theSeries) {
					var low_idx = 0;
					var high_idx = -1;
					var aSeries = theSeries[key];
					var data = [];

					for (var idx in aSeries.data) {
						idx = parseInt(idx, 10);
						if (idx > high_idx) {
							for (var idx2 in aSeries.data) {
								idx2 = parseInt(idx2, 10);
								if (idx2 < idx) { continue; }
								if (aSeries.data[idx2].y != 0) {
									high_idx = idx2;
									break;
								}
							}
							if (high_idx == -1) { high_idx = 0; }
						}
						if (aSeries.data[idx].y == 0) {
							var y1 = aSeries.data[low_idx].y;
							var y2 = aSeries.data[high_idx].y;
							var x1 = low_idx;
							var x2 = high_idx;
							aSeries.points[idx].y = y1 + ((y2 - y1) / (x2 - x1)) * (idx - x1);
							data.push(aSeries.points[idx].y);
						} else {
							low_idx = idx;
							data.push(aSeries.data[idx].y);
						}
					}
					this.series[key].update({data: data}, true);
				}
			},
		},`)
	}
	fmt.Fprintf(w, "\nzoomType: 'x', panning: true, panKey: 'shift', renderTo: '%s' },\n", name)
	fmt.Fprintf(w, "title: { text: '%s'	},\n", title)
	fmt.Fprintf(w, "legend: { layout: 'horizontal',	align: 'center', verticalAlign: 'bottom', borderWidth: 0 },\n")
	fmt.Fprintf(w, "xAxis: { categories: [")
	for idx := 0; idx < totalTime; idx++ {
		fmt.Fprintf(w, "%d, ", idx)
	}
	fmt.Fprintf(w, "], title: { text: '%s' }, },\n", xtitle)
	if _, ok := series["#VU"]; ok {
		fmt.Fprintf(w, "yAxis: [{ title: { text: '%s' }, }, { title: { text: '#VU' }, opposite: true }],\n", ytitle)
	} else {
		fmt.Fprintf(w, "yAxis: { title: { text: '%s' }, },\n", ytitle)
	}
	fmt.Fprintf(w, "series: [\n")
	for k, v := range series {
		if k == "#VU" {
			fmt.Fprintf(w, "{ name: '%s', yAxis: 1, data: [", k)
		} else {
			fmt.Fprintf(w, "{ name: '%s', data: [", k)
		}
		for idx := 0; idx < len(v); idx++ {
			fmt.Fprintf(w, "%d, ", v[idx])
		}
		fmt.Fprintf(w, "]},\n")
	}
	fmt.Fprintf(w, "]\n")
	fmt.Fprintf(w, "};\n")
	fmt.Fprintf(w, "var %s = new Highcharts.chart(%s_options);\n", name, name)
}

func bar_graph(w *os.File, name, title string, quantiles map[string]*quantile.Stream) {
	fmt.Fprintf(w, "%s_options = {\n", name)
	fmt.Fprintf(w, "chart: {\n")
	fmt.Fprintf(w, "\ntype: 'bar', renderTo: '%s' },\n", name)
	fmt.Fprintf(w, "title: { text: '%s'	},\n", title)
	fmt.Fprintf(w, "legend: { layout: 'vertical', align: 'right', verticalAlign: 'top', borderWidth: 1, floating: true, shadow: true, x: -40, y: 80 },\n")
	fmt.Fprintf(w, "xAxis: { categories: [ '0.1', '0.2', '0.3', '0.4', '0.5', '0.6', '0.7', '0.8', '0.9', '0.99'], title: {text: null }, },\n")
	fmt.Fprintf(w, "yAxis: { min: 0, title: { text: 'Latency in ms', aligne: 'high' }, labels: {overflow: 'justify'}},\n")

	fmt.Fprintf(w, "series: [\n")
	quant := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.99}
	for k, v := range quantiles {
		fmt.Fprintf(w, "{ name: '%s', data: [", k)
		for _, qt := range quant {
			fmt.Fprintf(w, "%f, ", v.Query(qt))
		}
		fmt.Fprintf(w, "]},\n")
	}
	fmt.Fprintf(w, "]\n")
	fmt.Fprintf(w, "};\n")
	fmt.Fprintf(w, "var %s = new Highcharts.chart(%s_options);\n", name, name)
}

// Copy templates and js in output directory
func copyTemplates(outputdir string) error {
	sub_fs, _ := fs.Sub(content, "graphs")
	fs.WalkDir(sub_fs, ".", func(path string, info fs.DirEntry, err error) error {
		log.Debugf("walkFn processing: %s", path)
		//log.Debugf("info=%v", info)
		dest := outputdir + "/" + path

		// Regular file or directory ?
		if info.IsDir() {
			log.Debugf("This a directory")
			_, err := os.Stat(dest)
			if os.IsNotExist(err) {
				if err := os.Mkdir(dest, 0755); err != nil {
					log.Errorf("copyTemplates could not create directory %s (%s)", dest, err)
				}
			}
		} else {
			log.Debugf("This is a file")
			source, err := fs.ReadFile(sub_fs, path)
			if err != nil {
				log.Errorf("copyTemplates could not open file %s for reading (%s)", path, err)
				return err
			}

			if err := os.WriteFile(dest, source, 0666); err != nil {
				log.Errorf("copyTemplates could not write file %s for writing (%s)", dest, err)
				return err
			}
		}
		return nil
	})

	return nil
}

// Add commas to integer representation
func IntComma(i int) string {
	if (i < 0) {
		return "-" + IntComma(-i)
	}
	if (i < 1000) {
		return fmt.Sprintf("%d", i)
	}
	return IntComma(i / 1000) + "," + fmt.Sprintf("%03d", i % 1000)
}