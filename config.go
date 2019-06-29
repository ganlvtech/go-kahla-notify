package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Config struct {
	Email            string `json:"Email"`
	Password         string `json:"Password"`
	ServerUrl        string `json:"ServerUrl"`
	OssUrl           string `json:"OssUrl"`
	EnableSnoreToast bool   `json:"EnableSnoreToast"`
	SnoreToastPath   string `json:"SnoreToastPath"`
	AvatarsDir       string `json:"AvatarsDir"`
}

func SaveConfig(config *Config) ([]byte, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func SaveConfigToFile(filename string, config *Config) error {
	data, err := SaveConfig(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

func LoadConfig(data []byte) (*Config, error) {
	config := &Config{}
	err := json.Unmarshal(data, config)
	if config.Email == "" {
		return config, errors.New("empty email")
	}
	if config.Password == "" {
		return config, errors.New("empty password")
	}
	if config.SnoreToastPath == "" {
		config.SnoreToastPath = "SnoreToast.exe"
	}
	if config.ServerUrl == "" {
		config.ServerUrl = "https://server.kahla.app"
	}
	if config.OssUrl == "" {
		config.OssUrl = "https://oss.aiursoft.com"
	}
	if config.AvatarsDir == "" {
		config.AvatarsDir = "avatars"
	}
	if err != nil {
		return config, err
	}
	return config, nil
}

func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config, err := LoadConfig(data)
	if err != nil {
		return config, err
	}
	return config, nil
}
