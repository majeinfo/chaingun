package web_proxy

import (
	"fmt"
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
}

var (
	dump_in_progress sync.Mutex
	proxiedRequests  = make([]proxiedRequest, 0)
)

// Start the Web Proxy
func StartProxy(listen_addr *string) {
	go signalHandler()

	remote_server := strings.ToLower("www.delamarche.com")
	exclude_suffixes := []string{
		".gif",	".png", ".jpg", ".jpeg", ".css", ".js", ".ico", ".ttf", ".woff", ".pdf",
	}

	log.Infof("Starting Web Proxy on adress: %s", *listen_addr)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Debugf("Proxy request: %s", r.URL)

			// Apply filter
			if strings.ToLower(r.Host) != remote_server {
				log.Debugf("Skipping... (not the remote server)")
				return r, nil
			}
			for _, suffix := range exclude_suffixes {
				if strings.HasSuffix(r.URL.Path, suffix) {
					log.Debugf("Skipping... (bad suffix)")
					return r, nil
				}
			}

			log.Infof("Going to proxy request: %s", r.URL)
			request := proxiedRequest{
				Host: r.Host,
				Method: r.Method,
				URL: *r.URL,
				//Body: r.Body,
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
		fmt.Println("Do you want to exit or create a Playbook ? [e/p] ")
		var answer string
		fmt.Scanf("%s", &answer)
		if answer == "e" {
			fmt.Println("bye bye...")
			os.Exit(0)
		}

		log.Infoln("Dump the Playbook...")
		dump_in_progress.Lock()

		// Make the Dump !
		header := `iterations: 1
users: 1
warmup: 1
default:
  method: GET`

		fmt.Println(header)
		for _, request := range proxiedRequests {
			fmt.Printf("%s %s\n", request.Host, request.URL)
		}

		dump_in_progress.Unlock()
	}
}