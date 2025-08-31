package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Cors          bool     `yaml:"cors"`
	Debug         bool     `yaml:"debug"`
	Folder        string   `yaml:"folder"`
	Host          string   `yaml:"host"`
	Port          uint16   `yaml:"port"`
	AllowIndex    bool     `yaml:"allow-index"`
	ShowListing   bool     `yaml:"show-listing"`
	TLSCert       string   `yaml:"tls-cert"`
	TLSKey        string   `yaml:"tls-key"`
	TLSMinVers    uint16   `yaml:"-"`
	TLSMinVersStr string   `yaml:"tls-min-vers"`
	URLPrefix     string   `yaml:"url-prefix"`
	Referrers     []string `yaml:"referrers"`
	AccessKey     string   `yaml:"access-key"`
}

// Default values kept as package-level variables for reuse.
var (
	defaultDebug         = false
	defaultFolder        = "/web"
	defaultHost          = ""
	defaultPort          = uint16(8080)
	defaultReferrers     = []string{}
	defaultAllowIndex    = true
	defaultShowListing   = true
	defaultTLSCert       = ""
	defaultTLSKey        = ""
	defaultTLSMinVersion = tls.VersionTLS10
	defaultTLSMinVers    = ""
	defaultURLPrefix     = ""
	defaultCors          = false
	defaultAccessKey     = ""
)

const (
	corsKey        = "CORS"
	debugKey       = "DEBUG"
	folderKey      = "FOLDER"
	hostKey        = "HOST"
	portKey        = "PORT"
	referrersKey   = "REFERRERS"
	allowIndexKey  = "ALLOW_INDEX"
	showListingKey = "SHOW_LISTING"
	tlsCertKey     = "TLS_CERT"
	tlsKeyKey      = "TLS_KEY"
	tlsMinVersKey  = "TLS_MIN_VERS"
	urlPrefixKey   = "URL_PREFIX"
	accessKeyKey   = "ACCESS_KEY"
)

// NewDefaultConfig returns a fresh Config populated with defaults.
func NewDefaultConfig() *Config {
	return &Config{
		Cors:          defaultCors,
		Debug:         defaultDebug,
		Folder:        defaultFolder,
		Host:          defaultHost,
		Port:          defaultPort,
		AllowIndex:    defaultAllowIndex,
		ShowListing:   defaultShowListing,
		TLSCert:       defaultTLSCert,
		TLSKey:        defaultTLSKey,
		TLSMinVers:    uint16(defaultTLSMinVersion),
		TLSMinVersStr: defaultTLSMinVers,
		URLPrefix:     defaultURLPrefix,
		Referrers:     defaultReferrers,
		AccessKey:     defaultAccessKey,
	}
}

// LoadConfig loads configuration from filename (YAML). If filename == "",
// returns a Config populated only from environment variables (via envMapper)
// and defaults. Returns the loaded and validated *Config.
func LoadConfig(filename string, envMapper EnvMapper) (*Config, error) {
	cfg := NewDefaultConfig()

	// If filename provided, load YAML file into cfg.
	if filename != "" {
		f, err := os.Open(filepath.Clean(filename))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
			return nil, err
		}
	}

	// Apply environment overrides and validate.
	cfg.overrideWithEnvVars(envMapper)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Log prints the configuration as YAML.
func (c *Config) Log() {
	contents, _ := yaml.Marshal(c)
	fmt.Println("Using the following configuration:")
	fmt.Println(string(contents))
}

// overrideWithEnvVars updates cfg fields from envMapper when present.
func (c *Config) overrideWithEnvVars(envMapper EnvMapper) {
	c.Cors = envAsBool(envMapper, corsKey, c.Cors)
	c.Debug = envAsBool(envMapper, debugKey, c.Debug)
	c.Folder = envAsStr(envMapper, folderKey, c.Folder)
	c.Host = envAsStr(envMapper, hostKey, c.Host)
	c.Port = envAsUint16(envMapper, portKey, c.Port)
	c.AllowIndex = envAsBool(envMapper, allowIndexKey, c.AllowIndex)
	c.ShowListing = envAsBool(envMapper, showListingKey, c.ShowListing)
	c.TLSCert = envAsStr(envMapper, tlsCertKey, c.TLSCert)
	c.TLSKey = envAsStr(envMapper, tlsKeyKey, c.TLSKey)
	c.TLSMinVersStr = envAsStr(envMapper, tlsMinVersKey, c.TLSMinVersStr)
	c.URLPrefix = envAsStr(envMapper, urlPrefixKey, c.URLPrefix)
	c.Referrers = envAsStrSlice(envMapper, referrersKey, c.Referrers)
	c.AccessKey = envAsStr(envMapper, accessKeyKey, c.AccessKey)
}

// Validate verifies the configuration is sane.
func (c *Config) Validate() error {
	useTLS := false
	if len(c.TLSCert) > 0 || len(c.TLSKey) > 0 {
		if len(c.TLSCert) == 0 || len(c.TLSKey) == 0 {
			msg := `if value for either 'TLS_CERT' or 'TLS_KEY' is set then
				value for the other must also be set (values are
				currently '%s' and '%s', respectively)`
			return fmt.Errorf(msg, c.TLSCert, c.TLSKey)
		}
		if _, err := os.Stat(c.TLSCert); nil != err {
			msg := "value of TLS_CERT is set with filename '%s' that returns %v"
			return fmt.Errorf(msg, c.TLSCert, err)
		}
		if _, err := os.Stat(c.TLSKey); nil != err {
			msg := "value of TLS_KEY is set with filename '%s' that returns %v"
			return fmt.Errorf(msg, c.TLSKey, err)
		}
		useTLS = true
	}

	c.TLSMinVers = tls.VersionTLS10
	if useTLS {
		if len(c.TLSMinVersStr) > 0 {
			var err error
			if c.TLSMinVers, err = tlsMinVersAsUint16(c.TLSMinVersStr); err != nil {
				return err
			}
		}

		switch c.TLSMinVers {
		case tls.VersionTLS10:
			c.TLSMinVersStr = "TLS1.0"
		case tls.VersionTLS11:
			c.TLSMinVersStr = "TLS1.1"
		case tls.VersionTLS12:
			c.TLSMinVersStr = "TLS1.2"
		case tls.VersionTLS13:
			c.TLSMinVersStr = "TLS1.3"
		}
	} else {
		if len(c.TLSMinVersStr) > 0 {
			msg := "value for 'TLS_MIN_VERS' is set but 'TLS_CERT' and 'TLS_KEY' are not"
			return errors.New(msg)
		}
	}

	if len(c.URLPrefix) > 0 &&
		(!strings.HasPrefix(c.URLPrefix, "/") || strings.HasSuffix(c.URLPrefix, "/")) {
		msg := `if value for 'URL_PREFIX' is set then the value must start
			with '/' and not end with '/' (current value of '%s' vs valid
			example of '/my/prefix')`
		return fmt.Errorf(msg, c.URLPrefix)
	}

	return nil
}

// env helper functions (unchanged semantics, operate with provided fallback).

func envAsStr(envMapper EnvMapper, key, fallback string) string {
	if value := envMapper.GetEnv(key); value != "" {
		return value
	}
	return fallback
}

func envAsStrSlice(envMapper EnvMapper, key string, fallback []string) []string {
	if value := envMapper.GetEnv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

func envAsUint16(envMapper EnvMapper, key string, fallback uint16) uint16 {
	valueStr := envMapper.GetEnv(key)
	if len(valueStr) == 0 {
		return fallback
	}
	base := 10
	bitSize := 16
	valueAsUint64, err := strconv.ParseUint(valueStr, base, bitSize)
	if err != nil {
		log.Printf(
			"Invalid value for '%s': %v\nUsing fallback: %d",
			key, err, fallback,
		)
		return fallback
	}
	return uint16(valueAsUint64)
}

func envAsBool(envMapper EnvMapper, key string, fallback bool) bool {
	valueStr := envMapper.GetEnv(key)
	if len(valueStr) == 0 {
		return fallback
	}
	value, err := strAsBool(valueStr)
	if err != nil {
		log.Printf(
			"Invalid value for '%s': %v\nUsing fallback: %t",
			key, err, fallback,
		)
		return fallback
	}
	return value
}

func strAsBool(value string) (result bool, err error) {
	switch strings.ToLower(value) {
	case "0", "false", "f", "no", "n":
		result = false
	case "1", "true", "t", "yes", "y":
		result = true
	default:
		result = false
		msg := "unknown conversion from string to bool for value '%s'"
		err = fmt.Errorf(msg, value)
	}
	return
}

func tlsMinVersAsUint16(value string) (result uint16, err error) {
	switch strings.ToLower(value) {
	case "tls10":
		result = tls.VersionTLS10
	case "tls11":
		result = tls.VersionTLS11
	case "tls12":
		result = tls.VersionTLS12
	case "tls13":
		result = tls.VersionTLS13
	default:
		err = fmt.Errorf("unknown value for TLS_MIN_VERS: %s", value)
	}
	return
}
