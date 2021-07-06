package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/tests/gptest"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/stretchr/testify/require"
)

const (
	gopassConfig = `autoclip: false
autoimport: true
cliptimeout: 45
exportkeys: false
parsing: true
`
	keyID = "BE73F104"
)

type tester struct {
	t *testing.T

	// Binary is the path to the gopass binary used for testing
	Binary    string
	sourceDir string
	tempDir   string
	resetFn   func()
}

func newTester(t *testing.T) *tester {
	t.Helper()
	sourceDir := "."
	if d := os.Getenv("GOPASS_TEST_DIR"); d != "" {
		sourceDir = d
	}

	gopassBin := ""
	if b := os.Getenv("GOPASS_BINARY"); b != "" {
		gopassBin = b
	}
	fi, err := os.Stat(gopassBin)
	if err != nil {
		t.Skipf("Failed to stat GOPASS_BINARY %s: %s", gopassBin, err)
	}
	if !strings.HasSuffix(gopassBin, ".exe") && fi.Mode()&0111 == 0 {
		t.Fatalf("GOPASS_BINARY is not executeable")
	}
	t.Logf("Using gopass binary: %s", gopassBin)

	ts := &tester{
		t:         t,
		sourceDir: sourceDir,
		Binary:    gopassBin,
	}
	// create tempDir
	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	t.Logf("Tempdir: %s", td)
	ts.tempDir = td

	// prepare ENVIRONMENT
	ts.resetFn = gptest.UnsetVars("GNUPGHOME", "GOPASS_DEBUG", "NO_COLOR", "GOPASS_CONFIG", "GOPASS_NO_NOTIFY", "GOPASS_HOMEDIR")
	require.NoError(t, os.Setenv("GNUPGHOME", ts.gpgDir()))
	require.NoError(t, os.Setenv("GOPASS_DEBUG", ""))
	require.NoError(t, os.Setenv("NO_COLOR", "true"))
	require.NoError(t, os.Setenv("GOPASS_CONFIG", ts.gopassConfig()))
	require.NoError(t, os.Setenv("GOPASS_NO_NOTIFY", "true"))
	require.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	// write config
	require.NoError(t, os.MkdirAll(filepath.Dir(ts.gopassConfig()), 0700))
	// we need to set the root path to something else than the root directory otherwise the mounts will show as regular entries
	if err := os.WriteFile(ts.gopassConfig(), []byte(gopassConfig+"\npath: "+ts.storeDir("root")+"\n"), 0600); err != nil {
		t.Fatalf("Failed to write gopass config to %s: %s", ts.gopassConfig(), err)
	}

	// copy gpg test files
	files := map[string]string{
		filepath.Join(ts.sourceDir, "can", "gnupg", "pubring.gpg"): filepath.Join(ts.gpgDir(), "pubring.gpg"),
		filepath.Join(ts.sourceDir, "can", "gnupg", "random_seed"): filepath.Join(ts.gpgDir(), "random_seed"),
		filepath.Join(ts.sourceDir, "can", "gnupg", "secring.gpg"): filepath.Join(ts.gpgDir(), "secring.gpg"),
		filepath.Join(ts.sourceDir, "can", "gnupg", "trustdb.gpg"): filepath.Join(ts.gpgDir(), "trustdb.gpg"),
	}
	for from, to := range files {
		buf, err := os.ReadFile(from)
		require.NoError(t, err, "Failed to read file %s", from)

		err = os.MkdirAll(filepath.Dir(to), 0700)
		require.NoError(t, err, "Failed to create dir for %s", to)

		err = os.WriteFile(to, buf, 0600)
		require.NoError(t, err, "Failed to write file %s", to)
	}

	return ts
}

func (ts tester) gpgDir() string {
	return filepath.Join(ts.tempDir, ".gnupg")
}

func (ts tester) gopassConfig() string {
	return filepath.Join(ts.tempDir, ".config", "gopass", "config.yml")
}

func (ts tester) storeDir(mount string) string {
	return filepath.Join(ts.tempDir, ".local", "share", "gopass", "stores", mount)
}

func (ts tester) workDir() string {
	return filepath.Dir(ts.tempDir)
}

func (ts tester) teardown() {
	ts.resetFn() // restore env vars
	if ts.tempDir == "" {
		return
	}
	err := os.RemoveAll(ts.tempDir)
	require.NoError(ts.t, err)
}

func (ts tester) runCmd(args []string, in []byte) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no command")
	}

	cmd := exec.CommandContext(context.Background(), args[0], args[1:]...)
	cmd.Dir = ts.workDir()
	cmd.Stdin = bytes.NewReader(in)

	ts.t.Logf("%+v", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return strings.TrimSpace(string(out)), nil
}

func (ts tester) run(arg string) (string, error) {
	if runtime.GOOS == "windows" {
		arg = strings.Replace(arg, "\\", "\\\\", -1)
	}
	args, err := shellquote.Split(arg)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(context.Background(), ts.Binary, args...)
	cmd.Dir = ts.workDir()

	ts.t.Logf("%+v", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return strings.TrimSpace(string(out)), nil
}

func (ts tester) runWithInput(arg, input string) ([]byte, error) {
	reader := strings.NewReader(input)
	return ts.runWithInputReader(arg, reader)
}

func (ts tester) runWithInputReader(arg string, input io.Reader) ([]byte, error) {
	args, err := shellquote.Split(arg)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ts.Binary, args...)
	cmd.Dir = ts.workDir()
	cmd.Stdin = input

	ts.t.Logf("%+v", cmd.Args)

	return cmd.CombinedOutput()
}

func (ts *tester) initStore() {
	out, err := ts.run("init --crypto=gpgcli --storage=fs " + keyID)
	require.NoError(ts.t, err, "failed to init password store:\n%s", out)
}

func (ts *tester) initSecrets(prefix string) {
	out, err := ts.run("generate -p " + prefix + "foo/bar 20")
	require.NoError(ts.t, err, "failed to generate password:\n%s", out)

	out, err = ts.run("generate -p " + prefix + "baz 40")
	require.NoError(ts.t, err, "failed to generate password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", prefix + "fixed/secret"}, []byte("moar"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", prefix + "fixed/twoliner"}, []byte("and\nmore stuff"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)
}
