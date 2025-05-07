package ust

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-ini/ini"
)

// Decode decodes a UST file.
func Decode(r io.Reader) (*File, error) {
	file := &File{}

	var err error
	loadopts := ini.LoadOptions{
		UnparseableSections: []string{"#VERSION"},
	}
	file.iniFile, err = ini.LoadSources(loadopts, r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse INI: %w", err)
	}

	for _, sec := range file.iniFile.Sections() {
		if sec.Name() == "DEFAULT" {
			continue
		}
		if sec.Name() == "#TRACKEND" {
			break
		}
		if sec.Name() == "#VERSION" {
			if err := file.parseVersion(sec); err != nil {
				return nil, fmt.Errorf("failed to parse version section: %w", err)
			}
		}
		if sec.Name() == "#SETTING" {
			if err := file.parseSetting(sec); err != nil {
				return nil, fmt.Errorf("failed to parse settings: %w", err)
			}
		}
	}

	return file, nil
}

func (f *File) parseVersion(sec *ini.Section) error {
	f.Version = Version1_2
	ver, err := ParseVersion(strings.TrimSpace(sec.Body()))
	if err != nil {
		return err
	}
	f.Version = ver
	return nil
}

func (f *File) parseSetting(sec *ini.Section) (err error) {
	f.Settings = Settings{}

	// Tempo
	f.Settings.Tempo, err = sec.Key("Tempo").Float64()
	if err != nil {
		return fmt.Errorf("failed to parse tempo: %w", err)
	}

	// ProjectName
	f.Settings.ProjectName = sec.Key("ProjectName").String()

	// Project (OpenUtau)
	f.Settings.Project = sec.Key("Project").String()

	// VoiceDir
	f.Settings.VoiceDir = sec.Key("VoiceDir").String()

	// OutFile
	f.Settings.OutFile = sec.Key("OutFile").String()

	// CacheDir
	f.Settings.CacheDir = sec.Key("CacheDir").String()

	// Tool1
	f.Settings.Tool1 = sec.Key("Tool1").String()

	// Tool2
	f.Settings.Tool2 = sec.Key("Tool2").String()

	// Mode2
	f.Settings.Mode2, err = sec.Key("Mode2").Bool()
	if err != nil {
		return fmt.Errorf("failed to parse Mode2: %w", err)
	}

	return nil
}
