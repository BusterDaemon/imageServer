package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Address     string   `json:"address"`
	Port        uint16   `json:"port"`
	RootFolder  []string `json:"rootFolder"`
	AllowGifs   bool     `json:"allowGifs"`
	DBPath      string   `json:"dbPath"`
	Compression struct {
		UseCompression bool  `json:"useCompression"`
		CompressionLvl uint8 `json:"compLevel"`
	} `json:"compression"`
	Cache struct {
		UseCache      bool     `json:"useCache"`
		ExpCache      uint     `json:"expirCache"`
		WhitelistResp []string `json:"whitelistTypes"`
	} `json:"caching"`
	Auth struct {
		Enable   bool   `json:"enable"`
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"auth"`
	Logger struct {
		LogMode     string `json:"logMode"`
		LogRequests bool   `json:"logRequests"`
	} `json:"logger"`
}

func ReadConfig(configPath string) (Config, error) {
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0765)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var conf Config

	decoder := json.NewDecoder(f)
	decoder.Decode(&conf)
	if conf.Cache.ExpCache > 3600 {
		conf.Cache.ExpCache = 30
	}
	return conf, nil
}
