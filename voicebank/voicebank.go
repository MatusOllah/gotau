// Package voicebank provides support for reading UTAU voicebanks.
package voicebank

import (
	"bufio"
	"fmt"
	"image"
	"io/fs"
	"strings"

	_ "image/png"

	"github.com/MatusOllah/resona/codec"
	_ "github.com/MatusOllah/resona/codec/au"
	_ "github.com/MatusOllah/resona/codec/qoa"
	_ "github.com/MatusOllah/resona/codec/wav"
	"github.com/go-ini/ini"
	_ "golang.org/x/image/bmp"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// InstallInfo represents metadata about installer voicebanks (i.e. install.txt).
type InstallInfo struct {
	// Type specifies the installer type (e.g. voiceset).
	Type string

	// Folder specifies the target directory where the voicebank will be installed,
	// i.e., the folder name that will appear inside UTAU's `voice/` directory.
	Folder string

	// ContentsDir specifies the source directory where the actual voicebank files are located.
	ContentsDir string

	// Description is a human-readable description or name of the voicebank.
	// This is shown in classic UTAU's installer dialog. Optional.
	Description string
}

// CharacterImage represents a character profile image referenced by character.txt.
type CharacterImage struct {
	// Path is the filesystem path to the image file.
	Path string

	// Image is the decoded image. It may be nil.
	// It gets populated by a call to [CharacterImage.Decode] or
	// by [Open] if asset decoding is enabled.
	Image image.Image

	// Format is the detected image format (e.g. png). It may be empty.
	// It gets populated by a call to [CharacterImage.Decode] or
	// by [Open] if asset decoding is enabled.
	Format string
}

// Decode reads and decodes the profile image file.
//
// By default, Decode supports PNG and BMP images. Additional image codecs must
// be imported separately. For example, to load a JPEG image, it suffices to have
//
//	import _ "image/jpeg"
//
// either in the program's package or somewhere in the import section.
//
// Returns an error if the file cannot be opened or decoded with the available decoders.
func (i *CharacterImage) Decode(fsys fs.FS) error {
	f, err := fsys.Open(i.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return err
	}

	i.Image = img
	i.Format = format

	return nil
}

// CharacterSample represents a character sample audio referenced by character.txt.
type CharacterSample struct {
	// Path is the filesystem path to the sample audio file.
	Path string

	// Sample is the decoded sample audio. It may be nil.
	// It gets populated by a call to [CharacterSample.Decode] or
	// by [Open] if asset decoding is enabled.
	Sample codec.Decoder

	// Format is the detected sample audio format (e.g. wav). It may be empty.
	// It gets populated by a call to [CharacterSample.Decode] or
	// by [Open] if asset decoding is enabled.
	Format string
}

// Decode reads and decodes the sample audio file.
//
// By default, Decode supports AU, WAV, and QOA files and uses [Resona]
// for decoding audio files. Additional audio codecs must
// be imported separately. For example, to load a MP3 file, it suffices to have
//
//	import _ "github.com/MatusOllah/resona/codec/mp3"
//
// either in the program's package or somewhere in the import section.
//
// Returns an error if the file cannot be opened or decoded with the available codecs.
//
// [Resona]: https://github.com/MatusOllah/resona
func (s *CharacterSample) Decode(fsys fs.FS) error {
	f, err := fsys.Open(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec, format, err := codec.Decode(f)
	if err != nil {
		return err
	}

	s.Sample = dec
	s.Format = format

	return nil
}

// CharacterInfo represents metadata about the voicebank character (i.e. character.txt).
type CharacterInfo struct {
	// Name is the character's name.
	Name string

	// Author is the creator or voice actor of the voicebank or character.
	Author string

	// Website is a website URL associated with the character.
	Website string

	// Image is the character's profile image. It is only valid
	// if the image field in the character.txt file is present.
	Image *CharacterImage

	// Sample is the character's sample audio. It is only valid
	// if the sample field in the character.txt file is present.
	Sample *CharacterSample

	// Extra contains unrecognized or unsupported lines from character.txt.
	Extra string
}

// Voicebank represents an UTAU voicebank.
type Voicebank struct {
	// InstallInfo is the voicebank's install information.
	// It is only valid if the voicebank is an installer voicebank.
	InstallInfo *InstallInfo

	// CharacterInfo is the voicebank's character information.
	// It is only valid if the character.txt file is present.
	CharacterInfo *CharacterInfo

	// Readme is the voicebank's readme document.
	// It is only valid if a README file (or similar) is present.
	Readme string

	// Oto is the voicebank's oto.ini configuration.
	Oto Oto
}

type voicebankConfig struct {
	fileEncoding encoding.Encoding
	decodeAssets bool
	strict       bool
}

// using a "universal" Option type here instead of something like OpenOption just in case
// if we want to add more functions beyond just reading voicebanks in the future.
// also voicebank.Option sounds better than voicebank.VoicebankOption imo and
// it still distinguishes from voicebank.OtoOption well enough :3

// Option represents an option for passing into voicebank-related functions and methods.
type Option func(*voicebankConfig)

// WithFileEncoding specifies the character encoding to use when reading or writing voicebanks.
func WithFileEncoding(encoding encoding.Encoding) Option {
	return func(cfg *voicebankConfig) {
		cfg.fileEncoding = encoding
	}
}

// WithDecodeAssets specifies whether to automatically decode assets (e.g. audio, images) when reading voicebanks.
func WithDecodeAssets(decodeAssets bool) Option {
	return func(cfg *voicebankConfig) {
		cfg.decodeAssets = decodeAssets
	}
}

// WithStrict specifies whether to fail on non-fatal errors.
func WithStrict(strict bool) Option {
	return func(cfg *voicebankConfig) {
		cfg.strict = strict
	}
}

// Open opens an UTAU voicebank from the given filesystem.
func Open(fsys fs.FS, opts ...Option) (*Voicebank, error) {
	cfg := &voicebankConfig{
		fileEncoding: encoding.Nop,
		decodeAssets: true,
		strict:       false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if fileExists(fsys, "install.txt") {
		info, err := parseInstallInfo(fsys, cfg.fileEncoding)
		if err != nil {
			return nil, fmt.Errorf("voicebank: failed to parse install.txt: %w", err)
		}

		vbFsys, err := fs.Sub(fsys, info.ContentsDir)
		if err != nil {
			return nil, fmt.Errorf("voicebank: failed to access voicebank directory %q: %w", info.ContentsDir, err)
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
	vb := &Voicebank{}

	// character.txt
	if fileExists(fsys, "character.txt") {
		charInfo, err := parseCharacterInfo(fsys, cfg)
		if err != nil {
			return nil, fmt.Errorf("voicebank: failed to parse character.txt: %w", err)
		}
		vb.CharacterInfo = charInfo
	}

	// README
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // do nothing
		}
		if strings.Contains(strings.ToLower(d.Name()), "readme") {
			readme, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			decoded, err := cfg.fileEncoding.NewDecoder().Bytes(readme)
			if err != nil {
				return err
			}
			vb.Readme = string(decoded)
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("voicebank: failed to access readme file: %w", err)
	}

	// oto.ini
	// parse all oto.ini files found
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // do nothing
		}
		if strings.EqualFold(d.Name(), "oto.ini") {
			oto, err := parseOtoIni(fsys, path, cfg.fileEncoding)
			if err != nil {
				return fmt.Errorf("failed to parse oto.ini file: %w", err)
			}
			vb.Oto = append(vb.Oto, oto...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("voicebank: failed to access oto.ini files: %w", err)
	}

	//TODO: parse prefix.map, + maybe some other stuff

	return vb, nil
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

func parseCharacterInfo(fsys fs.FS, cfg *voicebankConfig) (*CharacterInfo, error) {
	f, err := fsys.Open("character.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scan := bufio.NewScanner(transform.NewReader(f, cfg.fileEncoding.NewDecoder()))

	info := &CharacterInfo{}

	for scan.Scan() {
		// valid values: name, author, web, image, sample
		// the rest is gonna be put in Extra
		line := scan.Text()
		idx := strings.IndexByte(line, '=')
		if idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			switch key {
			case "name":
				info.Name = value
			case "author":
				info.Author = value
			case "web":
				info.Website = value
			case "image":
				info.Image = &CharacterImage{}
				value = strings.ReplaceAll(value, "\\", "/")
				info.Image.Path = value
				if cfg.decodeAssets {
					if err := info.Image.Decode(fsys); err != nil && cfg.strict {
						return nil, fmt.Errorf("failed to decode character image: %w", err)
					}
				}
			case "sample":
				info.Sample = &CharacterSample{}
				value = strings.ReplaceAll(value, "\\", "/")
				info.Sample.Path = value
				if cfg.decodeAssets {
					if err := info.Sample.Decode(fsys); err != nil && cfg.strict {
						return nil, fmt.Errorf("failed to decode character sample audio: %w", err)
					}
				}
			}
		} else {
			info.Extra += line + "\n"
		}

	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	return info, nil
}

func parseOtoIni(fsys fs.FS, path string, enc encoding.Encoding) (Oto, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	oto, err := DecodeOto(f, OtoWithEncoding(enc))
	if err != nil {
		return nil, err
	}

	return oto, nil
}

func fileExists(fsys fs.FS, name string) bool {
	_, err := fs.Stat(fsys, name)
	return err == nil
}
