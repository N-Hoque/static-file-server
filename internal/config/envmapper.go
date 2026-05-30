package config

import (
	"os"
	"strings"
)

type EnvMapper interface {
	GetEnv(key string) string
	SetEnv(key, value string) error
}

type RealEnvMapper map[string]string

func NewRealEnvMapper() EnvMapper {
	m := make(RealEnvMapper)
	// Populate the map with the current environment variables.
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func (m RealEnvMapper) GetEnv(key string) string {
	value, exists := m[key]
	if !exists {
		return ""
	}
	return value
}

func (m RealEnvMapper) SetEnv(key, value string) error {
	if value == "" {
		delete(m, key)
	} else {
		m[key] = value
	}
	return nil
}
