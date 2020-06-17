package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shyimo/kubeobserver/pkg/config"
	"github.com/shyimo/kubeobserver/pkg/controller"
	"github.com/shyimo/kubeobserver/pkg/server"
)

func serve(ctx context.Context) (err error) {
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(server.HealthHandler))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port()),
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Msg(fmt.Sprintf("listen:%s\n", err))
		}
	}()

	log.Info().Msg(fmt.Sprintf("server is listen on port %d..", config.Port()))
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
	go controller.StartWatch(time.Now())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	if err := serve(ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}
