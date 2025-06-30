package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvLoader_Load(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("TIMEOUT", "10s")
	os.Setenv("IDLE_TIMEOUT", "30s")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "mydb")

	loader := NewLoader()
	cfg, err := loader.Load()

	assert.NoError(t, err)
	assert.Equal(t, "local", cfg.App.Env)
	assert.Equal(t, ":8080", cfg.HTTP.Address)
	assert.Equal(t, "5432", cfg.Storage.Port)
	assert.Equal(t, "mydb", cfg.Storage.Name)

	os.Clearenv()
}
