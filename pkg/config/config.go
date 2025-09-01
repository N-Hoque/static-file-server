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

var Target Config

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

func init() {
	// init calls setDefaults to better support testing.
	setDefaults()
}

func setDefaults() {
	Target.Debug = defaultDebug
	Target.Folder = defaultFolder
	Target.Host = defaultHost
	Target.Port = defaultPort
	Target.Referrers = defaultReferrers
	Target.AllowIndex = defaultAllowIndex
	Target.ShowListing = defaultShowListing
	Target.TLSCert = defaultTLSCert
	Target.TLSKey = defaultTLSKey
	Target.TLSMinVersStr = defaultTLSMinVers
	Target.URLPrefix = defaultURLPrefix
	Target.Cors = defaultCors
	Target.AccessKey = defaultAccessKey
}

// Load the configuration file.
func Load(filename string, envMapper EnvMapper) error {
	// If no filename provided, assign envvars.
	if filename == "" {
		overrideWithEnvVars(envMapper)
		return validate()
	}

	configFile, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer configFile.Close()

	if err := yaml.NewDecoder(configFile).Decode(&Target); err != nil {
		return err
	}

	overrideWithEnvVars(envMapper)
	return validate()
}

// Log the current configuration.
func Log() {
	// YAML marshaling should never error, but if it could, the result is that
	// the contents of the configuration are not logged.
	contents, _ := yaml.Marshal(&Target)

	// Log the configuration.
	fmt.Println("Using the following configuration:")
	fmt.Println(string(contents))
}

// overrideWithEnvVars the default values and the configuration file values.
func overrideWithEnvVars(envMapper EnvMapper) {
	// Assign envvars, if set.
	Target.Cors = envAsBool(envMapper, corsKey, Target.Cors)
	Target.Debug = envAsBool(envMapper, debugKey, Target.Debug)
	Target.Folder = envAsStr(envMapper, folderKey, Target.Folder)
	Target.Host = envAsStr(envMapper, hostKey, Target.Host)
	Target.Port = envAsUint16(envMapper, portKey, Target.Port)
	Target.AllowIndex = envAsBool(envMapper, allowIndexKey, Target.AllowIndex)
	Target.ShowListing = envAsBool(envMapper, showListingKey, Target.ShowListing)
	Target.TLSCert = envAsStr(envMapper, tlsCertKey, Target.TLSCert)
	Target.TLSKey = envAsStr(envMapper, tlsKeyKey, Target.TLSKey)
	Target.TLSMinVersStr = envAsStr(envMapper, tlsMinVersKey, Target.TLSMinVersStr)
	Target.URLPrefix = envAsStr(envMapper, urlPrefixKey, Target.URLPrefix)
	Target.Referrers = envAsStrSlice(envMapper, referrersKey, Target.Referrers)
	Target.AccessKey = envAsStr(envMapper, accessKeyKey, Target.AccessKey)
}

// validate the configuration.
func validate() error {
	// If HTTPS is to be used, verify both TLS_* environment variables are set.
	useTLS := false
	if len(Target.TLSCert) > 0 || len(Target.TLSKey) > 0 {
		if len(Target.TLSCert) == 0 || len(Target.TLSKey) == 0 {
			msg := `if value for either 'TLS_CERT' or 'TLS_KEY' is set then
				value for the other must also be set (values are
				currently '%s' and '%s', respectively)`
			return fmt.Errorf(msg, Target.TLSCert, Target.TLSKey)
		}
		if _, err := os.Stat(Target.TLSCert); nil != err {
			msg := "value of TLS_CERT is set with filename '%s' that returns %v"
			return fmt.Errorf(msg, Target.TLSCert, err)
		}
		if _, err := os.Stat(Target.TLSKey); nil != err {
			msg := "value of TLS_KEY is set with filename '%s' that returns %v"
			return fmt.Errorf(msg, Target.TLSKey, err)
		}
		useTLS = true
	}

	// Verify TLS_MIN_VERS is only (optionally) set if TLS is to be used.
	Target.TLSMinVers = tls.VersionTLS10
	if useTLS {
		if len(Target.TLSMinVersStr) > 0 {
			var err error
			if Target.TLSMinVers, err = tlsMinVersAsUint16(Target.TLSMinVersStr); err != nil {
				return err
			}
		}

		// For logging minimum TLS version being used while debugging, backfill
		// the TLSMinVersStr field.
		switch Target.TLSMinVers {
		case tls.VersionTLS10:
			Target.TLSMinVersStr = "TLS1.0"
		case tls.VersionTLS11:
			Target.TLSMinVersStr = "TLS1.1"
		case tls.VersionTLS12:
			Target.TLSMinVersStr = "TLS1.2"
		case tls.VersionTLS13:
			Target.TLSMinVersStr = "TLS1.3"
		}
	} else {
		if len(Target.TLSMinVersStr) > 0 {
			msg := "value for 'TLS_MIN_VERS' is set but 'TLS_CERT' and 'TLS_KEY' are not"
			return errors.New(msg)
		}
	}

	// If the URL path prefix is to be used, verify it is properly formatted.
	if len(Target.URLPrefix) > 0 &&
		(!strings.HasPrefix(Target.URLPrefix, "/") || strings.HasSuffix(Target.URLPrefix, "/")) {
		msg := `if value for 'URL_PREFIX' is set then the value must start
			with '/' and not end with '/' (current value of '%s' vs valid
			example of '/my/prefix')`
		return fmt.Errorf(msg, Target.URLPrefix)
	}

	return nil
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
		log.Printf(
			"Invalid value for '%s': %v\nUsing fallback: %d",
			key, err, fallback,
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
		log.Printf(
			"Invalid value for '%s': %v\nUsing fallback: %t",
			key, err, fallback,
		)
		return fallback
	}
	return value
}

// strAsBool converts the intent of the passed value into a boolean
// representation.
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

// tlsMinVersAsUint16 converts the intent of the passed value into an
// enumeration for the crypto/tls package.
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
