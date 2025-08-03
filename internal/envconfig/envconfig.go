// Package envconfig provides helpers for configs and environment variables.
package envconfig

import "os"

// GetEnvOrDefault returns the value of the environment variable specified
// by key, or def if that environment variable isn't set.
func GetEnvOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
