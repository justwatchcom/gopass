package config

import (
	"net/url"
	"strings"
)

// Pre1127 is a pre-1.12.7 config
type Pre1127 struct {
	AutoClip      bool              `yaml:"autoclip"`      // decide whether passwords are automatically copied or not
	AutoImport    bool              `yaml:"autoimport"`    // import missing public keys w/o asking
	ClipTimeout   int               `yaml:"cliptimeout"`   // clear clipboard after seconds
	ExportKeys    bool              `yaml:"exportkeys"`    // automatically export public keys of all recipients
	NoColor       bool              `yaml:"nocolor"`       // do not use color when outputing text
	NoPager       bool              `yaml:"nopager"`       // do not invoke a pager to display long lists
	Notifications bool              `yaml:"notifications"` // enable desktop notifications
	Parsing       bool              `yaml:"parsing"`       // allows to switch off all output parsing
	Path          string            `yaml:"path"`
	SafeContent   bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	Mounts        map[string]string `yaml:"mounts"`

	ConfigPath string `yaml:"-"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// Config converts the Pre1127 config to the current config struct
func (c *Pre1127) Config() *Config {
	cfg := &Config{
		AutoClip:      c.AutoClip,
		AutoImport:    c.AutoImport,
		ClipTimeout:   c.ClipTimeout,
		ExportKeys:    c.ExportKeys,
		NoPager:       c.NoPager,
		Notifications: c.Notifications,
		Parsing:       true,
		Path:          c.Path,
		SafeContent:   c.SafeContent,
		Mounts:        make(map[string]string, len(c.Mounts)),
	}
	for k, v := range c.Mounts {
		cfg.Mounts[k] = v
	}
	return cfg
}

// CheckOverflow implements configer
func (c *Pre1127) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Pre1102 is a pre-1.10.2 config
type Pre1102 struct {
	AutoClip      bool              `yaml:"autoclip"`      // decide whether passwords are automatically copied or not
	AutoImport    bool              `yaml:"autoimport"`    // import missing public keys w/o asking
	ClipTimeout   int               `yaml:"cliptimeout"`   // clear clipboard after seconds
	ExportKeys    bool              `yaml:"exportkeys"`    // automatically export public keys of all recipients
	MIME          bool              `yaml:"mime"`          // enable gopass native MIME secrets
	NoColor       bool              `yaml:"nocolor"`       // do not use color when outputing text
	NoPager       bool              `yaml:"nopager"`       // do not invoke a pager to display long lists
	Notifications bool              `yaml:"notifications"` // enable desktop notifications
	Path          string            `yaml:"path"`
	SafeContent   bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	Mounts        map[string]string `yaml:"mounts"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// CheckOverflow implements configer
func (c *Pre1102) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Config converts the Pre1102 config to the current config struct
func (c *Pre1102) Config() *Config {
	cfg := &Config{
		AutoClip:      c.AutoClip,
		AutoImport:    c.AutoImport,
		ClipTimeout:   c.ClipTimeout,
		ExportKeys:    c.ExportKeys,
		NoPager:       c.NoPager,
		Notifications: c.Notifications,
		Parsing:       true,
		Path:          c.Path,
		SafeContent:   c.SafeContent,
		Mounts:        make(map[string]string, len(c.Mounts)),
	}
	for k, v := range c.Mounts {
		cfg.Mounts[k] = v
	}
	return cfg
}

// Pre193 is is pre-1.9.3 config
type Pre193 struct {
	Path   string `yaml:"-"`
	Root   *Pre193StoreConfig
	Mounts map[string]*Pre193StoreConfig

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// Pre193StoreConfig is a pre-1.9.3 store config
type Pre193StoreConfig struct {
	AutoClip       bool              `yaml:"autoclip"`   // decide whether passwords are automatically copied or not
	AutoImport     bool              `yaml:"autoimport"` // import missing public keys w/o asking
	AutoSync       bool              `yaml:"autosync"`   // push to git remote after commit, pull before push if necessary
	CheckRecpHash  bool              `yaml:"check_recipient_hash"`
	ClipTimeout    int               `yaml:"cliptimeout"`    // clear clipboard after seconds
	Concurrency    int               `yaml:"concurrency"`    // allow to run multiple thread when batch processing
	EditRecipients bool              `yaml:"editrecipients"` // edit recipients when confirming
	ExportKeys     bool              `yaml:"exportkeys"`     // automatically export public keys of all recipients
	NoColor        bool              `yaml:"nocolor"`        // do not use color when outputing text
	Confirm        bool              `yaml:"noconfirm"`      // do not confirm recipients when encrypting
	NoPager        bool              `yaml:"nopager"`        // do not invoke a pager to display long lists
	Notifications  bool              `yaml:"notifications"`  // enable desktop notifications
	Path           string            `yaml:"path"`           // path to the root store
	RecipientHash  map[string]string `yaml:"recipient_hash"`
	SafeContent    bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	UseSymbols     bool              `yaml:"usesymbols"`  // always use symbols when generating passwords
}

// CheckOverflow implements configer
func (c *Pre193) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Config converts the Pre193 config to the current config struct
func (c *Pre193) Config() *Config {
	cfg := &Config{
		AutoClip:      c.Root.AutoClip,
		AutoImport:    c.Root.AutoImport,
		ClipTimeout:   c.Root.ClipTimeout,
		NoPager:       c.Root.NoPager,
		Notifications: c.Root.Notifications,
		Parsing:       true,
		Path:          c.Root.Path,
		SafeContent:   c.Root.SafeContent,
		Mounts:        make(map[string]string, len(c.Mounts)),
	}
	if p, err := pathFromURL(c.Root.Path); err == nil {
		cfg.Path = p
	}
	for k, v := range c.Mounts {
		p, err := pathFromURL(v.Path)
		if err != nil {
			continue
		}
		cfg.Mounts[k] = p
	}
	return cfg
}

// Pre182 is the gopass config structure before version 1.8.2
type Pre182 struct {
	Path    string                        `yaml:"-"`
	Root    *Pre182StoreConfig            `yaml:"root"`
	Mounts  map[string]*Pre182StoreConfig `yaml:"mounts"`
	Version string                        `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// Pre182StoreConfig is a per-store (root or mount) config
type Pre182StoreConfig struct {
	AskForMore     bool              `yaml:"askformore"` // ask for more data on generate
	AutoClip       bool              `yaml:"autoclip"`   // decide whether passwords are automatically copied or not
	AutoImport     bool              `yaml:"autoimport"` // import missing public keys w/o asking
	AutoSync       bool              `yaml:"autosync"`   // push to git remote after commit, pull before push if necessary
	CheckRecpHash  bool              `yaml:"check_recipient_hash"`
	ClipTimeout    int               `yaml:"cliptimeout"`    // clear clipboard after seconds
	Concurrency    int               `yaml:"concurrency"`    // allow to run multiple thread when batch processing
	EditRecipients bool              `yaml:"editrecipients"` // edit recipients when confirming
	NoColor        bool              `yaml:"nocolor"`        // do not use color when outputing text
	Confirm        bool              `yaml:"noconfirm"`      // do not confirm recipients when encrypting
	NoPager        bool              `yaml:"nopager"`        // do not invoke a pager to display long lists
	Notifications  bool              `yaml:"notifications"`  // enable desktop notifications
	Path           string            `yaml:"path"`           // path to the root store
	RecipientHash  map[string]string `yaml:"recipient_hash"`
	SafeContent    bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	UseSymbols     bool              `yaml:"usesymbols"`  // always use symbols when generating passwords
}

// CheckOverflow implements configer
func (c *Pre182) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Config converts the Pre182 config to the current config struct
func (c *Pre182) Config() *Config {
	cfg := &Config{
		AutoClip:      c.Root.AutoClip,
		AutoImport:    c.Root.AutoImport,
		ClipTimeout:   c.Root.ClipTimeout,
		NoPager:       c.Root.NoPager,
		Notifications: c.Root.Notifications,
		Parsing:       true,
		Path:          c.Root.Path,
		SafeContent:   c.Root.SafeContent,
		Mounts:        make(map[string]string, len(c.Mounts)),
	}
	if p, err := pathFromURL(c.Root.Path); err == nil {
		cfg.Path = p
	}
	for k, v := range c.Mounts {
		p, err := pathFromURL(v.Path)
		if err != nil {
			continue
		}
		cfg.Mounts[k] = p
	}
	return cfg
}

// Pre140 is the gopass config structure before version 1.4.0
type Pre140 struct {
	AskForMore  bool              `yaml:"askformore"`  // ask for more data on generate
	AutoImport  bool              `yaml:"autoimport"`  // import missing public keys w/o asking
	AutoSync    bool              `yaml:"autosync"`    // push to git remote after commit, pull before push if necessary
	ClipTimeout int               `yaml:"cliptimeout"` // clear clipboard after seconds
	Mounts      map[string]string `yaml:"mounts,omitempty"`
	Confirm     bool              `yaml:"noconfirm"`   // do not confirm recipients when encrypting
	Path        string            `yaml:"path"`        // path to the root store
	SafeContent bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	Version     string            `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// CheckOverflow implements configer
func (c *Pre140) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Config converts the Pre140 config to the current config struct
func (c *Pre140) Config() *Config {
	cfg := &Config{
		AutoImport:  c.AutoImport,
		ClipTimeout: c.ClipTimeout,
		Parsing:     true,
		Path:        c.Path,
		SafeContent: c.SafeContent,
		Mounts:      make(map[string]string, len(c.Mounts)),
	}
	for k, v := range c.Mounts {
		cfg.Mounts[k] = v
	}
	return cfg
}

// Pre130 is the gopass config structure before version 1.3.0. Not all fields were
// available between 1.0.0 and 1.3.0, but this struct should cover all of them.
type Pre130 struct {
	AlwaysTrust bool              `yaml:"alwaystrust"` // always trust public keys when encrypting
	AskForMore  bool              `yaml:"askformore"`  // ask for more data on generate
	AutoImport  bool              `yaml:"autoimport"`  // import missing public keys w/o asking
	AutoPull    bool              `yaml:"autopull"`    // pull from git before push
	AutoPush    bool              `yaml:"autopush"`    // push to git remote after commit
	ClipTimeout int               `yaml:"cliptimeout"` // clear clipboard after seconds
	Debug       bool              `yaml:"debug"`       // enable debug output
	LoadKeys    bool              `yaml:"loadkeys"`    // load missing keys from store
	Mounts      map[string]string `yaml:"mounts,omitempty"`
	NoColor     bool              `yaml:"nocolor"`     // disable colors in output
	Confirm     bool              `yaml:"noconfirm"`   // do not confirm recipients when encrypting
	Path        string            `yaml:"path"`        // path to the root store
	PersistKeys bool              `yaml:"persistkeys"` // store recipient keys in store
	SafeContent bool              `yaml:"safecontent"` // avoid showing passwords in terminal
	Version     string            `yaml:"version"`

	// Catches all undefined files and must be empty after parsing
	XXX map[string]interface{} `yaml:",inline"`
}

// CheckOverflow implements configer
func (c *Pre130) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Config converts the Pre130 config to the current config struct
func (c *Pre130) Config() *Config {
	cfg := &Config{
		AutoImport:  c.AutoImport,
		ClipTimeout: c.ClipTimeout,
		Parsing:     true,
		Path:        c.Path,
		SafeContent: c.SafeContent,
		Mounts:      make(map[string]string, len(c.Mounts)),
	}
	for k, v := range c.Mounts {
		cfg.Mounts[k] = v
	}
	return cfg
}

func pathFromURL(u string) (string, error) {
	if !strings.Contains(u, "://") {
		return u, nil
	}

	up, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return up.Path, nil
}
