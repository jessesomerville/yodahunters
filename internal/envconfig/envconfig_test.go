package envconfig

import (
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		envVal *string // nil means unset
		def    string
		want   string
	}{
		{
			name:   "returns env var when set",
			key:    "TEST_ENVCONFIG_SET",
			envVal: new("custom_value"),
			def:    "default_value",
			want:   "custom_value",
		},
		{
			name:   "returns default when env var not set",
			key:    "TEST_ENVCONFIG_UNSET",
			envVal: nil,
			def:    "default_value",
			want:   "default_value",
		},
		{
			name:   "returns empty string when env var set to empty",
			key:    "TEST_ENVCONFIG_EMPTY",
			envVal: new(""),
			def:    "default_value",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != nil {
				t.Setenv(tt.key, *tt.envVal)
			}
			if got := GetEnvOrDefault(tt.key, tt.def); got != tt.want {
				t.Errorf("GetEnvOrDefault(%q, %q) = %q, want %q", tt.key, tt.def, got, tt.want)
			}
		})
	}
}

//go:fix inline
func strPtr(s string) *string { return new(s) }
