package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/horacehylee/fzd"
	"github.com/spf13/viper"
)

type config struct {
	Index struct {
		BasePath string
	}
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
	viper.AddConfigPath(filepath.Join("$HOME", ".fzd"))

	viper.SetDefault("index.basepath", "$HOME/.fzd/indexes")

	err := viper.ReadInConfig()
	if err != nil {
		var e *viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			return config{}, fmt.Errorf("could not read config: %w", err)
		}
	}
	var c config
	err = c.parse()
	if err != nil {
		return config{}, err
	}

	err = c.validate()
	if err != nil {
		return config{}, err
	}

	return c, nil

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

func (c *config) parse() error {
	err := viper.Unmarshal(c)
	if err != nil {
		return err
	}

	c.Index.BasePath = absPathify(c.Index.BasePath)
	for i := range c.Locations {
		c.Locations[i].Path = absPathify(c.Locations[i].Path)
	}
	return nil
}

func (c *config) validate() error {
	// TODO: check all mandatory fields
	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func absPathify(inPath string) string {
	inPath = filepath.Clean(inPath)
	if inPath == "$HOME" || strings.HasPrefix(inPath, "$HOME"+string(os.PathSeparator)) {
		inPath = userHomeDir() + inPath[5:]
	}
	inPath = os.ExpandEnv(inPath)

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}
	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}
	return ""
}
