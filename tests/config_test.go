package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestBaseConfig(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("config")
	assert.NoError(t, err)
	wanted := `autoclip: false
autoimport: true
cliptimeout: 45
exportkeys: false
nopager: false
notifications: false
parsing: true
`
	wanted += "path: " + ts.storeDir("root") + "\n"
	wanted += "safecontent: false"

	assert.Equal(t, wanted, out)

	invertables := []string{
		"autoimport",
		"safecontent",
		"parsing",
	}

	for _, invert := range invertables {
		t.Run("invert "+invert, func(t *testing.T) {
			out, err = ts.run("config " + invert + " false")
			assert.NoError(t, err)
			assert.Equal(t, invert+": false", out)

			out, err = ts.run("config " + invert)
			assert.NoError(t, err)
			assert.Equal(t, invert+": false", out)
		})
	}

	t.Run("cliptimeout", func(t *testing.T) {
		out, err = ts.run("config cliptimeout 120")
		assert.NoError(t, err)
		assert.Equal(t, "cliptimeout: 120", out)

		out, err = ts.run("config cliptimeout")
		assert.NoError(t, err)
		assert.Equal(t, "cliptimeout: 120", out)
	})
}

func TestMountConfig(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// we add a mount:
	_, err := ts.run("init --store mnt/m1 --path " + ts.storeDir("m1") + " --storage=fs " + keyID)
	require.NoError(t, err)

	_, err = ts.run("config")
	assert.NoError(t, err)

	wanted := `autoclip: false
autoimport: true
cliptimeout: 45
exportkeys: false
nopager: false
notifications: false
parsing: true
path: `
	wanted += ts.storeDir("root") + "\n"
	wanted += `safecontent: false
mount "mnt/m1" => "`
	wanted += ts.storeDir("m1") + "\"\n"

	out, err := ts.run("config")
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(wanted), out)
}
