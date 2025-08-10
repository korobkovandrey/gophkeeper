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
	v              *viper.Viper
	write          bool
	configPath     string
	Addr           string   `mapstructure:"addr"`
	LogLevel       int8     `mapstructure:"log_level"`
	LogOutputs     []string `mapstructure:"log_outputs"`
	PrivateKeyPath string   `mapstructure:"private_key_path"`
	CAPath         string   `mapstructure:"ca_path"`
}

const (
	defaultAddr       = "localhost:3200"
	defaultLogLevel   = zap.InfoLevel
	defaultLogOutput  = "client.log"
	defaultConfigFile = "config_client.json"
)

func NewConfig() (*Config, error) {
	v := viper.New()
	cfg := &Config{v: v, configPath: defaultConfigFile}

	pf := pflag.NewFlagSet("client", pflag.ExitOnError)
	pf.StringVarP(&cfg.configPath, "config", "c", cfg.configPath, "path to config file")
	pf.BoolVarP(&cfg.write, "write", "w", false, "write config to file")
	pf.String("addr", defaultAddr, "gRPC server address")
	pf.Int8("log-level", int8(defaultLogLevel), "log level")
	pf.StringSlice("log-outputs", []string{defaultLogOutput}, "comma separated log output paths")
	pf.String("private-key", "", "private key path")
	pf.String("ca", "", "CA path (use ./certs/ca.crt)")
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
	_ = v.BindPFlag("log_outputs", pf.Lookup("log-outputs"))
	_ = v.BindPFlag("private_key_path", pf.Lookup("private-key"))
	_ = v.BindPFlag("ca_path", pf.Lookup("ca"))

	v.SetDefault("addr", defaultAddr)
	v.SetDefault("log_level", int8(defaultLogLevel))
	v.SetDefault("log_outputs", []string{defaultLogOutput})
	v.SetDefault("private_key_path", "")
	v.SetDefault("ca_path", "")

	err := v.ReadInConfig()
	if err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	if err := v.Unmarshal(cfg); err != nil {
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

func (c *Config) ConfigPath() string {
	return c.configPath
}

func (c *Config) IsWrite() bool {
	return c.write
}

func (c *Config) WriteConfig() error {
	return c.v.WriteConfig()
}
