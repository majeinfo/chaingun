package designer

import (
	"embed"
	"net/http"
	log "github.com/sirupsen/logrus"
)

var designerUrl string = "/designer/"

//go:embed designer/*
var content embed.FS

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, designerUrl, 301)
}

// Start creates the HTTP server and creates the Web Interface for the Designer
func StartDesignerMode(listen_addr string) error {
	http.Handle("/designer/", http.FileServer(http.FS(content)))
	http.HandleFunc("/", redirect)
	if err := http.ListenAndServe(listen_addr, nil); err != nil {
		log.Fatalf("Could not listen on address %s: %v", listen_addr, err)
	}
	return nil
}
