package viewer

import (
	"fmt"
	"sort"
)

// Generate JSON result on stdout
func BuildJSON(datafile, scriptname string) error {
	results, err := computeResults(datafile, scriptname)
	if err != nil {
		return err
	}

	fmt.Println("{")
	fmt.Println("\t\"global\": {")
	fmt.Printf("\t\t\"total_elapsed_time\": %d\n", results.total_elapsed_time)
	fmt.Println("\t},")

	fmt.Println("\t\"overall_stats\": {")
	fmt.Println("\t\t\"vu_count\": {")
	for idx, value := range results.vus {
		if idx != 0 { fmt.Println(",") }
		fmt.Printf("\t\t\t\"%d\": %d", idx, value)
	}
	fmt.Println("\n\t\t},")
	fmt.Println("\t\t\"request_count\": {")
	for idx, value := range results.nbReq {
		if idx != 0 { fmt.Println(",") }
		fmt.Printf("\t\t\t\"%d\": %d", idx, value)
	}
	fmt.Println("\n\t\t},")
	fmt.Println("\t\t\"latency_in_ms\": {")
	for idx, value := range results.meanTime {
		if idx != 0 { fmt.Println(",") }
		fmt.Printf("\t\t\t\"%d\": %d", idx, value)
	}
	fmt.Println("\n\t\t},")
	fmt.Println("\t\t\"error_count\": {")
	for idx, value := range results.errors {
		if idx != 0 { fmt.Println(",") }
		fmt.Printf("\t\t\t\"%d\": %d", idx, value)
	}
	fmt.Println("\n\t\t},")
	fmt.Println("\t\t\"network_error_count\": {")
	for idx, value := range results.netErrors {
		if idx != 0 { fmt.Println(",") }
		fmt.Printf("\t\t\t\"%d\": %d", idx, value)
	}
	fmt.Println("\n\t\t},")
	fmt.Println("\t\t\"received_bytes\": {")
	for idx, value := range results.rcvBytes {
		if idx != 0 { fmt.Println(",") }
		fmt.Printf("\t\t\t\"%d\": %d", idx, value)
	}
	fmt.Println("\n\t\t}")
	fmt.Println("\t},")

	fmt.Println("\t\"latency_per_request\": {")
	first := true
	for req, serie := range results.series {
		if !first { fmt.Println(",") }
		first = false
		fmt.Printf("\t\t\"%s\": {\n", req)
		for idx, value := range serie {
			if idx != 0 { fmt.Println(",") }
			fmt.Printf("\t\t\t\"%d\": %d", idx, value)
		}
		fmt.Print("\n\t\t}")
	}
	fmt.Println("\n\t},")

	fmt.Println("\t\"returned_code_per_second\": {")
	first = true
	for code, serie := range results.err_series {
		if !first { fmt.Println(",") }
		first = false
		fmt.Printf("\t\t\"%s\": {\n", code)
		for idx, value := range serie {
			if idx != 0 { fmt.Println(",") }
			fmt.Printf("\t\t\t\"%d\": %d", idx, value)
		}
		fmt.Print("\n\t\t}")
	}
	fmt.Println("\n\t},")

	fmt.Println("\t\"latency_per_vu\": {")
	first = true
	for req, serie := range results.latency_per_vu_series {
		if !first { fmt.Println(",") }
		first = false
		fmt.Printf("\t\t\"%s\": {\n", req)
		for idx, value := range serie {
			if idx != 0 { fmt.Println(",") }
			fmt.Printf("\t\t\t\"%d\": %d", idx, value)
		}
		fmt.Print("\n\t\t}")
	}
	fmt.Println("\n\t},")

	fmt.Println("\t\"decile_per_page\": {")
	first = true
	for title, _ := range results.quantilesByPage {
		if !first { fmt.Println(",") }
		first = false
		fmt.Printf("\t\t\"%s\": {\n", title)
		fmt.Printf("\t\t\t\"10%%\": %d,\n", int(results.quantilesByPage[title].Query(0.1)))
		fmt.Printf("\t\t\t\"20%%\": %d,\n", int(results.quantilesByPage[title].Query(0.2)))
		fmt.Printf("\t\t\t\"30%%\": %d,\n", int(results.quantilesByPage[title].Query(0.3)))
		fmt.Printf("\t\t\t\"40%%\": %d,\n", int(results.quantilesByPage[title].Query(0.4)))
		fmt.Printf("\t\t\t\"50%%\": %d,\n", int(results.quantilesByPage[title].Query(0.5)))
		fmt.Printf("\t\t\t\"60%%\": %d,\n", int(results.quantilesByPage[title].Query(0.6)))
		fmt.Printf("\t\t\t\"70%%\": %d,\n", int(results.quantilesByPage[title].Query(0.7)))
		fmt.Printf("\t\t\t\"80%%\": %d,\n", int(results.quantilesByPage[title].Query(0.8)))
		fmt.Printf("\t\t\t\"90%%\": %d,\n", int(results.quantilesByPage[title].Query(0.9)))
		fmt.Printf("\t\t\t\"99%%\": %d\n", int(results.quantilesByPage[title].Query(0.99)))
		fmt.Print("\t\t}")
	}
	fmt.Println("\n\t},")

	fmt.Println("\t\"average_response_time_per_page\": {")
	first = true
	for _, title := range results.pageTitles {
		if !first { fmt.Println(",") }
		first = false
		fmt.Printf("\t\t\"%s\": {\n", title)
		count := results.uniqTitleCount[title]
		fmt.Printf("\t\t\t\"request_count\": %d,\n", count)
		fmt.Printf("\t\t\t\"average_response_time\": %d,\n", results.uniqTitleLatency[title]/count)
		fmt.Printf("\t\t\t\"average_response_size\": %d\n", results.uniqTitleRcvBytes[title]/count)
		fmt.Print("\t\t}")
	}
	fmt.Println("\n\t},")

	// First sort the HTTP codes (keys of the colUniqStatus map)
	var keys []int
	for k := range results.colUniqStatus {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	fmt.Println("\t\"returned_code_by_page\": {")
	first = true
	for _, title := range results.pageTitles {
		errs := results.errorsByPage[title]
		if !first { fmt.Println(",") }
		first = false
		fmt.Printf("\t\t\"%s\": {\n", title)
		for _, err := range keys {
			if val, ok := errs[err]; ok {
				fmt.Printf("\t\t\t\"%d\": %d,\n", err, val)
			}
		}
		fmt.Printf("\t\t\t\"total\": %d\n", results.uniqTitleCount[title])
		fmt.Print("\t\t}")
	}
	fmt.Println("\n\t}")

	fmt.Println("}")

	return nil
}
