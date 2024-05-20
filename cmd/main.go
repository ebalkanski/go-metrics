package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
)

const (
	serviceName = "metrics"
	port        = ":80"
)

func main() {
	log.Println("Starting service...")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// setup Prometheus
	promReporter := prometheus.NewReporter(prometheus.Options{})
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Tags:                   map[string]string{"service": serviceName},
		CachedReporter:         promReporter,
		Separator:              prometheus.DefaultSeparator,
		OmitCardinalityMetrics: true,
	}, 1*time.Second)
	defer closer.Close()

	// create basic request counter
	counter := scope.Counter("homepage_counter")

	r := http.NewServeMux()
	r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		counter.Inc(1)

		w.Write([]byte("Hey!\n"))
	}))
	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()

	// waiting for interrupt signal
	<-ctx.Done()

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Bye bye...")
}
