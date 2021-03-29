package web_proxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
)

type proxiedRequest struct {
	Host string
	Method string
	URL url.URL
	Body string
	// PostForm url.Values
	// MultipartForm *multipart.Form
	ContentType string
	PostForm url.Values
}

var (
	dump_in_progress sync.Mutex
	proxiedRequests  = make([]proxiedRequest, 0)
	domain *string
)

// Start the Web Proxy
func StartProxy(listen_addr *string, proxy_domain *string, ignored_suffixes *string) {
	go signalHandler()

	domain = proxy_domain
	remote_server := strings.ToLower(*proxy_domain)
	exclude_suffixes := strings.Split(*ignored_suffixes, ",")

	log.Infof("Starting Web Proxy on adress: %s", *listen_addr)
	log.Infof("Proxied Domain: %s", *proxy_domain)
	log.Infof("Ignored Suffixes: %s", *ignored_suffixes)
	proxy := goproxy.NewProxyHttpServer()
	//proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Debugf("Proxy request: %s", r.URL)

			// Apply filter
			if !strings.HasSuffix(strings.ToLower(r.Host), remote_server) {
				log.Debugf("Skipping... (not the remote server)")
				return r, nil
			}
			for _, suffix := range exclude_suffixes {
				if strings.HasSuffix(r.URL.Path, suffix) {
					log.Debugf("Skipping... (bad suffix)")
					return r, nil
				}
			}

			log.Infof("Going to proxy request: %s %s", r.Method, r.URL)

			// Copy the Request in case of Body analysis
			var c http.Request
			c.Method = r.Method
			c.Proto = r.Proto
			c.ProtoMajor = r.ProtoMajor
			c.ProtoMinor = r.ProtoMinor
			c.Host = r.Host
			c.RequestURI = r.RequestURI
			c.ContentLength = r.ContentLength
			c.URL = r.URL

			c.Header = http.Header{}
			for k, vv := range r.Header {
				for _, v := range vv {
					log.Debugf("Header: %s: %s", k, v)
					c.Header.Set(k, v)
				}
			}

			var b bytes.Buffer
			io.Copy(&b, r.Body)
			r.Body.Close()
			buf := b.Bytes()

			var b1 bytes.Buffer
			b1.Write(buf)
			r.Body = ioutil.NopCloser(&b1)

			c.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

			//var body string
			ct := r.Header.Get("Content-Type")
			postForm := url.Values{}
			if c.Method == "POST" {
				c.ParseMultipartForm(1024 * 1024)
				/*
				if ct == "application/x-www-form-urlencoded" {
					c.ParseForm() // or call ParseMultipartForm() si Content-Type: multipart/form-data
				} else if ct == "multipart/form-data" {
					c.ParseMultipartForm(1024 * 1024)
				}
				*/

				for k, vv := range c.PostForm {
					log.Debugf("PostForm(%s, %s)", k, vv)
					for _, v := range vv {
						postForm.Add(k, v)
					}
				}
				/*
				for k, vv := range c.Form {
					log.Debugf("Form(%s, %s)", k, vv)
					for _, v := range vv {
						postForm.Set(k, v)
					}
				}
				*/
			}

			request := proxiedRequest{
				Host: r.Host,
				Method: r.Method,
				URL: *r.URL,
				//Body: body,
				ContentType: ct,
				PostForm: postForm,
			}
			dump_in_progress.Lock()
			proxiedRequests = append(proxiedRequests, request)
			dump_in_progress.Unlock()

			return r, nil
		})

	log.Fatal(http.ListenAndServe(*listen_addr, proxy))
}

// Handle the SIGHUP and dump the Playbook draft
func signalHandler() {
	// signChan channel is used to transmit signal notifications.
	dumpChan := make(chan os.Signal, 1)
	// Catch and relay certain signal(s) to signChan channel.
	signal.Notify(dumpChan, os.Interrupt, syscall.SIGTERM)

	for {
		// Blocking until a signal is sent over signChan channel
		<-dumpChan
		fmt.Print("\nDo you want to (e)xit, (c)ontinue, (d)isplay the Playbook, (r)eset the Playbook ? [e/c/d/r] ")
		var answer string
		fmt.Scanf("%s", &answer)
		if answer == "e" {
			fmt.Println("bye bye...")
			os.Exit(0)
		} else if answer == "c" {
			continue
		} else if answer == "r" {
			dump_in_progress.Lock()
			proxiedRequests  = proxiedRequests[:0]
			dump_in_progress.Unlock()
			continue
		}

		log.Infoln("Dump the Playbook...")
		dump_in_progress.Lock()

		// Make the Dump !
		header := `iterations: 1
users: 1
warmup: 1
default:
  server: %s
  protocol: http
  method: GET
actions:
`
		fmt.Printf(header, *domain)
		for idx, request := range proxiedRequests {
			fmt.Println("  - http:")
			fmt.Printf("      title: Action %d\n", idx+1)
			fmt.Printf("      url: %s", request.URL.Path)
			if request.URL.RawQuery != "" {
				fmt.Printf("?%s", request.URL.RawQuery)
			}
			fmt.Println()

			if idx == 0 {
				fmt.Println("      store_cookie: __all__")
			}

			if request.Method == "GET" {
				continue
			}

			fmt.Printf("      method: %s\n", request.Method)
			if request.Method == "POST" {
				if request.ContentType == "application/x-www-form-urlencoded" {
					fmt.Printf("      body: %s\n", request.PostForm.Encode())
					//     body: name=${name}&age=${age}
				} else if strings.HasPrefix(request.ContentType, "multipart/form-data") {
					fmt.Println("      formdata:")
					for k, v := range request.PostForm {
						fmt.Printf("        - name: %s\n", k)
						fmt.Printf("          value: %v\n", v[0])
					}
					/*
					    formdata:
					      - name: name
					        value: ${a_variable} Doe
					      - name: fileToUpload
					        value: a_filename.txt
					        type: file		# mandatory for files
					      - name: submit
					 */
				}
			}
		}

		dump_in_progress.Unlock()
	}
}

