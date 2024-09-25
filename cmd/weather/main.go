package main

import (
	"github.com/betonetotbo/go-expert-labs-otel/internal/config"
	"github.com/betonetotbo/go-expert-labs-otel/internal/http_utils"
	"github.com/betonetotbo/go-expert-labs-otel/internal/weather"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"time"
)

func main() {
	var cfg weather.Config
	config.Load(&cfg)
	if cfg.ServerPort <= 0 {
		log.Fatalf("invalid server port %d", cfg.ServerPort)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(time.Second * 15))

	router.Post("/", weather.NewWeatherQueryHandler(&cfg))

	http_utils.StartServer(cfg.ServerPort, router)
}
