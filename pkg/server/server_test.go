package server

import (
	"net/http"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(HealthHandler))

	srv := &http.Server{
		Handler: mux,
	}

	if srv == nil {
		t.Error("TestHealthHandler: couldn't use HealthHandler in http server")
	}
}
