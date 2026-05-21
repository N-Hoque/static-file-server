package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
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

func New() *Config {
	return &Config{
		Debug:         defaultDebug,
		Folder:        defaultFolder,
		Host:          defaultHost,
		Port:          defaultPort,
		Referrers:     defaultReferrers,
		AllowIndex:    defaultAllowIndex,
		ShowListing:   defaultShowListing,
		TLSCert:       defaultTLSCert,
		TLSKey:        defaultTLSKey,
		TLSMinVersStr: defaultTLSMinVers,
		URLPrefix:     defaultURLPrefix,
		Cors:          defaultCors,
		AccessKey:     defaultAccessKey,
	}
}

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

var (
	defaultDebug       = false
	defaultFolder      = "/web"
	defaultHost        = ""
	defaultPort        = uint16(8080)
	defaultReferrers   = []string{}
	defaultAllowIndex  = true
	defaultShowListing = true
	defaultTLSCert     = ""
	defaultTLSKey      = ""
	defaultTLSMinVers  = ""
	defaultURLPrefix   = ""
	defaultCors        = false
	defaultAccessKey   = ""
)

// Load the configuration file.
func Load(filename string, envMapper EnvMapper) (*Config, error) {
	// If no filename provided, assign envvars.
	if filename == "" {
		config := New()
		overrideWithEnvVars(envMapper, config)
		return validate(config)
	}

	configFile, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, err
	}

	overrideWithEnvVars(envMapper, &config)
	return validate(&config)
}

// Log the current configuration.
func Log(config *Config) {
	// YAML marshaling should never error, but if it could, the result is that
	// the contents of the configuration are not logged.

	// Log the configuration.
	slog.Info("current configuration",
		"debug", config.Debug,
		"cors", config.Cors,
		"folder", config.Folder,
		"host", config.Host,
		"port", config.Port,
		"allow-index", config.AllowIndex,
		"show-listing", config.ShowListing,
		"tls-min-vers", config.TLSMinVersStr,
		"url-prefix", config.URLPrefix,
		"referrers", config.Referrers,
		"tls-cert-set", len(config.TLSCert) > 0,
		"tls-key-set", len(config.TLSKey) > 0,
		"access-key-set", len(config.AccessKey) > 0,
	)
}

// overrideWithEnvVars the default values and the configuration file values.
func overrideWithEnvVars(envMapper EnvMapper, config *Config) {
	// Assign envvars, if set.
	config.Cors = envAsBool(envMapper, corsKey, config.Cors)
	config.Debug = envAsBool(envMapper, debugKey, config.Debug)
	config.Folder = envAsStr(envMapper, folderKey, config.Folder)
	config.Host = envAsStr(envMapper, hostKey, config.Host)
	config.Port = envAsUint16(envMapper, portKey, config.Port)
	config.AllowIndex = envAsBool(envMapper, allowIndexKey, config.AllowIndex)
	config.ShowListing = envAsBool(envMapper, showListingKey, config.ShowListing)
	config.TLSCert = envAsStr(envMapper, tlsCertKey, config.TLSCert)
	config.TLSKey = envAsStr(envMapper, tlsKeyKey, config.TLSKey)
	config.TLSMinVersStr = envAsStr(envMapper, tlsMinVersKey, config.TLSMinVersStr)
	config.URLPrefix = envAsStr(envMapper, urlPrefixKey, config.URLPrefix)
	config.Referrers = envAsStrSlice(envMapper, referrersKey, config.Referrers)
	config.AccessKey = envAsStr(envMapper, accessKeyKey, config.AccessKey)
}

// validate the configuration.
func validate(config *Config) (*Config, error) {
	// If HTTPS is to be used, verify both TLS_* environment variables are set.
	useTLS := false
	if len(config.TLSCert) > 0 || len(config.TLSKey) > 0 {
		if len(config.TLSCert) == 0 || len(config.TLSKey) == 0 {
			msg := `if value for either 'TLS_CERT' or 'TLS_KEY' is set then
				value for the other must also be set (values are
				currently '%s' and '%s', respectively)`
			return nil, fmt.Errorf(msg, config.TLSCert, config.TLSKey)
		}
		if _, err := os.Stat(config.TLSCert); nil != err {
			msg := "value of TLS_CERT is set with filename '%s' that returns %v"
			return nil, fmt.Errorf(msg, config.TLSCert, err)
		}
		if _, err := os.Stat(config.TLSKey); nil != err {
			msg := "value of TLS_KEY is set with filename '%s' that returns %v"
			return nil, fmt.Errorf(msg, config.TLSKey, err)
		}
		useTLS = true
	}

	// Verify TLS_MIN_VERS is only (optionally) set if TLS is to be used.
	config.TLSMinVers = tls.VersionTLS10
	if useTLS {
		if len(config.TLSMinVersStr) > 0 {
			var err error
			if config.TLSMinVers, err = tlsMinVersAsUint16(config.TLSMinVersStr); err != nil {
				return nil, err
			}
		}

		// For logging minimum TLS version being used while debugging, backfill
		// the TLSMinVersStr field.
		switch config.TLSMinVers {
		case tls.VersionTLS10:
			config.TLSMinVersStr = "TLS1.0"
		case tls.VersionTLS11:
			config.TLSMinVersStr = "TLS1.1"
		case tls.VersionTLS12:
			config.TLSMinVersStr = "TLS1.2"
		case tls.VersionTLS13:
			config.TLSMinVersStr = "TLS1.3"
		}
	} else {
		if len(config.TLSMinVersStr) > 0 {
			msg := "value for 'TLS_MIN_VERS' is set but 'TLS_CERT' and 'TLS_KEY' are not"
			return nil, errors.New(msg)
		}
	}

	// If the URL path prefix is to be used, verify it is properly formatted.
	if len(config.URLPrefix) > 0 &&
		(!strings.HasPrefix(config.URLPrefix, "/") || strings.HasSuffix(config.URLPrefix, "/")) {
		msg := `if value for 'URL_PREFIX' is set then the value must start
			with '/' and not end with '/' (current value of '%s' vs valid
			example of '/my/prefix')`
		return nil, fmt.Errorf(msg, config.URLPrefix)
	}

	return config, nil
}

// envAsStr returns the value of the environment variable as a string if set.
func envAsStr(envMapper EnvMapper, key, fallback string) string {
	if value := envMapper.GetEnv(key); value != "" {
		return value
	}
	return fallback
}

// envAsStrSlice returns the value of the environment variable as a slice of
// strings if set.
func envAsStrSlice(envMapper EnvMapper, key string, fallback []string) []string {
	if value := envMapper.GetEnv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

// envAsUint16 returns the value of the environment variable as a uint16 if set.
func envAsUint16(envMapper EnvMapper, key string, fallback uint16) uint16 {
	// Retrieve the string value of the environment variable. If not set,
	// fallback is used.
	valueStr := envMapper.GetEnv(key)
	if len(valueStr) == 0 {
		return fallback
	}

	// Parse the string into a uint16.
	base := 10
	bitSize := 16
	valueAsUint64, err := strconv.ParseUint(valueStr, base, bitSize)
	if err != nil {
		slog.Warn(
			"invalid value for key, using fallback",
			"key", key,
			"error", err,
			"fallback", fallback,
		)
		return fallback
	}
	return uint16(valueAsUint64)
}

// envAsBool returns the value for an environment variable or, if not set, a
// fallback value as a boolean.
func envAsBool(envMapper EnvMapper, key string, fallback bool) bool {
	// Retrieve the string value of the environment variable. If not set,
	// fallback is used.
	valueStr := envMapper.GetEnv(key)
	if len(valueStr) == 0 {
		return fallback
	}

	// Parse the string into a boolean.
	value, err := strAsBool(valueStr)
	if err != nil {
		slog.Warn(
			"invalid value for key, using fallback",
			"key", key,
			"error", err,
			"fallback", fallback,
		)
		return fallback
	}
	return value
}

// strAsBool converts the intent of the passed value into a boolean
// representation.
func strAsBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "0", "false", "f", "no", "n":
		return false, nil
	case "1", "true", "t", "yes", "y":
		return true, nil
	default:
		msg := "unknown conversion from string to bool for value '%s'"
		err := fmt.Errorf(msg, value)
		return false, err
	}
}

// tlsMinVersAsUint16 converts the intent of the passed value into an
// enumeration for the crypto/tls package.
func tlsMinVersAsUint16(value string) (uint16, error) {
	switch strings.ToLower(value) {
	case "tls10":
		return tls.VersionTLS10, nil
	case "tls11":
		return tls.VersionTLS11, nil
	case "tls12":
		return tls.VersionTLS12, nil
	case "tls13":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unknown value for TLS_MIN_VERS: %s", value)
	}
}
