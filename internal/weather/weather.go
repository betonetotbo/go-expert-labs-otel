package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/betonetotbo/go-expert-labs-otel/internal/entity"
	"github.com/betonetotbo/go-expert-labs-otel/internal/utils"
	"net/http"
	"net/url"
	"time"
)

type (
	Config struct {
		ServerPort    int    `mapstructure:"SERVER_PORT"`
		WeatherApiKey string `mapstructure:"WEATHER_API_KEY"`
	}
)

const (
	weatherApiUrl = "https://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=pt"
	viacepApiUrl  = "https://viacep.com.br/ws/%s/json/"
)

func NewWeatherQueryHandler(cfg *Config) http.HandlerFunc {
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

		q, err := callViacep(input.Cep)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		out, err := callWeatherApi(q, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		_ = json.NewEncoder(w).Encode(out)
	}
}

func callViacep(cep entity.CEP) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(viacepApiUrl, cep.GetDigits()), nil)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	data["pais"] = "Brasil"
	q := utils.ConcatFields(data, "logradouro", "bairro", "localidade", "uf", "pais")
	return q, nil
}

func callWeatherApi(q string, cfg *Config) (*entity.WeatherQueryOutputDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	u := fmt.Sprintf(weatherApiUrl, cfg.WeatherApiKey, url.QueryEscape(q))
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	current, ok := data["current"].(map[string]interface{})
	if !ok {
		return nil, errors.New("no current weather found")
	}

	var out entity.WeatherQueryOutputDTO

	j, _ := json.Marshal(current)
	// TODO calcular temp_k
	err = json.Unmarshal(j, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
