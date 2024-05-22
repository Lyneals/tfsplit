package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Layers []Layer `yaml:"layers"`
}

type Layer struct {
	Name     string `yaml:"name"`
	RootNode string `yaml:"rootNode"`
}

// Load a yaml file and return a Config struct
func LoadConfig(path string) (*Config, error) {
	c := Config{}
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal: %v", err)
	}

	return &c, nil
}
