package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	v          *viper.Viper
	write      bool
	configPath string
	Addr       string   `mapstructure:"addr"`
	LogLevel   int8     `mapstructure:"log_level"`
	LogOutputs []string `mapstructure:"log_outputs"`
	DSN        string   `mapstructure:"dsn"`
	CertPath   string   `mapstructure:"cert_path"`
	KeyPath    string   `mapstructure:"key_path"`
}

const (
	defaultAddr       = "localhost:3200"
	defaultLogLevel   = zap.InfoLevel
	defaultLogOutput  = "stderr"
	defaultConfigFile = "config_server.json"
)

func NewConfig() (*Config, error) {
	v := viper.New()
	logOutputs := []string{defaultLogOutput}
	cfg := &Config{v: v, configPath: defaultConfigFile}

	pf := pflag.NewFlagSet("client", pflag.ExitOnError)
	pf.StringVarP(&cfg.configPath, "config", "c", cfg.configPath, "path to config file")
	pf.BoolVarP(&cfg.write, "write", "w", false, "write config to file")
	pf.String("addr", defaultAddr, "gRPS server host")
	pf.Int8("log-level", int8(defaultLogLevel), "log Level")
	pf.StringSlice("log-outputs", logOutputs, "comma separated log output paths")
	pf.String("dsn", "", "database dsn")
	pf.String("cert", "", "Cert path (use ./certs/server.crt)")
	pf.String("key", "", "Key path (use ./certs/server.key)")
	_ = pf.Parse(os.Args[1:])

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigFile(cfg.configPath)
	dir := filepath.Dir(cfg.configPath)
	if dir == "" {
		dir = "."
	}
	v.AddConfigPath(dir)

	_ = v.BindPFlag("addr", pf.Lookup("addr"))
	_ = v.BindPFlag("log_level", pf.Lookup("log-level"))
	_ = v.BindPFlag("log_outputs", pf.Lookup("log-output"))
	_ = v.BindPFlag("dsn", pf.Lookup("dsn"))
	_ = v.BindPFlag("cert_path", pf.Lookup("cert"))
	_ = v.BindPFlag("key_path", pf.Lookup("key"))

	v.SetDefault("addr", defaultAddr)
	v.SetDefault("log_level", int8(defaultLogLevel))
	v.SetDefault("log_outputs", logOutputs)
	v.SetDefault("dsn", "")
	v.SetDefault("cert_path", "")
	v.SetDefault("key_path", "")

	err := v.ReadInConfig()
	if err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	_ = pf.Parse(os.Args[1:])

	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func (c *Config) ConfigPath() string {
	return c.configPath
}

func (c *Config) IsWrite() bool {
	return c.write
}

func (c *Config) WriteConfig() error {
	return c.v.WriteConfig()
}
