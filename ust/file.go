package ust

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-ini/ini"
)

// File represents a parsed UST file.
type File struct {
	Version Version
}

// Decode decodes a UST file.
func Decode(r io.Reader) (*File, error) {
	file := &File{}

	loadopts := ini.LoadOptions{
		UnparseableSections: []string{"#VERSION"},
	}
	inifile, err := ini.LoadSources(loadopts, r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse INI: %w", err)
	}

	// Version
	file.Version = Version1_2
	if sec, err := inifile.GetSection("#VERSION"); err == nil {
		ver, err := ParseVersion(strings.TrimSpace(sec.Body()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse version: %w", err)
		}
		file.Version = ver
	}

	return file, nil
}
