// backend/internal/config/config.go
package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration values.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Cache       CacheConfig
	Performance PerformanceConfig
	JWT         JWTConfig
	Shopee      ShopeeAPIConfig `mapstructure:"shopee_api"`
	Logging     LoggingConfig
	MaxThreads  int `mapstructure:"max_threads"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host        string
	Port        string
	CorsOrigins []string `mapstructure:"cors_origins"`
}

// DatabaseConfig contains DB connection info.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

// JWTConfig contains settings for JWT authentication.
type JWTConfig struct {
	Secret string
}

// CacheConfig contains Redis cache settings.
type CacheConfig struct {
	RedisURL     string `mapstructure:"redis_url"`
	Password     string
	DB           int
	MaxRetries   int    `mapstructure:"max_retries"`
	DialTimeout  string `mapstructure:"dial_timeout"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
	DefaultTTL   string `mapstructure:"default_ttl"`
	Enabled      bool
}

// PerformanceConfig contains performance-related settings.
type PerformanceConfig struct {
	BatchSize              int    `mapstructure:"batch_size"`
	SlowQueryThreshold     string `mapstructure:"slow_query_threshold"`
	ShopeeRateLimit        int    `mapstructure:"shopee_rate_limit"`
	ShopeeRetryMaxAttempts int    `mapstructure:"shopee_retry_max_attempts"`
	ShopeeRetryDelay       string `mapstructure:"shopee_retry_delay"`
	EnableMetrics          bool   `mapstructure:"enable_metrics"`
}

// LoggingConfig specifies where log files are stored.
type LoggingConfig struct {
	Dir string
}

// ShopeeAPIConfig holds credentials for calling the Shopee Partner API.
type ShopeeAPIConfig struct {
	PartnerID     string `mapstructure:"partner_id"`
	PartnerKey    string `mapstructure:"partner_key"`
	ShopID        string `mapstructure:"shop_id"`
	AccessToken   string `mapstructure:"access_token"`
	RefreshToken  string `mapstructure:"refresh_token"`
	BaseURL       string `mapstructure:"base_url"`
	BaseURLShopee string `mapstructure:"base_url_shopee"`
	AuthURL       string `mapstructure:"auth_url"`
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
	viper.SetDefault("logging.dir", "logs")
	viper.SetDefault("max_threads", 5)

	// Cache defaults
	viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.redis_url", "redis://localhost:6379")
	viper.SetDefault("cache.db", 0)
	viper.SetDefault("cache.max_retries", 3)
	viper.SetDefault("cache.dial_timeout", "5s")
	viper.SetDefault("cache.read_timeout", "3s")
	viper.SetDefault("cache.write_timeout", "3s")
	viper.SetDefault("cache.default_ttl", "5m")

	// Database connection pool defaults
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	// Performance defaults
	viper.SetDefault("performance.batch_size", 100)
	viper.SetDefault("performance.slow_query_threshold", "2s")
	viper.SetDefault("performance.shopee_rate_limit", 1000)
	viper.SetDefault("performance.shopee_retry_max_attempts", 3)
	viper.SetDefault("performance.shopee_retry_delay", "1s")
	viper.SetDefault("performance.enable_metrics", true)

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
	cfg.Logging.Dir = viper.GetString("logging.dir")
	cfg.MaxThreads = viper.GetInt("max_threads")
	if cfg.MaxThreads <= 0 {
		cfg.MaxThreads = 5
	}

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

// ShopeeAuthURL builds the authorization link for a store.
func (c *Config) ShopeeAuthURL(storeID int64) string {
	base := c.Shopee.BaseURLShopee
	if base == "" {
		base = c.Shopee.AuthURL
	}
	if base == "" {
		base = "https://partner.test-stable.shopeemobile.com"
	}
	path := "/api/v2/shop/auth_partner"
	ts := time.Now().Unix()
	msg := fmt.Sprintf("%s%s%d", c.Shopee.PartnerID, path, ts)
	h := hmac.New(sha256.New, []byte(c.Shopee.PartnerKey))
	h.Write([]byte(msg))
	sign := hex.EncodeToString(h.Sum(nil))
	redirect := fmt.Sprintf("%s/stores/%d", strings.TrimSuffix(c.Server.CorsOrigins[0], "/"), storeID)
	return fmt.Sprintf("%s%s?partner_id=%s&timestamp=%d&sign=%s&redirect=%s",
		base, path, c.Shopee.PartnerID, ts, sign, url.QueryEscape(redirect))
}
