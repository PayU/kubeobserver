package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/PayU/kubeobserver/pkg/controller"
	"github.com/PayU/kubeobserver/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func serve(ctx context.Context) (err error) {
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(server.HealthHandler))
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port()),
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msg(fmt.Sprintf("listen:%s\n", err))
		}
	}()

	log.Info().Msg(fmt.Sprintf("server is listening on port %d..", config.Port()))
	<-ctx.Done()
	log.Info().Msg("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Error().Msg(fmt.Sprintf("server Shutdown Failed:%s", err))
	}

	log.Info().Msg("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

func main() {
	zerolog.SetGlobalLevel(config.LogLevel())

	// start k8s controller watchers
	go controller.StartWatch(time.Now())

	// create a channel for listening to OS signals
	// and connecting OS interrupts to the channel.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// create a context with cancel() callback function
	ctx, cancel := context.WithCancel(context.Background())

	// the cancelling of context happens after an OS interrupt.
	go func() {
		oscall := <-c
		log.Info().Msg(fmt.Sprintf("system call:%s", oscall))
		cancel()
	}()

	// start the http server
	if err := serve(ctx); err != nil {
		log.Error().Msg(fmt.Sprintf("failed to serve:%s\n", err))
	}
}
