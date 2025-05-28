package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ICS   ICSConfig    `yaml:"ICS"`
	Email MailSettings `yaml:"Email"`
}

type ICSConfig struct {
	Calendars []CalendarConfig `yaml:"Calendars"`
}

type CalendarConfig struct {
	Name             string `yaml:"Name"`
	URL              string `yaml:"URL"`
	EXPOResourceName string `yaml:"EXPOResourceName"`
}

type MailSettings struct {
	SendEmails          bool          `yaml:"SendEmails"`
	MailContent         string        `yaml:"MailContent"`
	MailContentFallback string        `yaml:"MailContentFallback"`
	Mappings            []MailMapping `yaml:"Mappings"`
	FallbackEmail       MailAddress   `yaml:"FallbackEmail"`
	From                MailAddress   `yaml:"From"`
	Subject             string        `yaml:"Subject"`
}

type MailMapping struct {
	IcsSummary string `yaml:"icsSummary"`
	Address    string `yaml:"address"`
}

type MailAddress struct {
	Address string `yaml:"Address"`
	Name    string `yaml:"Name"`
}

func Load(filePath string) (*Config, error) {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err = yaml.Unmarshal(buf, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	// Check if we have any calendars to check
	if len(config.ICS.Calendars) == 0 {
		return nil, fmt.Errorf("no ICS configurations found in the config file")
	}
	if config.Email.FallbackEmail.Address == "" {
		return nil, fmt.Errorf("fallback email address is not set in the config file")
	}
	if config.Email.From.Address == "" {
		return nil, fmt.Errorf("from email address is not set in the config file")
	}
	return &config, nil
}
