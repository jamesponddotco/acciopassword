// Package config contains the configuration for the application.
package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrInvalidConfigFile is returned when the configuration file is invalid.
	ErrInvalidConfigFile xerrors.Error = "invalid configuration file"
)

const (
	// DefaultAddress is the default address of the application.
	DefaultAddress string = ":1997"

	// DefaultPID is the default path to the PID file.
	DefaultPID string = "/var/run/acciopassword.pid"

	// DefaultDSN is the default data source name for the SQLite database.
	DefaultDSN string = "file:/var/share/acciopassword/sqlite.db?cache=shared&mode=rwc&_pragma_cache_size=-20000&_journal_mode=WAL&_synchronous=NORMAL"

	// DefaultMinTLSVersion is the default minimum TLS version supported by the
	// server.
	DefaultMinTLSVersion string = "TLS13"
)

// TLS represents the TLS configuration.
type TLS struct {
	// Certificate is the path to the TLS certificate.
	Certificate string `json:"certificate"`

	// Key is the path to the TLS key.
	Key string `json:"key"`

	// Version is the TLS version to use.
	Version string `json:"version"`
}

// Server represents the server configuration.
type Server struct {
	// TLS is the TLS configuration.
	TLS *TLS `json:"tls"`

	// Address is the address of the application.
	Address string `json:"address"`

	// PID is the path to the PID file.
	PID string `json:"pid"`
}

// Database represents the database configuration.
type Database struct {
	// DSN is the data source name.
	DSN string `json:"dsn"`
}

// Config represents the application configuration.
type Config struct {
	// Server is the server configuration.
	Server *Server `json:"server"`

	// Database is the database configuration.
	Database *Database `json:"database"`

	// PrivacyPolicy is the link to the service's privacy policy.
	PrivacyPolicy string `json:"privacyPolicy"`

	// TermsOfService is the link to the service's terms of service.
	TermsOfService string `json:"termsOfService"`
}

// LoadConfig opens a file and reads the configuration from it.
func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfigFile, err)
	}
	defer configFile.Close()

	var cfg *Config
	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfigFile, err)
	}

	if cfg.Server.Address == "" {
		cfg.Server.Address = DefaultAddress
	}

	if cfg.Server.PID == "" {
		cfg.Server.PID = DefaultPID
	}

	if cfg.Database.DSN == "" {
		cfg.Database.DSN = DefaultDSN
	}

	if cfg.Server.TLS.Version == "" {
		cfg.Server.TLS.Version = DefaultMinTLSVersion
	}

	return cfg, nil
}

// Validate checks the configuration for errors.
func (cfg *Config) Validate() error {
	if cfg.Server == nil {
		return fmt.Errorf("%w: missing server configuration", ErrInvalidConfigFile)
	}

	if cfg.Server.TLS == nil {
		return fmt.Errorf("%w: missing TLS configuration", ErrInvalidConfigFile)
	}

	if cfg.Server.TLS.Certificate == "" {
		return fmt.Errorf("%w: missing TLS certificate", ErrInvalidConfigFile)
	}

	if cfg.Server.TLS.Key == "" {
		return fmt.Errorf("%w: missing TLS key", ErrInvalidConfigFile)
	}

	if cfg.Database == nil {
		return fmt.Errorf("%w: missing database configuration", ErrInvalidConfigFile)
	}

	if cfg.Database.DSN == "" {
		return fmt.Errorf("%w: missing database DSN", ErrInvalidConfigFile)
	}

	if cfg.PrivacyPolicy == "" {
		return fmt.Errorf("%w: missing privacy policy URL", ErrInvalidConfigFile)
	}

	if cfg.TermsOfService == "" {
		return fmt.Errorf("%w: missing terms of service URL", ErrInvalidConfigFile)
	}

	if _, err := url.Parse(cfg.PrivacyPolicy); err != nil {
		return fmt.Errorf("%w: invalid privacy policy URL: %w", ErrInvalidConfigFile, err)
	}

	if _, err := url.Parse(cfg.TermsOfService); err != nil {
		return fmt.Errorf("%w: invalid terms of service URL: %w", ErrInvalidConfigFile, err)
	}

	return nil
}
