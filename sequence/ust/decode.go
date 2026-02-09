package ust

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/gomidi/midi/v2"
	"golang.org/x/net/html/charset"
	"gopkg.in/ini.v1"
)

var noteRe *regexp.Regexp = regexp.MustCompile(`#\d+`)

// Decode decodes a UST file.
func Decode(r io.Reader) (file *File, err error) {
	// detect encoding, step 1: sniff
	const sniffLen = 256
	buf := make([]byte, sniffLen)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read UST file for encoding detection: %w", err)
	}

	// step 2: find charset in the sniff
	_charset := "shift_jis" // default
	for line := range strings.SplitSeq(string(buf[:n]), "\n") {
		if strings.Contains(line, "Charset=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				_charset = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	concatReader := io.MultiReader(bytes.NewReader(buf[:n]), r)

	// step 3: transcode to UTF-8
	newReader := concatReader
	lc := strings.ToLower(strings.TrimSpace(_charset))
	if lc != "" && lc != "utf-8" && lc != "utf8" {
		dec, err := charset.NewReaderLabel(lc, concatReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create decoder for charset %s: %w", strconv.Quote(_charset), err)
		}
		newReader = dec
	}

	file = &File{}
	loadopts := ini.LoadOptions{
		UnparseableSections: []string{"#VERSION"},
	}
	file.iniFile, err = ini.LoadSources(loadopts, newReader)
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
			return file, nil
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
	for line := range strings.SplitSeq(sec.Body(), "\n") {
		if strings.Contains(line, "UST Version") {
			ver, err := ParseVersion(strings.TrimSpace(line))
			if err != nil {
				return err
			}
			f.Version = ver
		}
	}
	return nil
}

func (f *File) parseSetting(sec *ini.Section) (err error) {
	f.Settings = Settings{}

	tempo, err := strconv.ParseFloat(sec.Key("Tempo").String(), 32) // Tempo
	if err != nil {
		return fmt.Errorf("failed to parse tempo: %w", err)
	}
	f.Settings.Tempo = float32(tempo)

	f.Settings.ProjectName = sec.Key("ProjectName").String() // ProjectName
	f.Settings.Project = sec.Key("Project").String()         // Project (OpenUtau)
	f.Settings.VoiceDir = sec.Key("VoiceDir").String()       // VoiceDir
	f.Settings.OutFile = sec.Key("OutFile").String()         // OutFile
	f.Settings.CacheDir = sec.Key("CacheDir").String()       // CacheDir
	f.Settings.Tool1 = sec.Key("Tool1").String()             // Tool1
	f.Settings.Tool2 = sec.Key("Tool2").String()             // Tool2

	f.Settings.Mode2, err = sec.Key("Mode2").Bool() // Mode2
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
		intensity, err := strconv.ParseFloat(key.String(), 32)
		if err != nil {
			return fmt.Errorf("failed to parse intensity: %w", err)
		}
		note.Intensity = float32(intensity)
	}

	// Velocity
	note.Velocity = nil
	if key, err := sec.GetKey("Velocity"); err == nil && key.String() != "" {
		velocity, err := strconv.ParseFloat(key.String(), 32)
		if err != nil {
			return fmt.Errorf("failed to parse velocity: %w", err)
		}
		velocity32 := float32(velocity)
		note.Velocity = &velocity32
	}

	// Modulation
	note.Modulation = 0
	if key, err := sec.GetKey("Modulation"); err == nil && key.String() != "" {
		modulation, err := strconv.ParseFloat(key.String(), 32)
		if err != nil {
			return fmt.Errorf("failed to parse modulation: %w", err)
		}
		note.Modulation = float32(modulation)
	}

	// PreUtterance
	note.PreUtterance = nil
	if key, err := sec.GetKey("PreUtterance"); err == nil && key.String() != "" {
		preUtterance, err := strconv.ParseFloat(key.String(), 32)
		if err != nil {
			return fmt.Errorf("failed to parse pre-utterance: %w", err)
		}
		preUtterance32 := float32(preUtterance)
		note.PreUtterance = &preUtterance32
	}

	// VoiceOverlap
	note.VoiceOverlap = nil
	if key, err := sec.GetKey("VoiceOverlap"); err == nil && key.String() != "" {
		voiceOverlap, err := strconv.ParseFloat(key.String(), 32)
		if err != nil {
			return fmt.Errorf("failed to parse voice overlap: %w", err)
		}
		voiceOverlap32 := float32(voiceOverlap)
		note.VoiceOverlap = &voiceOverlap32
	}

	// StartPoint
	note.StartPoint = nil
	if key, err := sec.GetKey("StartPoint"); err == nil && key.String() != "" {
		startPoint, err := strconv.ParseFloat(key.String(), 32)
		if err != nil {
			return fmt.Errorf("failed to parse start point: %w", err)
		}
		startPoint32 := float32(startPoint)
		note.StartPoint = &startPoint32
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
