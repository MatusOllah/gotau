// Package voicebank provides support for reading UTAU voicebanks.
package voicebank

import (
	"fmt"
	"io/fs"

	"github.com/go-ini/ini"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

type InstallInfo struct {
	Type        string
	Folder      string
	ContentsDir string
	Description string
}

// Voicebank represents an UTAU voicebank.
type Voicebank struct {
	// Oto is the voicebank's oto.ini configuration.
	Oto Oto

	// InstallInfo is the voicebank's install information.
	// It is only valid if the voicebank is an installer voicebank.
	InstallInfo *InstallInfo
}

type voicebankConfig struct {
	fileEncoding encoding.Encoding
}

// using a "universal" Option type here instead of something like OpenOption just in case
// if we want to add more functions beyond just reading voicebanks in the future.
// also voicebank.Option sounds better than voicebank.VoicebankOption imo and
// it still distinguishes from voicebank.OtoOption well enough :3

// Option represents an option for passing into voicebank-related functions and methods.
type Option func(*voicebankConfig)

// WithFileEncoding specifies the character encoding to use when reading or writing voicebank files.
func WithFileEncoding(encoding encoding.Encoding) Option {
	return func(cfg *voicebankConfig) {
		cfg.fileEncoding = encoding
	}
}

// Open opens an UTAU voicebank from the given filesystem.
func Open(fsys fs.FS, opts ...Option) (*Voicebank, error) {
	cfg := &voicebankConfig{
		fileEncoding: encoding.Nop,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if fileExists(fsys, "install.txt") {
		info, err := parseInstallInfo(fsys, cfg.fileEncoding)
		if err != nil {
			return nil, fmt.Errorf("voicebank: failed to parse install.txt: %w", err)
		}

		vbFsys, err := fs.Sub(fsys, info.Folder)
		if err != nil {
			return nil, fmt.Errorf("voicebank: failed to access voicebank folder %q: %w", info.Folder, err)
		}

		vb, err := openNonInstaller(vbFsys, cfg)
		if err != nil {
			return nil, err
		}
		vb.InstallInfo = info

		return vb, nil
	}
	return openNonInstaller(fsys, cfg)
}

func openNonInstaller(fsys fs.FS, cfg *voicebankConfig) (*Voicebank, error) {
	//TODO: this
	return &Voicebank{}, nil
}

func parseInstallInfo(fsys fs.FS, enc encoding.Encoding) (*InstallInfo, error) {
	f, err := fsys.Open("install.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	newReader := transform.NewReader(f, enc.NewDecoder())

	iniFile, err := ini.Load(newReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ini: %w", err)
	}

	return &InstallInfo{
		Type:        iniFile.Section("DEFAULT").Key("type").String(),
		Folder:      iniFile.Section("DEFAULT").Key("folder").String(),
		ContentsDir: iniFile.Section("DEFAULT").Key("contentsdir").String(),
		Description: iniFile.Section("DEFAULT").Key("description").String(),
	}, nil
}

func fileExists(fsys fs.FS, name string) bool {
	_, err := fs.Stat(fsys, name)
	return err == nil
}
