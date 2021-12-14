package config

import (
	"errors"

	"github.com/BurntSushi/toml"
)

var (
	ErrNoConf    = errors.New("no orders found in config")
	ErrWrongType = errors.New("the type must is not limit or stop_limit")
)

type Config struct {
	Password         string   `toml:"wallet_password"`
	PancakeRouterHex string   `toml:"pancakeswap_router_address"`
	EndpointURL      string   `toml:"endpoint_url"`
	ChainID          int64    `toml:"chain_id"`
	WalletPath       string   `toml:"wallet_path"`
	DBPath           string   `toml:"database_path"`
	TelegramAPIKey   string   `toml:"telegram_api_key"`
	Users            []string `toml:"telegram_users"`
}

// Reads from a file with the given path and returns the config
func Load(path string) (*Config, error) {
	conf := &Config{
		EndpointURL: "https://bsc-dataseed.binance.org/",
		WalletPath:  "./.wallets",
		Password:    "",
	}
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
