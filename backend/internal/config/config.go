// backend/internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration values.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Shopee   ShopeeAPIConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host        string
	Port        string
	CorsOrigins []string `mapstructure:"cors_origins"`
}

// DatabaseConfig contains DB connection info.
type DatabaseConfig struct {
	URL string
}

// JWTConfig contains settings for JWT authentication.
type JWTConfig struct {
	Secret string
}

// ShopeeAPIConfig holds credentials for calling the Shopee Partner API.
type ShopeeAPIConfig struct {
	PartnerID    string `mapstructure:"partner_id"`
	PartnerKey   string `mapstructure:"partner_key"`
	ShopID       string `mapstructure:"shop_id"`
	AccessToken  string `mapstructure:"access_token"`
	RefreshToken string `mapstructure:"refresh_token"`
	BaseURL      string `mapstructure:"base_url"`
}

// LoadConfig reads configuration from config.yaml and environment variables.
//   - It expects a file named config.yaml in the working directory.
//   - Environment variables override values from the file, using UPPERCASE and underscores.
//     e.g., SERVER_HOST overrides server.host, DATABASE_URL overrides database.url.
func LoadConfig() (*Config, error) {
	// Tell Viper the file name (without extension) and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")        // look in working directory
	viper.AddConfigPath("./config") // optionally look in ./config folder

	// Environment variables: use uppercase with underscores
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	// Default CORS origin for local development
	viper.SetDefault("server.cors_origins", []string{"http://localhost:5173"})

	// Read from config.yaml
	if err := viper.ReadInConfig(); err != nil {
		// If the file is not found, that’s fatal.
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal into our Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	// Handle slice parsing from env vars
	cfg.Server.CorsOrigins = viper.GetStringSlice("server.cors_origins")

	// Validate required fields
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("database.url must be set in config or via DATABASE_URL")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("jwt.secret must be set in config or via JWT_SECRET")
	}
	if cfg.Server.Port == "" {
		return nil, fmt.Errorf("server.port must be set in config or via SERVER_PORT")
	}

	return &cfg, nil
}

// MustLoadConfig is like LoadConfig but panics on error.
// Use it in main() if you want to fail-fast on missing/invalid config.
func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error loading config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
