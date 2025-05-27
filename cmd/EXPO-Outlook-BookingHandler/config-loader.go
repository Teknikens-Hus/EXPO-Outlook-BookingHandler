package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigSettings struct {
	ICSConfig    ICSConfig     `yaml:"ICS"`
	MailSettings MailSettings  `yaml:"Email"`
	Resources    []ResourceMap `yaml:"ResourceMap"`
}

type CalendarConfig struct {
	Name string `yaml:"Name"`
	URL  string `yaml:"URL"`
}

type ICSConfig struct {
	Calendars []CalendarConfig `yaml:"Calendars"`
}

type MailMapping struct {
	IcsSummary string `yaml:"icsSummary"`
	Address    string `yaml:"address"`
}

type MailSettings struct {
	SendEmails    bool          `yaml:"SendEmails"`
	Mappings      []MailMapping `yaml:"Mappings"`
	FallbackEmail MailAddress   `yaml:"FallbackEmail"`
	From          MailAddress   `yaml:"From"`
	Subject       string        `yaml:"Subject"`
}

type MailAddress struct {
	Address string `yaml:"Address"`
	Name    string `yaml:"Name"`
}

type ResourceMap struct {
	Name             string `yaml:"Name"`
	LinkedCalName    string `yaml:"LinkedCalName"`
	EXPOResourceName string `yaml:"EXPOResourceName"`
}

func LoadConfigFile() (*ConfigSettings, error) {
	config := &ConfigSettings{}
	buf, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	return config, nil
}
