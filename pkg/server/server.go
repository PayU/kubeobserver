package server

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/shyimo/kubeobserver/pkg/controller"
)

type healthResponse struct {
	IsHealthy           bool `json:"is_healthy"`
	IsPodControllerSync bool `json:"is_pod_controller_sync"`
}

// HealthHandler is the handler function for GET /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Debug().Msg("got GET /health request")
	isHealthy := controller.IsSPodControllerSync()

	resBody := healthResponse{
		IsHealthy:           isHealthy,
		IsPodControllerSync: isHealthy,
	}

	jsResponse, _ := json.Marshal(resBody)

	w.Write(jsResponse)
}
