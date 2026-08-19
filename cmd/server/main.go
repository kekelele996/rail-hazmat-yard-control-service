package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type summary struct {
	Yard    string `json:"yard"`
	Status  string `json:"status"`
	Modules int    `json:"modules"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}
	yard := os.Getenv("YARD_CODE")
	if yard == "" {
		yard = "NORTH-TRANSFER"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(summary{Yard: yard, Status: "operational", Modules: 10})
	})
	mux.Handle("/", http.FileServer(http.Dir("web")))
	addr := ":" + port
	log.Printf("rail hazmat yard service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(fmt.Errorf("serve yard control: %w", err))
	}
}
