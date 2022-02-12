package main

import (
	"errors"
	"fmt"

	"github.com/horacehylee/fzd"
	"github.com/spf13/viper"
)

type config struct {
	Locations []struct {
		Path    string
		Filters []fzd.Filter
		Ignores []interface{}
	}
}

func newConfig() (config, error) {
	viper.SetConfigName(".fzd")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	err := viper.ReadInConfig()
	if err != nil {
		var e *viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			return config{}, fmt.Errorf("could not read config: %w", err)
		}
	}
	return parseConfig()

	// TODO: help user write an example config
	// err = viper.SafeWriteConfig()
	// if err != nil {
	// 	var e viper.ConfigFileAlreadyExistsError
	// 	if !errors.As(err, &e) {
	// 		log.Fatal(fmt.Errorf("failed to write config: %w", err))
	// 	}
	// }
	// return config
}

func parseConfig() (config, error) {
	var config config
	err := viper.Unmarshal(&config)
	return config, err
}
