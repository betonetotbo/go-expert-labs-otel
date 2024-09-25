package input

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/betonetotbo/go-expert-labs-otel/internal/entity"
	"net/http"
	"time"
)

type (
	Config struct {
		ServerPort        int    `mapstructure:"SERVER_PORT"`
		WeatherServiceUrl string `mapstructure:"WEATHER_SERVICE_URL"`
	}
)

func NewInputHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var input entity.WeatherQueryInputDTO
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !input.Cep.IsValid() {
			http.Error(w, "invalid cep", http.StatusBadRequest)
			return
		}

		out, err := callQueryWeather(&input, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(out)
	}
}

func callQueryWeather(input *entity.WeatherQueryInputDTO, cfg *Config) (*entity.WeatherQueryOutputDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WeatherServiceUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var out entity.WeatherQueryOutputDTO
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
