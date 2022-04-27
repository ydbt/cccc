package config_test

import (
	"cccc/pkg/config"
	"context"
	"io/ioutil"
	"os"
	"testing"
)

type HttpConfig struct {
	bRunNormalize bool
	Port          int    `cccc:"port"`
	Address       string `cccc:"address"`
}
type RedisConfig struct {
	Address []string `cccc:"address"`
}
type UserConfig struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
type LogConfig struct {
	Path   string       `cccc:"path"`
	File   string       `cccc:"file"`
	Suffix string       `cccc:"suffix"`
	Users  []UserConfig `cccc:"users"`
}

func (h *HttpConfig) Normalize() {
	h.bRunNormalize = true
}
func (h *HttpConfig) RedepInit(context.Context) {
}

type ServerConfig struct {
	Http  HttpConfig  `cccc:"http"`
	Redis RedisConfig `cccc:"redis"`
	Log   LogConfig   `cccc:"log"`
}

func TestPartConfig(t *testing.T) {
	h := &HttpConfig{}
	ctx := context.Background()
	cfg := config.NewConfigServer(ctx)
	cfg.Regist("http", h)
	var err error
	defer func() {
		if err == nil {
			os.Remove("./demo_test.yaml")
		}
	}()
	err = ioutil.WriteFile("./demo_test.yaml", []byte(`http: 
    address: "127.0.0.1"
    port: 8080
redis: 
    address: ["10.130.29.10:6379","10.130.29.20:6379"]`), os.FileMode(os.O_CREATE|os.O_TRUNC))
	err = cfg.LoadConfigFile("./demo_test.yaml")
	if h.Address != "127.0.0.1" {
		t.Fatal("parse yaml failed")
	}
}
