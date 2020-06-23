package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"testing"
	"time"

	"github.com/PayU/kubeobserver/pkg/controller"
	"github.com/rs/zerolog/log"
)

func TestHealthHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	r := http.Request{}

	go controller.StartWatch(time.Now())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	HealthHandler(recorder, &r)
	_, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Info().Msg(fmt.Sprintf("system call:%s", oscall))
		cancel()
	}()

	response := recorder.Result()
	body, err := ioutil.ReadAll(response.Body)

	if !strings.Contains(string(body), "{\"is_healthy\":false,\"is_pod_controller_sync\":false}") || err != nil {
		t.Error("TestHealthHandler: Health handler hasn't responded correctly")
	}
}
