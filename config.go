package main

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Config struct {
	Slack   SlackConfig   `yaml:"slack"`
	Wildfly WildflyConfig `yaml:"wildfly"`
	App     AppConfig     `yaml:"app"`
}

type SlackConfig struct {
	ApiUrl  string `yaml:"api_url"`
	Channel string `yaml:"channel"`
}

type WildflyConfig struct {
	WarPath string `yaml:"war_path"`
}

type AppConfig struct {
	LogPath      string        `yaml:"log_path"`
	Duration     time.Duration `yaml:"duration"`
	NotifyMarker []string      `yaml:"notify_marker"`
}

var (
	ConfigReadErr  = errors.New("not found config file.")
	ConfigParseErr = errors.New("parse error config file.")
)

func validateConfig(config Config) error {
	if config.Slack.ApiUrl == "" {
		return errors.New("require config.slack.api_url.")
	}
	if config.Slack.Channel == "" {
		return errors.New("require config.slack.channel.")
	}
	if config.Wildfly.WarPath == "" {
		return errors.New("require config.wildfly.war_path.")
	}
	return nil
}

func loadConfig(config_path string) (Config, error) {
	fp, err := ioutil.ReadFile(config_path)
	if err != nil {
		return (Config{}), errors.Wrap(ConfigReadErr, config_path)
	}

	var config Config
	err = yaml.Unmarshal(fp, &config)
	if err != nil {
		return (Config{}), errors.Wrap(ConfigParseErr, config_path)
	}
	err = validateConfig(config)

	return config, err
}
