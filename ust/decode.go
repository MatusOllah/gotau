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

	// Version
	if err := file.parseVersion(); err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	switch file.Version {
	case Version1_2:
		if err := file.parseUST1_2(); err != nil {
			return nil, err
		}
	case Version2_0:
		if err := file.parseUST2_0(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported version: %s", file.Version.String())
	}

	return file, nil
}

func (f *File) parseVersion() error {
	f.Version = Version1_2
	if sec, err := f.iniFile.GetSection("#VERSION"); err == nil {
		ver, err := ParseVersion(strings.TrimSpace(sec.Body()))
		if err != nil {
			return err
		}
		f.Version = ver
	}
	return nil
}

func (f *File) parseUST1_2() error {
	// Settings
	if err := f.parseSetting(); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	return nil
}

func (f *File) parseUST2_0() error {
	// Settings
	if err := f.parseSetting(); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	return nil
}

func (f *File) parseSetting() error {
	f.Settings = &Settings{}

	sec, err := f.iniFile.GetSection("#SETTING")
	if err != nil {
		return err
	}

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
