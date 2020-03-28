package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"

	"function"
)

const (
	srvAddr     = ":8080"
	metricsAddr = ":8081"
	healthAddr  = ":8082"
)

// this example will show the simplest way of enabling the middleware
// using go std handlers without frameworks with the Prometheues recorder.
// it uses the prometheus default registry, the middleware default options
// and doesn't set the handler ID, so the middleware will set the handler id
// (handler label) to the request URL, WARNING: this creates high cardinality because
// `/profile/123` and `/profile/567` are different handlers. If you want to be safer
// you will need to tell the middleware what is the handler id.
func main() {
	// Create our middleware.
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	// Create our server.
	mux := function.NewHandler()

	// Wrap our main handler, we pass empty handler ID so the middleware inferes
	// the handler label from the URL.
	h := mdlw.Handler("", mux)

	// Serve our handler.
	go func() {
		if err := http.ListenAndServe(srvAddr, h); err != nil {
			log.Panicf("error while serving: %s", err)
		}
	}()

	// Serve our metrics.
	go func() {
		if err := http.ListenAndServe(metricsAddr, promhttp.Handler()); err != nil {
			log.Panicf("error while serving metrics: %s", err)
		}
	}()

	health := healthcheck.NewHandler()
	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

	// Serve our health.
	go func() {
		if err := http.ListenAndServe(healthAddr, health); err != nil {
			log.Panicf("error while serving health: %s", err)
		}
	}()

	// Wait until some signal is captured.
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
}
