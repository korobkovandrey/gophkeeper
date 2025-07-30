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
	v              *viper.Viper
	Addr           string   `mapstructure:"addr"`
	LogLevel       int8     `mapstructure:"log_level"`
	LogOutputs     []string `mapstructure:"log_outputs"`
	PrivateKeyPath string   `mapstructure:"private_key_path"`
	CAPath         string   `mapstructure:"ca_path"`
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
	conf.SetConfigFile("client_config.json")
	conf.AutomaticEnv()
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	pf := pflag.NewFlagSet("client", pflag.ExitOnError)

	conf.SetDefault("addr", defaultAddr)
	pf.String("addr", defaultAddr, "gRPS server address")
	_ = conf.BindPFlag("addr", pf.Lookup("addr"))

	conf.SetDefault("log_level", int8(defaultLogLevel))
	pf.Int8("log-level", int8(defaultLogLevel), "log Level")
	_ = conf.BindPFlag("log_level", pf.Lookup("log-level"))

	logOutputs := []string{defaultLogOutput}
	conf.SetDefault("log_outputs", logOutputs)
	pf.StringSlice("log-outputs", logOutputs, "comma separated log output paths")
	_ = conf.BindPFlag("log_outputs", pf.Lookup("log-outputs"))

	conf.SetDefault("private_key_path", "")
	pf.String("private-key", "", "private key path")
	_ = conf.BindPFlag("private_key_path", pf.Lookup("private-key"))

	conf.SetDefault("ca_path", "./certs/ca.crt")
	pf.String("ca", "./certs/ca.crt", "CA path")
	_ = conf.BindPFlag("ca_path", pf.Lookup("ca"))

	err := conf.ReadInConfig()
	if err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	_ = pf.Parse(os.Args[1:])
	cfg := &Config{
		v: conf,
	}

	err = conf.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	err = cfg.WriteConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}
	return cfg, nil
}

func (c *Config) SetPrivateKeyPath(privateKeyPath string) *Config {
	c.PrivateKeyPath = privateKeyPath
	c.v.Set("private_key_path", c.PrivateKeyPath)
	return c
}

func (c *Config) WriteConfig() error {
	return c.v.WriteConfig()
}
