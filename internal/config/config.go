package config

import (
	"encoding/json"
	"github.com/spf13/viper"
	"log"
	"reflect"
	"strings"
)

func Load(cfg any) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	initProperties(v, cfg)

	err := v.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("unable to decode config: %v", err)
	}
}

func getFields(v interface{}) map[string]any {
	val := reflect.ValueOf(v)

	// Se a struct é um ponteiro, obtemos o valor apontado
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fields := make(map[string]any)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)

		if field.Type.Kind() == reflect.Struct {
			v := val.FieldByName(field.Name).Interface()
			subFields := getFields(v)
			for k, v := range subFields {
				fields[k] = v
			}
		} else {
			env := field.Tag.Get("mapstructure")
			if env != "" {
				defaultValueStr := field.Tag.Get("default")
				var defaultValue any
				if defaultValueStr != "" {
					_ = json.Unmarshal([]byte(defaultValueStr), &defaultValue)
				}
				fields[env] = defaultValue
			}
		}
	}

	return fields
}

func initProperties(v *viper.Viper, cfg any) {
	fields := getFields(cfg)

	for env, defValue := range fields {
		_ = v.BindEnv(env)
		v.SetDefault(env, defValue)
	}
}
