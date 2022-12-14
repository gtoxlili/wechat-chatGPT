package config

import (
	"bytes"
	"encoding/json"
	m "github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
	"wxChatGPT/util/signature"
)

var (
	config = &Config{}
	// 在每次更新 config.json 时，需要执行的事件
	configChangeCallbacks = make([]func(), 0)
	fileHash              = make([]byte, 0)
)

type Config struct {
	SessionToken string `json:"session-token"`
	CfClearance  string `json:"cf_clearance"`
	UserAgent    string `json:"user-agent"`
	Debug        bool   `json:"debug"`
	LogLevel     string `json:"log-level"`
}

func init() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Fatalln("初始化失败: ", err)
		}
	}()
	readConfig()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorln("监听 config.json 文件失败: ", err)
				if GetIsDebug() {
					m.PrintPrettyStack(err)
				}
			}
		}()
		// 每 30 秒同步一次 config.json
		for range time.Tick(30 * time.Second) {
			readConfig()
		}
	}()
}

func readConfig() {
	configFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	// 获取当前 config.json 的 hash 值
	hash, err := signature.GetFileHash(configFile)
	if err != nil {
		panic(err)
	}
	// 如果 hash 值与上一次的 hash 值相同，则不需要更新
	if bytes.Equal(hash, fileHash) {
		return
	} else {
		fileHash = hash
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(config); err != nil {
		panic(err)
	}
	for _, callback := range configChangeCallbacks {
		callback()
	}
}

func ReadConfig() *Config {
	return config
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

func GetIsDebug() bool {
	return config.Debug
}

func GetLogLevel() logrus.Level {
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		return logrus.DebugLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func AddConfigChangeCallback(callback func()) {
	configChangeCallbacks = append(configChangeCallbacks, callback)
}
