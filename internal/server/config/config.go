package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Addr       string   `mapstructure:"addr"`
	LogLevel   int8     `mapstructure:"log_level"`
	LogOutputs []string `mapstructure:"log_outputs"`
	DSN        string   `mapstructure:"dsn"`
	CertPath   string   `mapstructure:"cert_path"`
	KeyPath    string   `mapstructure:"key_path"`
}

const (
	defaultAddr      = "localhost:3200"
	defaultLogLevel  = zap.InfoLevel
	defaultLogOutput = "stderr"
)

func NewConfig() (*Config, error) {
	conf := viper.New()
	conf.AddConfigPath(".")
	conf.SetConfigType("json")
	conf.SetConfigFile("server_config.json")
	conf.AutomaticEnv()
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	pf := pflag.NewFlagSet("client", pflag.ExitOnError)
	conf.SetDefault("addr", defaultAddr)
	pf.String("addr", defaultAddr, "gRPS server host")
	_ = conf.BindPFlag("addr", pf.Lookup("addr"))

	conf.SetDefault("log_level", int8(defaultLogLevel))
	pf.Int8("log-level", int8(defaultLogLevel), "log Level")
	_ = conf.BindPFlag("log_level", pf.Lookup("log-level"))

	logOutputs := []string{defaultLogOutput}
	conf.SetDefault("log_outputs", logOutputs)
	pf.StringSlice("log-outputs", logOutputs, "comma separated log output paths")
	_ = conf.BindPFlag("log_outputs", pf.Lookup("log-output"))

	conf.SetDefault("dsn", "")
	pf.String("dsn", "", "database dsn")
	_ = conf.BindPFlag("dsn", pf.Lookup("dsn"))

	conf.SetDefault("cert_path", "./certs/server.crt")
	pf.String("cert", "./certs/server.crt", "Cert path")
	_ = conf.BindPFlag("cert_path", pf.Lookup("cert"))

	conf.SetDefault("key_path", "./certs/server.key")
	pf.String("key", "./certs/server.key", "Key path")
	_ = conf.BindPFlag("key_path", pf.Lookup("key"))

	err := conf.ReadInConfig()
	if err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	_ = pf.Parse(os.Args[1:])
	cfg := &Config{}

	err = conf.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	err = conf.WriteConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}
	return cfg, nil
}
