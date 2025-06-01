package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Resend   ResendConfig
	Redis    RedisConfig
	Metrics  MetricsConfig
	Contact  ContactConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string `mapstructure:"ssl_mode"`
}

type ResendConfig struct {
	APIKey    string `mapstructure:"api_key"`
	FromEmail string `mapstructure:"from_email"`
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type MetricsConfig struct {
	Enabled bool
	Port    int
}

type ContactConfig struct {
	Email string `mapstructure:"email"`
}

func LoadConfig() (*Config, error) {
	// Cargar variables de entorno desde .env si existe
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

	// Configurar viper para leer variables de entorno
	viper.AutomaticEnv()

	// Mapear variables de entorno a las claves de configuración
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "SERVER_MODE")

	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.name", "DATABASE_NAME")
	viper.BindEnv("database.ssl_mode", "DATABASE_SSL_MODE")

	viper.BindEnv("resend.api_key", "RESEND_API_KEY")
	viper.BindEnv("resend.from_email", "RESEND_FROM_EMAIL")

	viper.BindEnv("contact.email", "CONTACT_EMAIL")

	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	viper.BindEnv("metrics.enabled", "METRICS_ENABLED")
	viper.BindEnv("metrics.port", "METRICS_PORT")

	// Intentar cargar config.yaml como fallback
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Si existe config.yaml, leerlo (las env vars tienen prioridad)
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No config.yaml found, using only environment variables: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
