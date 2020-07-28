package main

import (
	"net/http"

	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"
	_ "statik"
)

var designerUrl string = "/designer/"

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, designerUrl, 301)
}

// Start creates the HTTP server and creates the Web Interface for the Designer
func startDesignerMode(listen_addr *string) error {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	//handler := new(Handler)
	http.Handle("/designer/", http.FileServer(statikFS))
	http.HandleFunc("/", redirect)
	http.ListenAndServe(*listen_addr, nil)
	return nil
}
