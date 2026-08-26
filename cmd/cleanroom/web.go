package main

import (
	"net/http"
	"os"
	"path/filepath"
)

func (s *Server) serveConsole(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.webDir, "console.html")
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "console unavailable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
