package main

import (
	"github.com/betonetotbo/go-expert-labs-otel/internal/config"
	"github.com/betonetotbo/go-expert-labs-otel/internal/http_utils"
	"github.com/betonetotbo/go-expert-labs-otel/internal/input"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/url"
	"time"
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(time.Second * 15))

	router.Post("/", input.NewInputHandler(&cfg))

	http_utils.StartServer(cfg.ServerPort, router)
}
