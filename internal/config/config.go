package config

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

var (
	// ErrConfigNotFound is returned on load if the config was not found
	ErrConfigNotFound = fmt.Errorf("config not found")
	// ErrConfigNotParsed is returned on load if the config could not be decoded
	ErrConfigNotParsed = fmt.Errorf("config not parseable")
)

// Config is the current config struct
type Config struct {
	AutoClip      bool              `yaml:"autoclip"`      // decide whether passwords are automatically copied or not
	AutoImport    bool              `yaml:"autoimport"`    // import missing public keys w/o asking
	ClipTimeout   int               `yaml:"cliptimeout"`   // clear clipboard after seconds
	ExportKeys    bool              `yaml:"exportkeys"`    // automatically export public keys of all recipients
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

// New creates a new config with sane default values
func New() *Config {
	return &Config{
		AutoImport:    true,
		ClipTimeout:   45,
		ExportKeys:    true,
		Mounts:        make(map[string]string),
		Notifications: true,
		Parsing:       true,
		Path:          PwStoreDir(""),
		ConfigPath:    configLocation(),
	}
}

// CheckOverflow implements configer. It will check for any extra config values not
// handled by the current struct
func (c *Config) CheckOverflow() error {
	return checkOverflow(c.XXX)
}

// Config will return a current config
func (c *Config) Config() *Config {
	return c
}

// SetConfigValue will try to set the given key to the value in the config struct
func (c *Config) SetConfigValue(key, value string) error {
	if err := c.setConfigValue(key, value); err != nil {
		return err
	}
	return c.Save()
}

// setConfigValue will try to set the given key to the value in the config struct
func (c *Config) setConfigValue(key, value string) error {
	value = strings.ToLower(value)
	o := reflect.ValueOf(c).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("yaml")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		if jsonArg != key {
			continue
		}
		f := o.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.SetString(value)
			return nil
		case reflect.Bool:
			if value == "true" || value == "on" {
				f.SetBool(true)
				return nil
			} else if value == "false" || value == "off" {
				f.SetBool(false)
				return nil
			} else {
				return fmt.Errorf("not a bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("failed to convert %q to integer: %w", value, err)
			}
			f.SetInt(int64(iv))
			return nil
		default:
			continue
		}
	}
	return fmt.Errorf("unknown config option %q", key)
}

func (c *Config) String() string {
	return fmt.Sprintf("%#v", c)
}

// Directory returns the directory this config is using
func (c *Config) Directory() string {
	return filepath.Dir(c.Path)
}

// ConfigMap returns a map of stringified config values for easy printing
func (c *Config) ConfigMap() map[string]string {
	m := make(map[string]string, 20)
	o := reflect.ValueOf(c).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("yaml")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		f := o.Field(i)
		var strVal string
		switch f.Kind() {
		case reflect.String:
			strVal = f.String()
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", f.Bool())
		case reflect.Int:
			strVal = fmt.Sprintf("%d", f.Int())
		default:
			continue
		}
		m[jsonArg] = strVal
	}
	return m
}
