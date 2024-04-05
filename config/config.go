package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Issuer    string `yaml:"issuer"`
	SecretKey string `yaml:"secret_key"`
	MongoDB   struct {
		ConnectionString string `yaml:"connection_string"`
	} `yaml:"mongo_db"`
}

var config Config

func init() {
	readConfig, err := readConfigFile("config.yaml")
	fmt.Println("readConfig")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	config = readConfig
}

func readConfigFile(filename string) (Config, error) {
	var config Config

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func GetConfig() Config {
	return config
}
