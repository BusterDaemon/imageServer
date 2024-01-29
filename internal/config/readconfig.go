package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Address    string   `json:"address"`
	Port       uint16   `json:"port"`
	RootFolder []string `json:"rootFolder"`
	AllowGifs  bool     `json:"allowGifs"`
	Database   struct {
		DbType     string `json:"dbType"`
		DbAddress  string `json:"dbAddress"`
		DbPort     uint   `json:"dbPort"`
		DbSSL      string `json:"dbSSL"`
		DbLogin    string `json:"dbLogin"`
		DbPassword string `json:"dbPassword"`
		DbName     string `json:"dbName"`
	} `json:"database"`
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

func New() Config {
	return Config{}
}

func (c *Config) ReadConfig(configPath string) error {
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0765)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.Decode(c)
	if c.Cache.ExpCache > 3600 {
		c.Cache.ExpCache = 30
	}
	return nil
}
