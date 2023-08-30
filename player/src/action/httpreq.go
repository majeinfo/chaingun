package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"golang.org/x/net/http2"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_HTTP string = "HTTP"
)

var (
	cookiePrefix       = config.COOKIE_PREFIX
	cookiePrefixLength = len(cookiePrefix)
)

// DoHTTPRequest accepts a Httpaction and a one-way channel to write the results to.
func DoHTTPRequest(httpAction HTTPAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, _ *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	var trace_req string
	sampleReqResult := buildSampleResult(REPORTER_HTTP, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, httpAction.Title, "")

	req, err := buildHTTPRequest(httpAction, sessionMap, vulog)
	if err != nil {
		vulog.Error(err)
		return false
	}
	if req.Method != "POST" {
		if must_trace_request {
			trace_req = fmt.Sprintf("%s %s", req.Method, req.URL)
		} else {
			vulog.Debugf("New Request: Method: %s, URL: %s", req.Method, req.URL)
		}
	} else {
		if must_trace_request {
			trace_req = fmt.Sprintf("%s %s; BODY(%s)", req.Method, req.URL, req.Body)
		} else {
			vulog.Debugf("New Request: Method: %s, URL: %s, Body: %s", req.Method, req.URL, req.Body)
		}
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
			//DisableCompression: true,
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
			//DisableCompression: true,
		}
	}

	resp, err := DefaultTransport.RoundTrip(req)
	vulog.Debugf("%v", resp)

	if err != nil {
		if must_trace_request {
			vulog.Infof("%s: FAILED (%s)", trace_req, err)
		}
		vulog.Errorf("HTTP request failed: %s", err)
		completeSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
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
		completeSampleResult(&sampleReqResult, 0, resp.StatusCode, elapsed.Nanoseconds(), req.URL.String())
		resultsChannel <- sampleReqResult
		return false
	}

	if must_trace_request {
		vulog.Infof("%s; RetCode=%d; RcvdBytes=%d", trace_req, resp.StatusCode, len(responseBody))
	}
	if must_display_srv_resp {
		vulog.Debugf("[HTTP Response=%d] Received data: %s", resp.StatusCode, responseBody)
	}
	store_srv_resp(httpAction.Title, sessionMap["UID"], vulog.Data["iter"].(int), responseBody)

	/*
		for _, cookie := range resp.Cookies() {
			vulog.Debugf("Cookie name: %s, cookie value: %s", cookie.Name, cookie.Value)
		}
		vulog.Debugf("StoreCookie=%s", httpAction.StoreCookie)
	*/

	if httpAction.StoreCookie != "" {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == httpAction.StoreCookie || httpAction.StoreCookie == "__all__" {
				vulog.Debugf("Store cookie: %s=%s", httpAction.StoreCookie, cookie.Value)
				sessionMap[cookiePrefix+cookie.Name] = cookie.Value
			}
		}
	}
	valid := true

	// If the HTTP response code is listed in "http_error_codes" (404, 403, 500...),
	// the result is not processed and a false value is returned
	if strings.Contains(playbook.HttpErrorCodes, strconv.FormatInt(int64(resp.StatusCode), 10)) {
		vulog.Errorf("HTTP response code is considered as an error: %d", resp.StatusCode)
		valid = false
	}

	// if action specifies response action, parse using regexp/jsonpath
	if valid && !processResult(httpAction.ResponseHandlers, sessionMap, vulog, responseBody, resp.Header) {
		valid = false
	}
	completeSampleResult(&sampleReqResult, len(responseBody), resp.StatusCode, elapsed.Nanoseconds(), req.URL.String())
	resultsChannel <- sampleReqResult
	return valid
}

func buildHTTPRequest(httpAction HTTPAction, sessionMap map[string]string, vulog *log.Entry) (*http.Request, error) {
	var req *http.Request
	var err error
	vulog.Debug("buildHttpRequest")

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
	} else if len(httpAction.FormDatas) > 0 {
		// FORM-DATA
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		for _, formdata := range httpAction.FormDatas {
			if formdata.Type != "file" {
				_ = writer.WriteField(formdata.Name, SubstParams(sessionMap, formdata.Value, vulog))
			} else {
				part, err := writer.CreateFormFile(formdata.Name, filepath.Base(formdata.Value))
				if err != nil {
					err := fmt.Errorf("Error while creating FormFile Part: %s", err)
					return nil, err
				}
				_, err = part.Write(formdata.Content)
			}
		}

		err = writer.Close()
		if err != nil {
			err := fmt.Errorf("Error while closing the FormData Writer: %s", err)
			return nil, err
		}

		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else if httpAction.Method == config.HTTP_PUT && httpAction.UploadFile != nil {
		log.Debugf("prepare for uploading file content with PUT")
		reader := bytes.NewReader(httpAction.UploadFile)
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), reader)
	} else {
		// DEFAULT
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), nil)
	}

	// Add the Basic Auth if required
	if err == nil && req.URL != nil {
		pwd, _ := req.URL.User.Password()
		req.SetBasicAuth(req.URL.User.Username(), pwd)
	}

	if err != nil {
		err := fmt.Errorf("http.newRequest failed in buildHttpRequest: %s", err)
		return nil, err
	}

	// Add headers
	for hdr, value := range httpAction.Headers {
		req.Header.Add(hdr, SubstParams(sessionMap, value, vulog))
	}
	if _, ok := httpAction.Headers["content-type"]; !ok && httpAction.Method == config.HTTP_POST {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// Add cookies stored by subsequent requests in the sessionMap having the kludgy __cookie__ prefix
	for key, value := range sessionMap {
		//vulog.Debugf("in sessionMap: key=%s", key)
		if strings.HasPrefix(key, cookiePrefix) {
			cookie := http.Cookie{
				Name:  key[cookiePrefixLength:len(key)],
				Value: value,
			}

			req.AddCookie(&cookie)
		}
	}

	return req, nil
}

