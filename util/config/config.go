package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DbUser                      string `mapstructure:"DB_USER"`
	DbPassword                  string `mapstructure:"DB_PASSWORD"`
	DbName                      string `mapstructure:"DB_NAME"`
	DbHost                      string `mapstructure:"DB_HOST"`
	DbPort                      string `mapstructure:"DB_PORT"`
	DbDriver                    string `mapstructure:"DB_DRIVER"`
	DbTimezone                  string `mapstructure:"DB_TIMEZONE"`
	DbMaxOpenConns              int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	DbMaxIdleConns              int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	JwtSecret                   string `mapstructure:"JWT_SECRET"`
	JwtIssuer                   string `mapstructure:"JWT_ISSUER"`
	ApiKey                      string `mapstructure:"API_KEY"`
	HostDomain                  string `mapstructure:"HOST_DOMAIN"`
	ServerPort                  string `mapstructure:"SERVER_PORT"`
	LogType                     string `mapstructure:"LOG_TYPE"`
	StorageRegion               string `mapstructure:"STORAGE_REGION"`
	StorageBucket               string `mapstructure:"STORAGE_BUCKET"`
	StorageBasePath             string `mapstructure:"STORAGE_BASE_PATH"`
	StorageBaseUrl              string `mapstructure:"STORAGE_BASE_URL"`
	OpenTelemetryMetricsEnabled bool   `mapstructure:"OTEL_METRICS_ENABLED"`
	OpenTelemetryTraceEndPoint  string `mapstructure:"OTEL_TRACE_ENDPOINT"`
}

func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err = v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return config, fmt.Errorf(
				"config.env not found in %q: copy config.env.example to config.env and fill in the values",
				path,
			)
		}
		return config, fmt.Errorf("failed to read config: %w", err)
	}

	err = v.Unmarshal(&config)
	return config, err
}
