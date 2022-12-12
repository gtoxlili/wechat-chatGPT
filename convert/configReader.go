package convert

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	SessionToken string `json:"session-token"`
	CfClearance  string `json:"cf_clearance"`
}

func ReadConfig() *Config {
	var config Config
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if jsonParser.Decode(&config) != nil {
		log.Fatalln("解析配置失败 : ", err)
	}
	return &config
}

func SaveConfig(config *Config) {
	configFile, err := os.Create("config.json")
	if err != nil {
		log.Warnln("更新配置失败 : ", err)
	}
	defer configFile.Close()
	jsonEncoder := json.NewEncoder(configFile)
	if jsonEncoder.Encode(config) != nil {
		log.Warnln("解析配置失败 : ", err)
	}
}
