package ust

import (
	"github.com/go-ini/ini"
)

// File represents a parsed UST file.
type File struct {
	Version  Version
	Settings *Settings

	iniFile *ini.File
}
