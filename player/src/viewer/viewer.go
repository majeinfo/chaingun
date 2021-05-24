package viewer

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"

	"github.com/bmizerany/perks/quantile"
	log "github.com/sirupsen/logrus"
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

	results, err := computeResults(datafile, scriptname)
	if err != nil {
		return err
	}

	// Create the result file (data.js)
	outputfilename := outputdir + "/data.js"
	output, err := os.Create(outputfilename)
	defer output.Close()
	if err != nil {
		return fmt.Errorf("Could not create result file %s: %s", outputfilename, err.Error())
	}

	fmt.Fprintf(output, "var elapsed_time = %d;\n", results.total_elapsed_time)
	fmt.Fprintf(output, "var total_requests = %d;\n", results.total_requests)
	fmt.Fprintf(output, "var total_netErrors = %d;\n", results.total_netErrors)
	fmt.Fprintf(output, "var playbook_name = \"%s\";\n\n", scriptname)

	fmt.Fprintf(output, "$(function () {\n")

	graph(output,
		results.total_elapsed_time,
		"overall_stats",
		"Overall Statistics per Second",
		"Elapsed Time (seconds)",
		"",
		map[string][]int{
			"#VU":             results.vus,
			"#Req":            results.nbReq,
			"Latency (in ms)": results.meanTime,
			"#Appl Errors":    results.errors,
			"#Net Errors":     results.netErrors,
			"#Rcv Bytes":      results.rcvBytes,
		}, false)

	series := make(map[string][]int, len(results.colUniqTitle))
	for title, _ := range results.colUniqTitle {
		series[title] = results.meanTimePerReq[title]
	}
	graph(output,
		results.total_elapsed_time,
		"stats_per_req",
		"Latency per Request",
		"Elapsed Time (seconds)",
		"time(ms)",
		series,
		false)

	err_series := make(map[string][]int)
	for errCode, _ := range results.errorsPerSecond {
		if errCode != -1 {
			err_series[strconv.Itoa(errCode)] = results.errorsPerSecond[errCode]
		} else {
			err_series["Error"] = results.errorsPerSecond[errCode]
		}
	}
	graph(output,
		results.total_elapsed_time,
		"errors_by_code",
		"Returned codes per second",
		"Elapsed Time (seconds)",
		"#err",
		err_series,
		false)

	// Compute stats with  #VU as x-values
	// Find the higher #VU and stops once it is reached
	max_idx := 0
	max_vus := results.vus[0]
	for second, value := range results.vus {
		if value > max_vus {
			max_vus = value
			max_idx = second
		}
	}
	log.Debugf("Maximum number of VU found is %d on second #%d", max_vus, max_idx)

	// Now we can build the new series
	latency_per_vu_series := make(map[string][]int, len(results.colUniqTitle))
	for title, _ := range results.colUniqTitle {
		latency_per_vu_series[title] = make([]int, max_vus+1)
		for second, vu := range results.vus {
			if second > max_idx {
				break
			}
			latency_per_vu_series[title][vu] = results.meanTimePerReq[title][second]
		}
	}
	graph(output,
		results.total_elapsed_time,
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
		results.quantilesByPage)

	// We want the page sorted by title
	page_titles := make([]string, 0, len(results.colUniqTitle))
	for title := range results.colUniqTitle {
		page_titles = append(page_titles, title)
	}
	sort.Strings(page_titles)

	// Output the average response time per page
	firstRow := true
	row := ""
	//for title, count := range uniqTitleCount {
	for _, title := range page_titles {
		count := results.uniqTitleCount[title]
		log.Debugf("Page %s has %d count and %d total latency", title, count, results.uniqTitleLatency[title])
		if firstRow {
			firstRow = false
			row = "<tr><th>Page Title</th><th>#Req</th><th>Avg Response Time (in ms)</th><th>Avg Response Size (in Bytes)</th></tr>"
			fmt.Fprintf(output, "$('#avg_resp_by_page > thead').append('"+row+"');\n")
		}

		row = "<tr><td>" + title + "</td>"
		row += "<td>" + strconv.Itoa(count) + "</td>"
		row += "<td>" + strconv.Itoa(results.uniqTitleLatency[title]/count) + "</td>"
		row += "<td>" + IntComma(int(results.uniqTitleRcvBytes[title]/count)) + "</td>"
		row += "</tr>"
		fmt.Fprintf(output, "$('#avg_resp_by_page > tbody:last-child').append('"+row+"');\n")
	}

	// Output the HTTP Code array
	// First sort the HTTP codes (keys of the colUniqStatus map)
	var keys []int
	for k := range results.colUniqStatus {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	firstRow = true
	//for title, errs := range errorsByPage {
	for _, title := range page_titles {
		errs := results.errorsByPage[title]
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
