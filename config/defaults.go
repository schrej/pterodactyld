package config

import (
	"errors"
	"reflect"

	"github.com/spf13/viper"
)

func SetDefaults() error {
	return setDefaultsFromStruct("", reflect.ValueOf(Configuration{}))
}

func setDefaultsFromStruct(prefix string, s reflect.Value) error {
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		return errors.New("invalid kind")
	}

	for i := 0; i < s.NumField(); i++ {
		f := s.Type().Field(i)
		name := getStructFieldName(f)
		if f.Type.Kind() == reflect.Struct {
			if err := setDefaultsFromStruct(joinConfigKey(prefix, name), s.Field(i)); err != nil {
				return err
			}
			continue
		}
		key := joinConfigKey(prefix, name)
		if !viper.IsSet(key) {
			if def := f.Tag.Get("default"); def != "" {
				viper.Set(key, def)
			}
		}
	}
	return nil
}

func getStructFieldName(f reflect.StructField) string {
	name := f.Tag.Get("mapstructure")
	if name == "" {
		name = f.Name
	}
	return name
}

func joinConfigKey(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
