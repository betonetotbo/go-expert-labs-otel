include Makefile.env
export

build:
	@cd deployments && docker compose build

rm:
	@cd deployments && docker compose rm

run-input:
	WEATHER_SERVICE_URL=http://localhost:3000 SERVER_PORT=3001 go run cmd/input/main.go

run-weather:
	SERVER_PORT=3001 go run cmd/weather/main.go

up:
	@cd deployments && docker compose up -d

down:
	@cd deployments && docker compose down

.PHONY: build rm run-input run-weather up down