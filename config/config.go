package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	EthProxy string `yaml:"ethProxy"`
	Me       struct {
		Public  string `yaml:"public"`
		Private string `yaml:"private"`
		Elrond  string `yaml:"elrond"`
	} `json:"me"`
	Token   string `yaml:"token"`
	Genesis string `yaml:"genesis"`
}

func NewConfig(configPath string) (*Config, error) {
	configData, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	var cfg Config

	if err = yaml.Unmarshal(configData, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
