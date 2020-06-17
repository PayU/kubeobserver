package server

import (
	"fmt"
	"net/http"
)

// HealthHandler is
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "healthy")
}
