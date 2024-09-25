package input

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/betonetotbo/go-expert-labs-otel/internal/entity"
	"github.com/betonetotbo/go-expert-labs-otel/internal/http_utils"
	"net/http"
	"time"
)

type (
	Config struct {
		http_utils.HttpConfig `mapstructure:",squash"`
		WeatherServiceUrl     string `mapstructure:"WEATHER_SERVICE_URL"`
	}

	Handler struct {
		cfg     *Config
		spanner http_utils.Spanner
	}
)

func NewHandler(cfg *Config, spanner http_utils.Spanner) *Handler {
	return &Handler{cfg: cfg, spanner: spanner}
}

func (i *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var input entity.WeatherQueryInputDTO
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal payload: %v", err), http.StatusBadRequest)
		return
	}
	if !input.Cep.IsValid() {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	out, err := i.callQueryWeather(r.Context(), &input)
	if err != nil {
		http.Error(w, fmt.Sprintf("call to weather microservice failed: %v", err), http_utils.GetStatusCode(err))
		return
	}

	_ = json.NewEncoder(w).Encode(out)
}

func (i *Handler) callQueryWeather(ctx context.Context, input *entity.WeatherQueryInputDTO) (*entity.WeatherQueryOutputDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	ctx, span := i.spanner.Start(ctx, "call weather microservice")
	r, _ := http_utils.NewRequest(ctx, http.MethodPost, i.cfg.WeatherServiceUrl, bytes.NewBuffer(data))
	resp, err := http.DefaultClient.Do(r)
	span.End()

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, http_utils.NewHttpError(resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	var out entity.WeatherQueryOutputDTO
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
