package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Phase 0 scaffold: minimal health endpoint so the service builds, runs,
// and can be curled independently. Real notifications logic (routing, JWT, rate
// limiting) comes later — see docs/02-services.md.
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "notifications",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("notifications listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
