package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	App      AppConfig      `mapstructure:"app"`
	RustDesk RustDeskConfig `mapstructure:"rustdesk"`
	OIDC     OIDCConfig     `mapstructure:"oidc"`
	LDAP     LDAPConfig     `mapstructure:"ldap"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Addr     string `mapstructure:"addr"`
	Mode     string `mapstructure:"mode"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Key         string `mapstructure:"key"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Title            string `mapstructure:"title"`
	Register         bool   `mapstructure:"register"`
	CaptchaThreshold int    `mapstructure:"captcha_threshold"`
	WebClient        bool   `mapstructure:"web_client"`
}

// RustDeskConfig holds RustDesk server connection configuration.
type RustDeskConfig struct {
	IDServer  string `mapstructure:"id_server"`
	Relay     string `mapstructure:"relay_server"`
	APIServer string `mapstructure:"api_server"`
	Key       string `mapstructure:"key"`
	KeyFile   string `mapstructure:"key_file"`
}

// OIDCConfig holds OIDC authentication configuration.
type OIDCConfig struct {
	Enable      bool   `mapstructure:"enable"`
	ProviderURL string `mapstructure:"provider_url"`
	ClientID    string `mapstructure:"client_id"`
	Secret      string `mapstructure:"client_secret"`
}

// LDAPConfig holds LDAP configuration.
type LDAPConfig struct {
	Enable bool   `mapstructure:"enable"`
	URL    string `mapstructure:"url"`
	BaseDN string `mapstructure:"base_dn"`
}

// SMTPConfig holds SMTP email configuration.
type SMTPConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

// Load reads configuration from file and environment variables.
// Environment variables use the prefix RUSTDESK_API_ and replace dots with underscores.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.addr", "0.0.0.0:21114")
	v.SetDefault("server.mode", "release")
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.path", "./data/api.db")
	v.SetDefault("jwt.key", "")
	v.SetDefault("jwt.expire_hours", 168)
	v.SetDefault("app.title", "RustDesk Admin")
	v.SetDefault("app.register", true)
	v.SetDefault("app.captcha_threshold", 3)
	v.SetDefault("app.web_client", false)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.path", "./data/api.log")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Environment variable override: RUSTDESK_API_SERVER_ADDR, RUSTDESK_API_JWT_KEY, etc.
	v.SetEnvPrefix("RUSTDESK_API")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is acceptable; use defaults + env
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
