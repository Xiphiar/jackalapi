package jutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadEnvVarOrFallback(t *testing.T) {
	t.Setenv("MOCK_ENV_VAR", "lupulella-2")

	tt := []struct {
		name     string
		varId    string
		fallBack string
	}{
		{
			name:  "env var exists",
			varId: "MOCK_ENV_VAR",
		},
		{
			name:     "env var doesn't exist",
			varId:    "I_DONT_EXIST",
			fallBack: "fallback_var",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			value := LoadEnvVarOrFallback(tc.varId, tc.fallBack)

			if len(tc.fallBack) > 0 {
				require.Equal(t, value, "fallback_var")
			} else {
				require.Equal(t, value, "lupulella-2")
			}
		})
	}
}
