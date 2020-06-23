package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"testing"
	"time"

	"github.com/PayU/kubeobserver/pkg/controller"
)

func TestHealthHandler(t *testing.T) {
	// start k8s controller watchers
	go controller.StartWatch(time.Now())

	// create a channel for listening to OS signals
	// and connecting OS interrupts to the channel.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	recorder := httptest.NewRecorder()
	r := http.Request{}

	HealthHandler(recorder, &r)
	response := recorder.Result()
	body, _ := ioutil.ReadAll(response.Body)

	if !strings.Contains(string(body), "{\"is_healthy\":false,\"is_pod_controller_sync\":false}") {
		t.Error("TestHealthHandler: Health handler hasn't responded correctly")
	}
}
