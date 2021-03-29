package main

import (
	"embed"
	"net/http"
)

var designerUrl string = "/designer/"

//go:embed designer/*
var content embed.FS

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, designerUrl, 301)
}

// Start creates the HTTP server and creates the Web Interface for the Designer
func startDesignerMode(listen_addr *string) error {
	http.Handle("/designer/", http.FileServer(http.FS(content)))
	http.HandleFunc("/", redirect)
	http.ListenAndServe(*listen_addr, nil)
	return nil
}
