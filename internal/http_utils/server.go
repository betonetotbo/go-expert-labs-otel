package http_utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	HttpConfig struct {
		ServiceName               string
		ServerPort                int    `mapstructure:"SERVER_PORT" default:"8080"`
		OpenTelemetryCollectorUrl string `mapstructure:"OPEN_TELEMETRY_COLLECTOR_URL"`
	}
)

func StartServer(cfg *HttpConfig, registerRoutesCb func(*chi.Mux, Spanner)) {
	router, spanner := createRouter(cfg)
	defer spanner.Shutdown(context.Background())
	registerRoutesCb(router, spanner)

	svc := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on port %d...\n", cfg.ServerPort)
		err := svc.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("could not listen on port %d: %v", cfg.ServerPort, err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	log.Println("Waiting for shutdown signal...")

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := svc.Shutdown(ctx)
	if err != nil {
		log.Fatalf("could not shutdown server: %v", err)
	}
	log.Println("Server gracefully stopped")
}

func createRouter(cfg *HttpConfig) (*chi.Mux, Spanner) {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(time.Second * 15))

	var spanner Spanner
	if cfg.OpenTelemetryCollectorUrl != "" {
		o, err := NewOpenTelemetry(context.Background(), cfg.ServiceName, cfg.OpenTelemetryCollectorUrl)
		if err != nil {
			log.Fatal(err)
		}
		router.Use(o.Middleware)

		spanner = o
	} else {
		spanner = NewEmptySpanner()
	}

	router.Handle("/metrics", promhttp.Handler())

	return router, spanner
}
