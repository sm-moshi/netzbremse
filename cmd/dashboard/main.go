package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sm-moshi/netzbremse/internal/config"
	"github.com/sm-moshi/netzbremse/internal/postgres"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	ctx := context.Background()

	dbConfig, err := config.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	dashboardConfig, err := config.LoadDashboard()
	if err != nil {
		log.Fatal(err)
	}

	store, err := postgres.New(ctx, dbConfig.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if err := store.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	}

	staticRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/api/overview", func(w http.ResponseWriter, r *http.Request) {
		overview, err := store.LoadOverview(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, overview)
	})
	mux.HandleFunc("/api/measurements", func(w http.ResponseWriter, r *http.Request) {
		limit := dashboardConfig.Limit
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err == nil && parsed > 0 && parsed <= 500 {
				limit = parsed
			}
		}
		items, err := store.ListLatest(r.Context(), limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, items)
	})
	mux.Handle("/", http.FileServer(http.FS(staticRoot)))

	server := &http.Server{
		Addr:              dashboardConfig.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("dashboard listening on %s", dashboardConfig.ListenAddress)
	log.Fatal(server.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}
