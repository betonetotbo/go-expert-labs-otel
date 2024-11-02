package main

import (
	"github.com/betonetotbo/go-expert-labs-otel/internal/config"
	"github.com/betonetotbo/go-expert-labs-otel/internal/http_utils"
	"github.com/betonetotbo/go-expert-labs-otel/internal/weather"
	"github.com/go-chi/chi/v5"
	"log"
)

func main() {
	var cfg weather.Config
	config.Load(&cfg)
	if cfg.ServerPort <= 0 {
		log.Fatalf("invalid server port %d", cfg.ServerPort)
	}

	cfg.ServiceName = "weather-service"

	if cfg.WeatherApiKey == "" {
		log.Fatalf("missing weather api key")
	}

	http_utils.StartServer(&cfg.HttpConfig,
		func(router *chi.Mux, spanner http_utils.Spanner) {
			handler := weather.NewHandler(&cfg, spanner)
			router.Post("/", handler.Handle)
		},
	)
}
