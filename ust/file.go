package ust

import (
	"github.com/go-ini/ini"
)

// File represents a parsed UST file.
type File struct {
	Version  Version  // Version is the UST file format version.
	Settings Settings // Settings represents the settings of the UST file.
	Notes    []Note   // Notes holds the notes.

	iniFile *ini.File // iniFile holds the raw parsed INI file structure (used internally).
}
