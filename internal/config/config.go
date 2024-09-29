package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Address    string `json:"address"`
	Port       uint16 `json:"port"`
	RootFolder string `json:"rootFolder"`
	AllowGifs  bool   `json:"allowGifs"`
	Signing    struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"signingKey"`
	Database struct {
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
	Logger struct {
		LogMode     string `json:"logMode"`
		LogRequests bool   `json:"logRequests"`
	} `json:"logger"`
	RateLimiter struct {
		Enable      bool     `json:"enable"`
		WlIPs       []string `json:"whiteListedIPs"`
		MaxRecConns uint     `json:"maxRecConns"`
		ExpirTime   uint     `json:"expirTimeMin"`
	} `json:"rateLimiter"`
}

type ConfigEmptyKey struct{}
type ConfigWeakKey struct{}

func (e ConfigEmptyKey) Error() string {
	return "Empty key value"
}

func (e ConfigWeakKey) Error() string {
	return "Key is too weak"
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

func (c *Config) ReadKey() ([]byte, error) {
	var (
		key []byte
	)
	switch c.Signing.Type {
	case "file":
		f, err := os.Open(c.Signing.Value)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		buf := new(bytes.Buffer)

		_, err = io.Copy(buf, f)
		if err != nil {
			return nil, err
		}

		key = buf.Bytes()
	case "value":
		key = []byte(c.Signing.Value)
		if len(key) < 8 {
			return nil, ConfigWeakKey{}
		}
	default:
		return nil, ConfigEmptyKey{}
	}

	return key, nil
}
