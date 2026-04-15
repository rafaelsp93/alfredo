package config

import (
	"strings"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type GCalendarConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RefreshToken string `mapstructure:"refresh_token"`
}

type AppConfig struct {
	Timezone string `mapstructure:"timezone"`
}

type AuthConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	GCalendar GCalendarConfig `mapstructure:"gcalendar"`
	App       AppConfig       `mapstructure:"app"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Log       LogConfig       `mapstructure:"log"`
}

// Load reads configuration from config.yaml (optional) and APP_* environment variables.
func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.path", "./data/alfredo.db")
	v.SetDefault("gcalendar.client_id", "")
	v.SetDefault("gcalendar.client_secret", "")
	v.SetDefault("gcalendar.refresh_token", "")
	v.SetDefault("app.timezone", "America/Sao_Paulo")
	v.SetDefault("auth.api_key", "")
	v.SetDefault("log.level", "info")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
