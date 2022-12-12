package convert

import (
	"encoding/json"
	"os"
)

type Config struct {
	SessionToken string `json:"session-token"`
	CfClearance  string `json:"cf_clearance"`
	UserAgent    string `json:"user-agent"`
}

func ReadConfig() *Config {
	var config Config
	configFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		panic(err)
	}
	return &config
}

func SaveConfig(config *Config) {
	configFile, err := os.Create("config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()
	jsonEncoder := json.NewEncoder(configFile)
	if jsonEncoder.Encode(config) != nil {
		panic(err)
	}
}
