package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DbUser          string `mapstructure:"DB_USER"`
	DbPassword      string `mapstructure:"DB_PASSWORD"`
	DbName          string `mapstructure:"DB_NAME"`
	DbHost          string `mapstructure:"DB_HOST"`
	DbPort          string `mapstructure:"DB_PORT"`
	AuthSecret      string `mapstructure:"AUTH_SECRET"`
	JwtSecret       string `mapstructure:"JWT_SECRET"`
	JwtIssuer       string `mapstructure:"JWT_ISSUER"`
	ApiKey          string `mapstructure:"API_KEY"`
	HostDomain      string `mapstructure:"HOST_DOMAIN"`
	ServerPort      string `mapstructure:"SERVER_PORT"`
	LogType         string `mapstructure:"LOG_TYPE"`
	StorageRegion   string `mapstructure:"STORAGE_REGION"`
	StorageBucket   string `mapstructure:"STORAGE_BUCKET"`
	StorageBasePath string `mapstructure:"STORAGE_BASE_PATH"`
	StorageBaseUrl  string `mapstructure:"STORAGE_BASE_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config-local")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		// Get from config.env if config-local.env is not found
		viper.SetConfigName("config")
		viper.SetConfigType("env")
		viper.AutomaticEnv()

		err = viper.ReadInConfig()
		if err != nil {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return config, err
}
