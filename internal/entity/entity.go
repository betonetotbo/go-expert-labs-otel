package entity

import (
	"regexp"
	"strings"
)

type (
	CEP string

	WeatherQueryInputDTO struct {
		Cep CEP `json:"cep"`
	}

	WeatherQueryOutputDTO struct {
		City  string  `json:"city"`
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
		TempK float64 `json:"temp_k"`
	}
)

func (c CEP) IsValid() bool {
	matches, _ := regexp.Match(`^\d{5}-?\d{3}$`, []byte(c))
	return matches
}

func (c CEP) GetDigits() string {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(string(c), -1)
	return strings.Join(matches, "")
}
