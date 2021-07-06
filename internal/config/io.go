package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"

	"gopkg.in/yaml.v3"
)

// LoadWithFallbackRelaxed will try to load the config from one of the default
// locations but also accept a more recent config.
func LoadWithFallbackRelaxed() *Config {
	return loadWithFallback(true)
}

// LoadWithFallback will try to load the config from one of the default locations
func LoadWithFallback() *Config {
	return loadWithFallback(false)
}

func loadWithFallback(relaxed bool) *Config {
	for _, l := range configLocations() {
		if cfg := loadConfig(l, relaxed); cfg != nil {
			return cfg
		}
	}
	return loadDefault()
}

// Load will load the config from the default location or return a default config
func Load() *Config {
	if cfg := loadConfig(configLocation(), false); cfg != nil {
		return cfg
	}
	return loadDefault()
}

func loadConfig(l string, relaxed bool) *Config {
	debug.Log("Trying to load config from %s", l)
	cfg, err := load(l, relaxed)
	if err == ErrConfigNotFound {
		return nil
	}
	if err != nil {
		return nil
	}
	debug.Log("Loaded config from %s: %+v", l, cfg)
	return cfg
}

func loadDefault() *Config {
	cfg := New()
	cfg.Path = PwStoreDir("")
	debug.Log("Loaded default config: %+v", cfg)
	return cfg
}

func load(cf string, relaxed bool) (*Config, error) {
	// deliberately using os.Stat here, a symlinked
	// config is OK
	if _, err := os.Stat(cf); err != nil {
		return nil, ErrConfigNotFound
	}
	buf, err := os.ReadFile(cf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotFound
	}

	cfg, err := decode(buf, relaxed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config from %s: %s\n", cf, err)
		return nil, ErrConfigNotParsed
	}
	if cfg.Mounts == nil {
		cfg.Mounts = make(map[string]string)
	}
	cfg.ConfigPath = cf
	return cfg, nil
}

func checkOverflow(m map[string]interface{}) error {
	if len(m) < 1 {
		return nil
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Errorf("unknown fields: %+v", keys)
}

type configer interface {
	Config() *Config
	CheckOverflow() error
}

func decode(buf []byte, relaxed bool) (*Config, error) {
	mostRecent := &Config{
		ExportKeys: true,
		Parsing:    true,
	}
	cfgs := []configer{
		// most recent config must come first
		mostRecent,
		&Pre1127{},
		&Pre1102{},
		&Pre193{
			Root: &Pre193StoreConfig{},
		},
		&Pre182{
			Root: &Pre182StoreConfig{},
		},
		&Pre140{},
		&Pre130{},
	}
	if relaxed {
		// most recent config must come last as well, will be tried w/o
		// overflow checks
		cfgs = append(cfgs, mostRecent)
	}
	var warn string
	for i, cfg := range cfgs {
		debug.Log("Trying to unmarshal config into %T", cfg)
		if err := yaml.Unmarshal(buf, cfg); err != nil {
			debug.Log("Loading config %T failed: %s", cfg, err)
			continue
		}
		if err := cfg.CheckOverflow(); err != nil {
			debug.Log("Extra elements in config: %s", err)
			if i == 0 {
				warn = fmt.Sprintf("Failed to load config %T. Do you need to remove deprecated fields? %s\n", cfg, err)
			}
			// usually we are strict about extra fields, i.e. any field left
			// unparsed means this config failed and we try the next one.
			if i < len(cfgs)-1 {
				continue
			}
			// in relaxed mode we append an extra copy of the most recent
			// config to the end of the slice and might just ignore these
			// extra fields.
			if !relaxed {
				continue
			}
			debug.Log("Ignoring extra config fields for fallback config (only)")
		}
		debug.Log("Loaded config: %T: %+v", cfg, cfg)
		conf := cfg.Config()
		if i > 0 {
			debug.Log("Loaded legacy config. Should rewrite config.")
		}
		return conf, nil
	}
	// We try to provide a seamless config upgrade path for users of our
	// released versions. But some users build gopass from the master branch
	// and these might run into issues when we remove config options.
	// Since our config parser is pedantic (it has to) we fail parsing on
	// unknown config options. If we remove one and the user rebuilds it's
	// gopass binary without changing the config, it will fail to parse the
	// current config and the legacy configs will likely fail as well.
	// But if we always display the warning users with configs from previous
	// releases will always see the warning. So instead we only display the
	// warning if parsing of the most up to date config struct fails AND
	// not other succeeds.
	if warn != "" {
		fmt.Fprint(os.Stderr, warn)
	}
	return nil, ErrConfigNotParsed
}

// Save saves the config
func (c *Config) Save() error {
	buf, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	cfgLoc := configLocation()
	cfgDir := filepath.Dir(cfgLoc)
	if !fsutil.IsDir(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0700); err != nil {
			return fmt.Errorf("failed to create dir %q: %w", cfgDir, err)
		}
	}
	if err := os.WriteFile(cfgLoc, buf, 0600); err != nil {
		return fmt.Errorf("failed to write config file to %q: %w", cfgLoc, err)
	}
	debug.Log("Saved config to %s: %+v\n", cfgLoc, c)
	return nil
}
