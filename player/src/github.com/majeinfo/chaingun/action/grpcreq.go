package action

import (
	_ "bytes"
	_ "fmt"
	_ "io/ioutil"
	_ "net"
	"net/http"
	_ "strconv"
	_ "strings"
	_ "time"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	//"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_GRPC string = "gRPC"
)

// DoGRPCRequest accepts a GRPCAction and a one-way channel to write the results to.
func DoGRPCRequest(grpcAction GRPCAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	//var trace_req string
	//sampleReqResult := buildSampleResult(REPORTER_GRPC, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, grpcAction.Title, "")

	/*
	req, err := buildGRPCRequest(grpcAction, sessionMap, vulog)
	if err != nil {
		vulog.Error(err)
		return false
	}

	if must_trace_request {
		trace_req = fmt.Sprintf("%s %s", req.Call, req.Data)
	} else {
		vulog.Debugf("New Request: Call: %s, Data: %s", req.Call, req.Data)
	}

	// Try to substitute the server name by an IP address
	if !disable_dns_cache {
		if addr, status := utils.GetServerAddress(req.Host); status == true {
			req.URL.Host = addr
		}
	}

	start := time.Now()

	var DefaultTransport http.RoundTripper
	if httpAction.UseHTTP2 {
		DefaultTransport = &http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		DefaultTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(playbook.Timeout) * time.Second,
				KeepAlive: time.Duration(playbook.Timeout) * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: time.Duration(playbook.Timeout) * time.Second,
			DisableKeepAlives:     true,
		}
	}

	resp, err := DefaultTransport.RoundTrip(req)
	vulog.Debugf("%v", resp)

	if err != nil {
		if must_trace_request {
			vulog.Infof("%s: FAILED (%s)", trace_req, err)
		}
		vulog.Errorf("HTTP request failed: %s", err)
		buildHTTPSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
		if resp != nil {
			ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
		}
		resultsChannel <- sampleReqResult
		return false
	}

	defer resp.Body.Close()
	sessionMap[config.HTTP_RESPONSE] = strconv.Itoa(resp.StatusCode)
	elapsed := time.Since(start)
	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		if must_trace_request {
			vulog.Infof("%s: FAILED (%s)", trace_req, err)
		}
		vulog.Printf("Reading HTTP response failed: %s", err)
		buildHTTPSampleResult(&sampleReqResult, 0, resp.StatusCode, elapsed.Nanoseconds(), req.URL.String())
		resultsChannel <- sampleReqResult
		return false
	}

	if must_trace_request {
		vulog.Infof("%s; RetCode=%d; RcvdBytes=%d", trace_req, resp.StatusCode, len(responseBody))
	}
	if must_display_srv_resp {
		vulog.Debugf("[HTTP Response=%d] Received data: %s", resp.StatusCode, responseBody)
	}

	valid := true

	// If the HTTP response code is listed in "http_error_codes" (404, 403, 500...),
	// the result is not processed and a false value is returned
	if strings.Contains(playbook.HttpErrorCodes, strconv.FormatInt(int64(resp.StatusCode), 10)) {
		vulog.Errorf("HTTP response code is considered as an error: %d", resp.StatusCode)
		valid = false
	}

	// if action specifies response action, parse using regexp/jsonpath
	if valid && !processResult(grpcAction.ResponseHandlers, sessionMap, vulog, responseBody, resp.Header) {
		valid = false
	}
	buildHTTPSampleResult(&sampleReqResult, len(responseBody), resp.StatusCode, elapsed.Nanoseconds(), req.Call)
	resultsChannel <- sampleReqResult
	 */
	//return valid
	return true
}

func buildGRPCRequest(grpcAction GRPCAction, sessionMap map[string]string, vulog *log.Entry) (*http.Request, error) {
	//var req *http.Request
	//var err error
	vulog.Debug("buildGRPCRequest")

	/*
	// Hack: the Path has been concatened with EscapedPath() (from net/url.go)
	// We must re-convert strings like $%7Bxyz%7D into ${xyz} to make variable substitution work !
	unescapedURL := RedecodeEscapedPath(httpAction.URL)

	if httpAction.Body != "" {
		// BODY
		reader := strings.NewReader(SubstParams(sessionMap, httpAction.Body, vulog))
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), reader)
	} else if httpAction.Template != "" {
		// TEMPLATE
		reader := strings.NewReader(SubstParams(sessionMap, httpAction.Template, vulog))
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), reader)
	} else if httpAction.Method == "PUT" && httpAction.UploadFile != nil {
		log.Debugf("prepare for uploading file content with PUT")
		reader := bytes.NewReader(httpAction.UploadFile)
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), reader)
	} else {
		// DEFAULT
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), nil)
	}

	if err != nil {
		err := fmt.Errorf("http.newRequest failed in buildHttpRequest: %s", err)
		return nil, err
	}
*/
	//return req, nil
	return nil, nil
}

func buildGRPCSampleResult(sample *reporter.SampleReqResult, contentLength int, status int, elapsed int64, fullreq string) {
	sample.Status = status
	sample.Size = contentLength
	sample.Latency = elapsed
	sample.FullRequest = fullreq
}
