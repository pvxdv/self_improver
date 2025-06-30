package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		wantErrors []string
	}{
		{
			name: "valid config",
			config: Config{
				Address:     ":8080",
				Timeout:     5 * time.Second,
				IdleTimeout: 120 * time.Second,
			},
			wantErrors: nil,
		},
		{
			name: "empty address",
			config: Config{
				Address:     "",
				Timeout:     5 * time.Second,
				IdleTimeout: 120 * time.Second,
			},
			wantErrors: []string{
				errEmptyAddress,
			},
		},
		{
			name: "timeout too small",
			config: Config{
				Address:     ":8080",
				Timeout:     5 * time.Millisecond,
				IdleTimeout: 120 * time.Second,
			},
			wantErrors: []string{
				errInvalidTimeout,
			},
		},
		{
			name: "timeout too large",
			config: Config{
				Address:     ":8080",
				Timeout:     6 * time.Second,
				IdleTimeout: 120 * time.Second,
			},
			wantErrors: []string{
				errInvalidTimeout,
			},
		},
		{
			name: "idle timeout too small",
			config: Config{
				Address:     ":8080",
				Timeout:     5 * time.Second,
				IdleTimeout: 30 * time.Second,
			},
			wantErrors: []string{
				errInvalidIdleTimeout,
			},
		},
		{
			name: "idle timeout too large",
			config: Config{
				Address:     ":8080",
				Timeout:     5 * time.Second,
				IdleTimeout: 400 * time.Second,
			},
			wantErrors: []string{
				errInvalidIdleTimeout,
			},
		},
		{
			name: "multiple errors",
			config: Config{
				Address:     "",
				Timeout:     1 * time.Millisecond,
				IdleTimeout: 10 * time.Second,
			},
			wantErrors: []string{
				errEmptyAddress,
				errInvalidTimeout,
				errInvalidIdleTimeout,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()

			if tt.wantErrors == nil {
				assert.Empty(t, errs)
				return
			}

			require.Equal(t, len(tt.wantErrors), len(errs), "unexpected number of errors")

			for i, wantErr := range tt.wantErrors {
				assert.ErrorContains(t, errs[i], wantErr)
			}
		})
	}
}

func TestTimeoutConstants(t *testing.T) {
	assert.Equal(t, 10*time.Millisecond, MinTimeout)
	assert.Equal(t, 5*time.Second, MaxTimeout)
	assert.Equal(t, 60*time.Second, MinIdleTimeout)
	assert.Equal(t, 300*time.Second, MaxIdleTimeout)
}
