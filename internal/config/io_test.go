package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigs(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  string
		want *Config
	}{
		{
			name: "1.9.3",
			cfg: `autoclip: true
autoimport: false
cliptimeout: 45
exportkeys: true
nopager: false
notifications: true
path: /home/johndoe/.password-store
safecontent: false
mounts:
  foo/sub: /home/johndoe/.password-store-foo-sub
  work: /home/johndoe/.password-store-work`,
			want: &Config{
				AutoClip:      true,
				AutoImport:    false,
				ClipTimeout:   45,
				ExportKeys:    true,
				NoPager:       false,
				Notifications: true,
				Parsing:       true,
				Path:          "/home/johndoe/.password-store",
				SafeContent:   false,
				Mounts: map[string]string{
					"foo/sub": "/home/johndoe/.password-store-foo-sub",
					"work":    "/home/johndoe/.password-store-work",
				},
			},
		}, {
			name: "N+1",
			cfg: `autoclip: true
autoimport: false
cliptimeout: 45
exportkeys: true
nopager: false
foo: bar
notifications: true
path: /home/johndoe/.password-store
safecontent: false
mounts:
  foo/sub: /home/johndoe/.password-store-foo-sub
  work: /home/johndoe/.password-store-work`,
			want: &Config{
				AutoClip:      true,
				AutoImport:    false,
				ClipTimeout:   45,
				ExportKeys:    true,
				NoPager:       false,
				Notifications: true,
				Parsing:       true,
				Path:          "/home/johndoe/.password-store",
				SafeContent:   false,
				Mounts: map[string]string{
					"foo/sub": "/home/johndoe/.password-store-foo-sub",
					"work":    "/home/johndoe/.password-store-work",
				},
				XXX: map[string]interface{}{"foo": string("bar")},
			},
		}, {
			name: "1.8.2",
			cfg: `root:
  autoclip: true
  autoimport: false
  autosync: false
  check_recipient_hash: false
  cliptimeout: 45
  concurrency: 50
  editrecipients: true
  exportkeys: true
  confirm: false
  nopager: false
  notficiations: true
  path: gpgcli-gitcli-fs+file:///home/johndoe/.password-store
  safecontent: false
  usesymbols: true
mounts:
  foo/sub:
    autoclip: true
    autoimport: false
    autosync: false
    check_recipient_hash: false
    cliptimeout: 45
    concurrency: 50
    editrecipients: true
    exportkeys: true
    confirm: false
    nopager: false
    notficiations: true
    path: gpgcli-gitcli-fs+file:///home/johndoe/.password-store-foo-sub
    safecontent: false
    usesymbols: true
  work:
    autoclip: true
    autoimport: false
    autosync: false
    check_recipient_hash: false
    cliptimeout: 45
    concurrency: 50
    editrecipients: true
    exportkeys: true
    confirm: false
    nopager: false
    notficiations: true
    path: gpgcli-gitcli-fs+file:///home/johndoe/.password-store-work
    safecontent: false
    usesymbols: true
`,
			want: &Config{
				AutoClip:      true,
				AutoImport:    false,
				ClipTimeout:   45,
				ExportKeys:    false,
				NoPager:       false,
				Notifications: false,
				Parsing:       true,
				Path:          "/home/johndoe/.password-store",
				SafeContent:   false,
				Mounts: map[string]string{
					"foo/sub": "/home/johndoe/.password-store-foo-sub",
					"work":    "/home/johndoe/.password-store-work",
				},
			},
		}, {
			name: "1.4.0",
			cfg: `root:
  askformore: false
  autoimport: false
  autosync: false
  cliptimeout: 45
  noconfirm: false
  nopager: false
  path: /home/johndoe/.password-store
  safecontent: false
mounts:
  foo/sub:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-foo-sub
    safecontent: false
  work:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-work
    safecontent: false
version: 1.4.0`,
			want: &Config{
				AutoClip:      false,
				AutoImport:    false,
				ClipTimeout:   45,
				ExportKeys:    false,
				NoPager:       false,
				Notifications: false,
				Parsing:       true,
				Path:          "/home/johndoe/.password-store",
				SafeContent:   false,
				Mounts: map[string]string{
					"foo/sub": "/home/johndoe/.password-store-foo-sub",
					"work":    "/home/johndoe/.password-store-work",
				},
			},
		}, {
			name: "1.3.0",
			cfg: `askformore: false
autoimport: true
autosync: false
cliptimeout: 45
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/foo/.password-store
safecontent: true
version: "1.3.0"`,
			want: &Config{
				AutoClip:      false,
				AutoImport:    true,
				ClipTimeout:   45,
				ExportKeys:    false,
				NoPager:       false,
				Notifications: false,
				Parsing:       true,
				Path:          "/home/foo/.password-store",
				SafeContent:   true,
				Mounts: map[string]string{
					"dev":       "/Users/johndoe/.password-store-dev",
					"ops":       "/Users/johndoe/.password-store-ops",
					"personal":  "/Users/johndoe/secrets",
					"teststore": "/Users/johndoe/tmp/teststore",
				},
			},
		}, {
			name: "1.2.0",
			cfg: `alwaystrust: true
askformore: false
autoimport: true
autopull: true
autopush: true
cliptimeout: 45
debug: false
loadkeys: true
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
nocolor: false
noconfirm: false
path: /home/foo/.password-store
persistkeys: true
safecontent: true
version: "1.2.0"`,
			want: &Config{
				AutoClip:      false,
				AutoImport:    true,
				ClipTimeout:   45,
				ExportKeys:    false,
				NoPager:       false,
				Notifications: false,
				Parsing:       true,
				Path:          "/home/foo/.password-store",
				SafeContent:   true,
				Mounts: map[string]string{
					"dev":       "/Users/johndoe/.password-store-dev",
					"ops":       "/Users/johndoe/.password-store-ops",
					"personal":  "/Users/johndoe/secrets",
					"teststore": "/Users/johndoe/tmp/teststore",
				},
			},
		}, {
			name: "1.1.0",
			cfg: `alwaystrust: false
autoimport: false
autopull: true
autopush: true
cliptimeout: 45
loadkeys: false
mounts:
  dev: /home/johndoe/.password-store-dev
  ops: /home/johndoe/.password-store-ops
  personal: /home/johndoe/secrets
  teststore: /home/johndoe/tmp/teststore
nocolor: false
noconfirm: false
path: /home/johndoe/.password-store
persistkeys: true
safecontent: false
version: 1.1.0`,
			want: &Config{
				AutoClip:      false,
				AutoImport:    false,
				ClipTimeout:   45,
				ExportKeys:    false,
				NoPager:       false,
				Notifications: false,
				Parsing:       true,
				Path:          "/home/johndoe/.password-store",
				SafeContent:   false,
				Mounts: map[string]string{
					"dev":       "/home/johndoe/.password-store-dev",
					"ops":       "/home/johndoe/.password-store-ops",
					"personal":  "/home/johndoe/secrets",
					"teststore": "/home/johndoe/tmp/teststore",
				},
			},
		}, {
			name: "1.0.0",
			cfg: `alwaystrust: false
autoimport: false
autopull: true
autopush: false
cliptimeout: 45
loadkeys: false
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/foo/.password-store
persistkeys: false
version: "1.0.0"`,
			want: &Config{
				AutoClip:      false,
				AutoImport:    false,
				ClipTimeout:   45,
				ExportKeys:    false,
				NoPager:       false,
				Notifications: false,
				Parsing:       true,
				Path:          "/home/foo/.password-store",
				SafeContent:   false,
				Mounts: map[string]string{
					"dev":       "/Users/johndoe/.password-store-dev",
					"ops":       "/Users/johndoe/.password-store-ops",
					"personal":  "/Users/johndoe/secrets",
					"teststore": "/Users/johndoe/tmp/teststore",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decode([]byte(tc.cfg), true)
			require.NoError(t, err)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("decode(%s) mismatch for:\n%s\n(-want +got):\n%s", tc.name, tc.cfg, diff)
			}
		})
	}
}

const testConfig = `root:
  askformore: true
  autoimport: true
  autosync: true
  cliptimeout: 5
  noconfirm: true
  nopager: true
  path: /home/johndoe/.password-store
  safecontent: true
mounts:
  foo/sub:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-foo-sub
    safecontent: false
  work:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-work
    safecontent: false
version: 1.4.0`

func TestLoad(t *testing.T) {
	td := os.TempDir()
	gcfg := filepath.Join(td, ".gopass.yml")
	_ = os.Remove(gcfg)
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	require.NoError(t, os.WriteFile(gcfg, []byte(testConfig), 0600))

	cfg := Load()
	assert.True(t, cfg.SafeContent)
}

func TestLoadError(t *testing.T) {
	gcfg := filepath.Join(os.TempDir(), ".gopass-err.yml")
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))

	_ = os.Remove(gcfg)

	capture(t, func() error {
		_, err := load(gcfg, false)
		if err == nil {
			return fmt.Errorf("should fail")
		}
		return nil
	})

	_ = os.Remove(gcfg)
	cfg, err := load(gcfg, false)
	assert.Error(t, err)

	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(td)
	}()
	gcfg = filepath.Join(td, "foo", ".gopass.yml")
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))
	assert.NoError(t, cfg.Save())
}

func TestDecodeError(t *testing.T) {
	gcfg := filepath.Join(os.TempDir(), ".gopass-err2.yml")
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", gcfg))

	_ = os.Remove(gcfg)
	require.NoError(t, os.WriteFile(gcfg, []byte(testConfig+"\nfoobar: zab\n"), 0600))

	capture(t, func() error {
		_, err := load(gcfg, false)
		if err == nil {
			return fmt.Errorf("should fail")
		}
		return nil
	})
}

func capture(t *testing.T, fn func() error) string {
	t.Helper()
	old := os.Stdout

	oldcol := color.NoColor
	color.NoColor = true

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		_, _ = io.Copy(buf, r)
		done <- buf.String()
	}()

	err = fn()
	// back to normal
	_ = w.Close()
	os.Stdout = old
	color.NoColor = oldcol
	assert.NoError(t, err)
	out := <-done
	return strings.TrimSpace(out)
}
