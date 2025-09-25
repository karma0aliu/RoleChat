package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string `mapstructure:"sslmode"`
}

type LogConfig struct {
	Level string
}

type JWTConfig struct {
	AccessSecret        string `mapstructure:"access_secret"`
	RefreshSecret       string `mapstructure:"refresh_secret"`
	AccessExpiresMins   int    `mapstructure:"access_expires_mins"`
	RefreshExpiresHours int    `mapstructure:"refresh_expires_hours"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
