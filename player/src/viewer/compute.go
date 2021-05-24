package viewer

import (
	"encoding/csv"
	"fmt"
	"github.com/bmizerany/perks/quantile"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sort"
	"strconv"
)

const (
	DFLT_CAP = 3000
)

// TODO: simplify and remove unused fields ?
type test_results struct {
	measures []measure
	colUniqTitle map[string]bool
	uniqTitleCount map[string]int
	uniqTitleLatency map[string]int
	uniqTitleRcvBytes map[string]int
	colUniqStatus map[int]bool
	internalVus map[int]int
	quantilesByPage map[string]*quantile.Stream
	pageTitles []string
	series map[string][]int
	err_series map[string][]int
	latency_per_vu_series map[string][]int
	errorsByPage map[string]map[int]int
	total_elapsed_time int
	// Values per seconds
	vus []int
	vusSet map[int]map[int64]bool
	nbReq []int
	meanTime []int
	meanTimePerReq map[string][]int
	reqCountPerTime map[string][]int
	errors []int
	netErrors []int
	rcvBytes []int
}

// Function that computes the results from the generated CSV files
func computeResults(datafile, scriptname string) (*test_results, error) {
	log.Debugf("computeResults for: %s and script %s", datafile, scriptname)

	// Read the data file (csv)
	csvfile, err := os.Open(datafile)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open the csv file %s: %s", datafile, err)
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
	results := test_results{
		colUniqTitle: make(map[string]bool),
		uniqTitleCount: make(map[string]int),
		uniqTitleLatency: make(map[string]int),
		uniqTitleRcvBytes: make(map[string]int),
		colUniqStatus: make(map[int]bool),
		measures: make([]measure, 0, DFLT_CAP),
		internalVus: make(map[int]int),
		quantilesByPage: make(map[string]*quantile.Stream),
	}

	idx := 0
	firstRow := true

	// Read the raw data from CSV File
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error while reading CSV file %s: %s", datafile, err)
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
			results.measures = append(results.measures, m)

			results.colUniqTitle[title] = true
			results.uniqTitleCount[title]++
			results.uniqTitleLatency[title] += m.latency
			results.uniqTitleRcvBytes[title] += m.recvBytes
			results.colUniqStatus[int(curStatus)] = true
			if results.quantilesByPage[title] == nil {
				results.quantilesByPage[title] = quantile.NewTargeted(0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.99)
			}
			results.quantilesByPage[title].Insert(float64(m.latency))
			idx++
		} else {
			// The count of internal VU is stored in the Size field
			results.internalVus[int(curTime)/1000000000] += int(curRecvBytes)
		}
	}

	// Empty datafile ?
	if idx == 0 {
		return nil, fmt.Errorf("Datafile %s does not contain any data !", datafile)
	}

	// Sort the measures
	sort.Slice(results.measures, func(i, j int) bool {
		return results.measures[i].timestamp < results.measures[j].timestamp
	})

	// Compute stats per time
	results.total_elapsed_time = results.measures[idx-1].timestamp - results.measures[0].timestamp + 1
	total_requests := len(results.measures)
	total_netErrors := 0
	log.Debugf("Elapsed seconds=%d", results.total_elapsed_time)

	results.vus = make([]int, results.total_elapsed_time)
	results.vusSet = make(map[int]map[int64]bool)
	results.nbReq = make([]int, results.total_elapsed_time)
	results.meanTime = make([]int, results.total_elapsed_time)
	results.meanTimePerReq = make(map[string][]int, results.total_elapsed_time)
	results.reqCountPerTime = make(map[string][]int, results.total_elapsed_time)
	results.errors = make([]int, results.total_elapsed_time)
	results.netErrors = make([]int, results.total_elapsed_time)
	results.rcvBytes = make([]int, results.total_elapsed_time)

	for idx = 0; idx < total_requests; idx++ {
		nuSec := results.measures[idx].timestamp - results.measures[0].timestamp

		// With merged file, we should order the lines to compute the real elapsed time, so we must make
		// a consistency check :
		if nuSec >= results.total_elapsed_time {
			log.Warningf("Result line %d ignored: out of bounds", idx)
			continue
		}

		if results.measures[idx].status < 0 {
			results.netErrors[nuSec] += 1
			total_netErrors += 1
			total_requests -= 1
			continue
		}

		results.nbReq[nuSec] += 1
		if results.measures[idx].status >= 400 {
			results.errors[nuSec] += 1
		}
		results.rcvBytes[nuSec] += results.measures[idx].recvBytes
		results.meanTime[nuSec] += results.measures[idx].latency
		if results.vusSet[nuSec] == nil {
			results.vusSet[nuSec] = make(map[int64]bool)
		}
		results.vusSet[nuSec][results.measures[idx].vid] = true
	}

	// Compute latency average
	for idx := 0; idx < len(results.meanTime); idx++ {
		if results.nbReq[idx] > 0 {
			results.meanTime[idx] = int(results.meanTime[idx] / results.nbReq[idx])
		}
	}

	// Compute VU count per second (not used anymore...)
	for idx := 0; idx < len(results.vus); idx++ {
		results.vus[idx] = len(results.vusSet[idx])
	}

	// Compute Latency per Title
	for title, _ := range results.colUniqTitle {
		results.meanTimePerReq[title] = make([]int, results.total_elapsed_time)
		results.reqCountPerTime[title] = make([]int, results.total_elapsed_time)
	}
	for idx := 0; idx < total_requests; idx++ {
		nuSec := results.measures[idx].timestamp - results.measures[0].timestamp
		log.Debugf("idx=%d, colTitle[idx]=%s, nuSec=%d", idx, results.measures[idx].title, nuSec)
		// With merged file, we should order the lines to compute the real elapsed time, so we must make
		// a consistency check :
		if nuSec >= results.total_elapsed_time {
			continue
		}

		results.meanTimePerReq[results.measures[idx].title][nuSec] += results.measures[idx].latency
		results.reqCountPerTime[results.measures[idx].title][nuSec]++
	}
	for title, _ := range results.colUniqTitle {
		for idx := 0; idx < results.total_elapsed_time; idx++ {
			if results.reqCountPerTime[title][idx] > 0 {
				results.meanTimePerReq[title][idx] = int(results.meanTimePerReq[title][idx] / results.reqCountPerTime[title][idx])
			}
		}
	}

	// Compute error stats for each request
	errorsPerSeconds := make(map[int][]int, results.total_elapsed_time)
	for errCode, _ := range results.colUniqStatus {
		errorsPerSeconds[errCode] = make([]int, results.total_elapsed_time)
	}
	for idx := 0; idx < total_requests; idx++ {
		nuSec := results.measures[idx].timestamp - results.measures[0].timestamp
		// With merged file, we should order the lines to compute the real elapsed time, so we must make
		// a consistency check :
		if nuSec >= results.total_elapsed_time {
			continue
		}

		errorsPerSeconds[results.measures[idx].status][nuSec]++
	}

	// Compute errors per request/page
	results.errorsByPage = make(map[string]map[int]int, len(results.colUniqTitle))
	for title, _ := range results.colUniqTitle {
		results.errorsByPage[title] = make(map[int]int, len(results.colUniqStatus))
	}
	for idx := 0; idx < total_requests; idx++ {
		results.errorsByPage[results.measures[idx].title][results.measures[idx].status]++
	}

	// Display the Quantile
	for title, _ := range results.quantilesByPage {
		log.Debug("quantiles for ", title)
		log.Debug("10% ", results.quantilesByPage[title].Query(0.1))
		log.Debug("20% ", results.quantilesByPage[title].Query(0.2))
		log.Debug("30% ", results.quantilesByPage[title].Query(0.3))
		log.Debug("40% ", results.quantilesByPage[title].Query(0.4))
		log.Debug("50% ", results.quantilesByPage[title].Query(0.5))
		log.Debug("60% ", results.quantilesByPage[title].Query(0.6))
		log.Debug("70% ", results.quantilesByPage[title].Query(0.7))
		log.Debug("80% ", results.quantilesByPage[title].Query(0.8))
		log.Debug("90% ", results.quantilesByPage[title].Query(0.9))
		log.Debug("99% ", results.quantilesByPage[title].Query(0.99))
	}

	// Bug fixed: internalVus may miss some seconds if no answer rcvd for one second !
	// vus = make([]int, len(internalVus))
	results.vus = make([]int, results.total_elapsed_time)
	for k, v := range results.internalVus {
		if len(results.vus) > k-1 {
			results.vus[k-1] = v
		}
	}

	results.series = make(map[string][]int, len(results.colUniqTitle))
	for title, _ := range results.colUniqTitle {
		results.series[title] = results.meanTimePerReq[title]
	}

	results.err_series = make(map[string][]int)
	for errCode, _ := range errorsPerSeconds {
		if errCode != -1 {
			results.err_series[strconv.Itoa(errCode)] = errorsPerSeconds[errCode]
		} else {
			results.err_series["Error"] = errorsPerSeconds[errCode]
		}
	}

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
	results.latency_per_vu_series = make(map[string][]int, len(results.colUniqTitle))
	for title, _ := range results.colUniqTitle {
		results.latency_per_vu_series[title] = make([]int, max_vus+1)
		for second, vu := range results.vus {
			if second > max_idx {
				break
			}
			results.latency_per_vu_series[title][vu] = results.meanTimePerReq[title][second]
		}
	}

	// We want the pages be sorted by title
	results.pageTitles = make([]string, 0, len(results.colUniqTitle))
	for title := range results.colUniqTitle {
		results.pageTitles = append(results.pageTitles, title)
	}
	sort.Strings(results.pageTitles)

	return &results, nil
}
