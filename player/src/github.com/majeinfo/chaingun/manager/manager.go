package manager

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Start creates the HTTP server and creates the Web Interface
func Start(mgr_addr *string) error {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(*mgr_addr, nil))
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
