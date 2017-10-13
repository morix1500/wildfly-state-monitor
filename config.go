package main

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

// Config -- config of wildfly-state-monitor
type Config struct {
	Slack   SlackConfig   `yaml:"slack"`
	Wildfly WildflyConfig `yaml:"wildfly"`
	App     AppConfig     `yaml:"app"`
}

// SlackConfig -- abount slack settings
type SlackConfig struct {
	APIURL  string `yaml:"api_url"`
	Channel string `yaml:"channel"`
}

// WildflyConfig -- about wildfly settings
type WildflyConfig struct {
	WarPath string `yaml:"war_path"`
}

// AppConfig -- about this script settings
type AppConfig struct {
	LogPath      string        `yaml:"log_path"`
	Duration     time.Duration `yaml:"duration"`
	NotifyMarker []string      `yaml:"notify_marker"`
}

var (
	// ErrReadConfig -- file not found error
	ErrReadConfig  = errors.New("not found config file")
	// ErrParseConfig -- file failed parse error
	ErrParseConfig = errors.New("parse error config file")
)

func validateConfig(config Config) error {
	if config.Slack.APIURL == "" {
		return errors.New("require config.slack.api_url")
	}
	if config.Slack.Channel == "" {
		return errors.New("require config.slack.channel")
	}
	if config.Wildfly.WarPath == "" {
		return errors.New("require config.wildfly.war_path")
	}
	return nil
}

func loadConfig(configPath string) (Config, error) {
	fp, err := ioutil.ReadFile(configPath)
	if err != nil {
		return (Config{}), errors.Wrap(ErrReadConfig, configPath)
	}

	var config Config
	err = yaml.Unmarshal(fp, &config)
	if err != nil {
		return (Config{}), errors.Wrap(ErrParseConfig, configPath)
	}
	err = validateConfig(config)

	return config, err
}
