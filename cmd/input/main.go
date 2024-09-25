package main

import (
	"github.com/betonetotbo/go-expert-labs-otel/internal/config"
	"github.com/betonetotbo/go-expert-labs-otel/internal/http_utils"
	"github.com/betonetotbo/go-expert-labs-otel/internal/input"
	"github.com/go-chi/chi/v5"
	"log"
	"net/url"
)

func main() {
	var cfg input.Config
	config.Load(&cfg)
	if cfg.ServerPort <= 0 {
		log.Fatalf("invalid server port %d", cfg.ServerPort)
	}
	if _, err := url.ParseRequestURI(cfg.WeatherServiceUrl); err != nil {
		log.Fatalf("invalid weather service url: %s", cfg.WeatherServiceUrl)
	}

	cfg.ServiceName = "input-service"

	http_utils.StartServer(&cfg.HttpConfig,
		func(router *chi.Mux, spanner http_utils.Spanner) {
			handler := input.NewHandler(&cfg, spanner)
			router.Post("/", handler.Handle)
		},
	)
}
