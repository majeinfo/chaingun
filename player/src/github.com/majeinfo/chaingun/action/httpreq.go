package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

var (
	cookiePrefix       = "__cookie__"
	cookiePrefixLength = len(cookiePrefix)
)

// DoHTTPRequest accepts a Httpaction and a one-way channel to write the results to.
func DoHTTPRequest(httpAction HTTPAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	req, err := buildHTTPRequest(httpAction, sessionMap, vulog)
	if err != nil {
		vulog.Error(err)
		return false
	}
	if req.Method != "POST" {
		vulog.Debugf("New Request: Method: %s, URL: %s", req.Method, req.URL)
	} else {
		vulog.Debugf("New Request: Method: %s, URL: %s, Body: %v", req.Method, req.URL, req.Body)
	}

	start := time.Now()
	var DefaultTransport http.RoundTripper = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		ResponseHeaderTimeout: time.Duration(playbook.Timeout) * time.Second,
	}
	resp, err := DefaultTransport.RoundTrip(req)

	if err != nil {
		vulog.Errorf("HTTP request failed: %s", err)
		return false
	}
	sessionMap[config.HTTP_RESPONSE] = strconv.Itoa(resp.StatusCode)

	elapsed := time.Since(start)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		vulog.Printf("Reading HTTP response failed: %s", err)
		sampleReqResult := buildSampleResult("HTTP", sessionMap["UID"], 0, resp.StatusCode, elapsed.Nanoseconds(), httpAction.Title, req.URL.String())
		resultsChannel <- sampleReqResult
		return false
	}

	vulog.Debugf("[HTTP Response=%d] Received data: %s", resp.StatusCode, responseBody)
	defer resp.Body.Close()

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
	sampleReqResult := buildSampleResult("HTTP", sessionMap["UID"], len(responseBody), resp.StatusCode, elapsed.Nanoseconds(), httpAction.Title, req.URL.String())
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
				// TODO: should apply variable interpolation
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
	} else {
		// DEFAULT
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, unescapedURL, vulog), nil)
	}
	if err != nil {
		err := fmt.Errorf("http.newRequest failed in buildHttpRequest: %s", err)
		return nil, err
	}

	// Add headers
	for hdr, value := range httpAction.Headers {
		req.Header.Add(hdr, SubstParams(sessionMap, value, vulog))
	}
	if _, ok := httpAction.Headers["content-type"]; !ok && httpAction.Method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// Add cookies stored by subsequent requests in the sessionMap having the kludgy __cookie__ prefix
	for key, value := range sessionMap {
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
