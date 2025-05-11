package ust

import (
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/go-ini/ini"
	"gitlab.com/gomidi/midi/v2"
)

var noteRe *regexp.Regexp = regexp.MustCompile(`#\d+`)

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
		switch sec.Name() {
		case "DEFAULT":
			continue
		case "#VERSION":
			if err := file.parseVersion(sec); err != nil {
				return nil, fmt.Errorf("failed to parse version section: %w", err)
			}
		case "#SETTING":
			if err := file.parseSetting(sec); err != nil {
				return nil, fmt.Errorf("failed to parse settings: %w", err)
			}
		case "#TRACKEND":
			break
		default:
			if noteRe.MatchString(sec.Name()) {
				if err := file.parseNote(sec); err != nil {
					return nil, fmt.Errorf("failed to parse note %s: %w", sec.Name(), err)
				}
			} else {
				slog.Warn("invalid section, skipping", "name", sec.Name())
				continue
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

func (f *File) parseNote(sec *ini.Section) (err error) {
	note := Note{}

	// Length
	note.Length, err = sec.Key("Length").Int()
	if err != nil {
		return fmt.Errorf("failed to parse length: %w", err)
	}

	// Lyric
	note.Lyric = strings.TrimSpace(sec.Key("Lyric").String())

	// NoteNum
	_noteNum, err := sec.Key("NoteNum").Uint()
	if err != nil {
		return fmt.Errorf("failed to parse note number: %w", err)
	}
	note.NoteNum = midi.Note(_noteNum)

	// Intensity
	note.Intensity = 100
	if key, err := sec.GetKey("Intensity"); err == nil && key.String() != "" {
		note.Intensity, err = key.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse intensity: %w", err)
		}
	}

	// Velocity
	note.Velocity = nil
	if key, err := sec.GetKey("Velocity"); err == nil && key.String() != "" {
		_velocity, err := key.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse velocity: %w", err)
		}
		note.Velocity = &_velocity
	}

	// Modulation
	note.Modulation = 0
	if key, err := sec.GetKey("Modulation"); err == nil && key.String() != "" {
		note.Modulation, err = key.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse modulation: %w", err)
		}
	}

	// PreUtterance
	note.PreUtterance = nil
	if key, err := sec.GetKey("PreUtterance"); err == nil && key.String() != "" {
		_preUtterance, err := key.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse pre-utterance: %w", err)
		}
		note.PreUtterance = &_preUtterance
	}

	// VoiceOverlap
	note.VoiceOverlap = nil
	if key, err := sec.GetKey("VoiceOverlap"); err == nil && key.String() != "" {
		_voiceOverlap, err := key.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse voice overlap: %w", err)
		}
		note.VoiceOverlap = &_voiceOverlap
	}

	// StartPoint
	note.StartPoint = nil
	if key, err := sec.GetKey("StartPoint"); err == nil && key.String() != "" {
		_startPoint, err := key.Float64()
		if err != nil {
			return fmt.Errorf("failed to parse start point: %w", err)
		}
		note.StartPoint = &_startPoint
	}

	// Envelope
	note.Envelope = nil
	if key, err := sec.GetKey("Envelope"); err == nil && key.String() != "" {
		_envelope, err := ParseEnvelope(key.String())
		if err != nil {
			return fmt.Errorf("failed to parse envelope: %w", err)
		}
		note.Envelope = _envelope
	}

	// Pitch bend (PB*)
	note.PitchBend = nil
	if sec.HasKey("PBS") || sec.HasKey("PBStart") {
		_pb, err := ParsePitchBend(
			sec.Key("PBType").String(),
			sec.Key("PBStart").String(),
			sec.Key("PBS").String(),
			sec.Key("PBW").String(),
			sec.Key("PBY").String(),
			sec.Key("PBM").String(),
		)
		if err != nil {
			return fmt.Errorf("failed to parse pitch bend data: %w", err)
		}
		note.PitchBend = _pb
	}

	f.Notes = append(f.Notes, note)

	return nil
}
