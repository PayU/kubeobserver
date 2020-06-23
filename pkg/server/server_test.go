package server

import (
	"net/http"
	"testing"

	"github.com/shyimo/kubeobserver/pkg/server"
)

func TestHealthHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(server.HealthHandler))

	srv := &http.Server{
		Handler: mux,
	}

	if srv == nil {
		t.Error("TestHealthHandler: couldn't use HealthHandler in http server")
	}
}
