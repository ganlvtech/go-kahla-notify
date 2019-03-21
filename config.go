package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
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
	config := new(Config)
	err := json.Unmarshal(data, config)
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
