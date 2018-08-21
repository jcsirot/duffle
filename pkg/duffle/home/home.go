package home

import (
	"os"
	"path/filepath"
	"runtime"
)

const HomeEnvVar = "DUFFLE_HOME"
const PluginEnvVar = `DUFFLE_PLUGIN`

// Home describes the location of a CLI configuration.
//
// This helper builds paths relative to a Duffle Home directory.
type Home string

// Path returns Home with elements appended.
func (h Home) Path(elem ...string) string {
	p := []string{h.String()}
	p = append(p, elem...)
	return filepath.Join(p...)
}

// Config returns the path to the Duffle config file.
func (h Home) Config() string {
	return h.Path("config.toml")
}

// Logs returns the path to the Duffle logs.
func (h Home) Logs() string {
	return h.Path("logs")
}

// Plugins returns the path to the Duffle plugins.
func (h Home) Plugins() string {
	plugdirs := os.Getenv(PluginEnvVar)

	if plugdirs == "" {
		plugdirs = h.Path("plugins")
	}

	return plugdirs
}

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return string(h)
}

// DefaultHome gives the default value for $(duffle home)
func DefaultHome() string {
	if home := os.Getenv(HomeEnvVar); home != "" {
		return home
	}

	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".duffle")
}