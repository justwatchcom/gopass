package config_test

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestHomedir(t *testing.T) {
	assert.NotEqual(t, config.Homedir(), "")
}

func TestNewConfig(t *testing.T) {
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")))

	cfg := config.New()
	cs := cfg.String()
	assert.Contains(t, cs, `&config.Config{AutoClip:false, AutoImport:true, ClipTimeout:45, ExportKeys:true, NoPager:false, Notifications:true,`)
	assert.Contains(t, cs, `SafeContent:false, Mounts:map[string]string{},`)

	cfg = &config.Config{
		Mounts: make(map[string]string, 2),
	}
	cfg.Mounts["foo"] = ""
	cfg.Mounts["bar"] = ""
	cs = cfg.String()
	assert.Contains(t, cs, `&config.Config{AutoClip:false, AutoImport:false, ClipTimeout:0, ExportKeys:false, NoPager:false, Notifications:false,`)
	assert.Contains(t, cs, `SafeContent:false, Mounts:map[string]string{"bar":"", "foo":""},`)
}

func TestSetConfigValue(t *testing.T) {
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(os.TempDir(), ".gopass.yml")))

	cfg := config.New()
	assert.NoError(t, cfg.SetConfigValue("autoclip", "true"))
	assert.NoError(t, cfg.SetConfigValue("cliptimeout", "900"))
	assert.NoError(t, cfg.SetConfigValue("path", "/tmp"))
	assert.Error(t, cfg.SetConfigValue("autoclip", "yo"))
}
