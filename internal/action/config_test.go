package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	t.Run("display config", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t)
		assert.NoError(t, act.Config(c))
		want := `autoclip: true
autoimport: true
cliptimeout: 45
exportkeys: true
nopager: false
notifications: true
parsing: true
`
		want += "path: " + u.StoreDir("") + "\n"
		want += `safecontent: false
`
		assert.Equal(t, want, buf.String())
	})

	t.Run("set valid config value", func(t *testing.T) {
		defer buf.Reset()

		assert.NoError(t, act.setConfigValue(ctx, "nopager", "true"))
		assert.Equal(t, "nopager: true", strings.TrimSpace(buf.String()), "action.setConfigValue")
	})

	t.Run("set invalid config value", func(t *testing.T) {
		defer buf.Reset()

		assert.Error(t, act.setConfigValue(ctx, "foobar", "true"))
	})

	t.Run("print single config value", func(t *testing.T) {
		defer buf.Reset()

		act.printConfigValues(ctx, "nopager")

		want := "nopager: true"
		assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	})

	t.Run("print all config values", func(t *testing.T) {
		defer buf.Reset()

		act.printConfigValues(ctx)
		want := `autoclip: true
autoimport: true
cliptimeout: 45
exportkeys: true
nopager: true
notifications: true
parsing: true
`
		want += "path: " + u.StoreDir("") + "\n"
		want += `safecontent: false`
		assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")

		delete(act.cfg.Mounts, "foo")
	})

	t.Run("show autoimport value", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "autoimport")
		assert.NoError(t, act.Config(c))
		assert.Equal(t, "autoimport: true", strings.TrimSpace(buf.String()))
	})

	t.Run("disable autoimport", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "autoimport", "false")
		assert.NoError(t, act.Config(c))
		assert.Equal(t, "autoimport: false", strings.TrimSpace(buf.String()))
	})

	t.Run("complete config items", func(t *testing.T) {
		defer buf.Reset()

		act.ConfigComplete(gptest.CliCtx(ctx, t))
		want := `autoclip
autoimport
cliptimeout
exportkeys
nopager
notifications
parsing
path
remote
safecontent
`
		assert.Equal(t, want, buf.String())
	})

	t.Run("set autoimport to invalid value", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "autoimport", "false", "42")
		assert.Error(t, act.Config(c))
	})
}
