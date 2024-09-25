package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/betonetotbo/go-expert-labs-otel/internal/entity"
	"github.com/betonetotbo/go-expert-labs-otel/internal/http_utils"
	"github.com/betonetotbo/go-expert-labs-otel/internal/utils"
	"net/http"
	"net/url"
	"time"
)

type (
	Config struct {
		http_utils.HttpConfig `mapstructure:",squash"`
		WeatherApiKey         string `mapstructure:"WEATHER_API_KEY"`
	}

	Handler struct {
		cfg     *Config
		spanner http_utils.Spanner
	}
)

const (
	weatherApiUrl = "https://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=pt"
	viacepApiUrl  = "https://viacep.com.br/ws/%s/json/"
)

func NewHandler(cfg *Config, spanner http_utils.Spanner) *Handler {
	return &Handler{cfg: cfg, spanner: spanner}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var input entity.WeatherQueryInputDTO
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal payload: %v", err), http.StatusBadRequest)
		return
	}

	q, err := h.callViacep(r.Context(), input.Cep)
	if err != nil {
		http.Error(w, fmt.Sprintf("call to viacep.com.br failed: %v", err), http_utils.GetStatusCode(err))
		return
	}

	out, err := h.callWeatherApi(r.Context(), q)
	if err != nil {
		http.Error(w, fmt.Sprintf("call to weatherapi.com faield: %v", err), http_utils.GetStatusCode(err))
		return
	}

	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handler) callViacep(ctx context.Context, cep entity.CEP) (string, error) {
	if !cep.IsValid() {
		return "", http_utils.NewHttpError(http.StatusUnprocessableEntity, "invalid zipcode")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	r, _ := http_utils.NewRequest(ctx, http.MethodGet, fmt.Sprintf(viacepApiUrl, cep.GetDigits()), nil)
	resp, err := http_utils.DoRequest(h.spanner, "call viacep.com.br", r)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	var erro bool
	if ok := utils.ConvertField(data, "erro", &erro); ok && erro {
		return "", http_utils.NewHttpError(http.StatusNotFound, "can not find zipcode")
	}

	data["pais"] = "Brasil"
	q := utils.ConcatFields(data, "logradouro", "bairro", "localidade", "uf", "pais")
	return q, nil
}

func (h *Handler) callWeatherApi(ctx context.Context, q string) (*entity.WeatherQueryOutputDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	u := fmt.Sprintf(weatherApiUrl, h.cfg.WeatherApiKey, url.QueryEscape(q))
	r, _ := http_utils.NewRequest(ctx, http.MethodGet, u, nil)
	resp, err := http_utils.DoRequest(h.spanner, "call weatherapi.com", r)
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
		return nil, http_utils.NewHttpError(http.StatusNotFound, "can not find zipcode")
	}

	var out entity.WeatherQueryOutputDTO

	j, _ := json.Marshal(current)
	err = json.Unmarshal(j, &out)
	if err != nil {
		return nil, err
	}

	out.City = q
	out.TempK = out.TempC + 273.0

	return &out, nil
}
