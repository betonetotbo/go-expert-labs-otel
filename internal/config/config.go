package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"strings"
)

func Load(cfg any) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	var md mapstructure.Metadata
	e := mapstructure.DecodeMetadata(map[string]any{}, cfg, &md)
	if e != nil {
		log.Fatal(e)
	}
	for _, k := range md.Unset {
		v.SetDefault(k, nil)
	}

	err := v.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("unable to decode config: %v", err)
	}
}
