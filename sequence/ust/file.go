package ust

import (
	"fmt"

	"github.com/SladkyCitron/gotau/internal/timeutil"
	"github.com/SladkyCitron/gotau/sequence"
	"gopkg.in/ini.v1"
)

// 0,5,35,0,100,100,0,%,0,10,100
// p1,p2,p3,v1,v2,v3,v4,%,p4,p5,v5
var defaultEnvelope = sequence.Curve{
	{X: 0, Y: 0, Interp: sequence.CurveInterpolationLinear},
	{X: 5, Y: 1, Interp: sequence.CurveInterpolationLinear},
	{X: 35, Y: 1, Interp: sequence.CurveInterpolationLinear},
	{X: 0, Y: 0, Interp: sequence.CurveInterpolationLinear},
	{X: 10, Y: 1, Interp: sequence.CurveInterpolationLinear},
}

var _ sequence.Sequencer = (*File)(nil)

// File represents a parsed UST file.
type File struct {
	Version  Version  // Version is the UST file format version.
	Settings Settings // Settings represents the settings of the UST file.
	Notes    []Note   // Notes hold the notes.

	iniFile *ini.File // iniFile holds the raw parsed INI file structure (used internally).
}

func (f *File) Sequence() (sequence.Sequence, error) {
	seq := sequence.Sequence{
		Metadata: sequence.Metadata{
			Name:          f.Settings.ProjectName,
			VoicebankPath: f.Settings.VoiceDir,
			OutputPath:    f.Settings.OutFile,
			Resolution:    480, // MIDI sequencing default
			Tempo:         f.Settings.Tempo,
		},
	}

	var position int
	for i, note := range f.Notes {
		if IsLyricRest(note.Lyric) {
			position += note.Length
			continue
		}

		flags, err := sequence.ParseFlags(note.Flags)
		if err != nil {
			return sequence.Sequence{}, fmt.Errorf("ust Sequence: failed to parse flags for note #%04d: %w", i+1, err)
		}

		noteDurMs := timeutil.TicksToSeconds(note.Length, seq.Metadata.Resolution, seq.Metadata.Tempo) * 1000
		seq.Notes = append(seq.Notes, sequence.Note{
			Position:     position,
			Duration:     note.Length,
			Lyric:        note.Lyric,
			Note:         note.NoteNum,
			Intensity:    note.Intensity / 100,
			Velocity:     note.Velocity,
			Modulation:   note.Modulation,
			Preutterance: note.Preutterance,
			VoiceOverlap: note.VoiceOverlap,
			StartPoint:   note.StartPoint,
			Envelope:     envelopeToCurve(note.Envelope),
			PitchBend:    pitchBendToCurve(note.PitchBend, f.Settings.Mode2, noteDurMs),
			Flags:        flags,
		})
		position += note.Length
	}
	return seq, nil
}

func envelopeToCurve(env *Envelope) sequence.Curve {
	points := make(sequence.Curve, 0, 5)

	if env == nil {
		return defaultEnvelope
	}

	points = append(points, sequence.CurvePoint{X: env.P1, Y: env.V1 / 100, Interp: sequence.CurveInterpolationLinear})
	points = append(points, sequence.CurvePoint{X: env.P2, Y: env.V2 / 100, Interp: sequence.CurveInterpolationLinear})
	points = append(points, sequence.CurvePoint{X: env.P3, Y: env.V3 / 100, Interp: sequence.CurveInterpolationLinear})
	points = append(points, sequence.CurvePoint{X: env.P4, Y: env.V3 / 100, Interp: sequence.CurveInterpolationLinear})
	if env.HasRelease {
		points[3].Y = env.V4 / 100
		points = append(points, sequence.CurvePoint{X: env.P5, Y: env.V5 / 100, Interp: sequence.CurveInterpolationLinear})
	}

	return points
}

func pitchBendToCurve(pb *PitchBend, mode2 bool, noteDurMs float64) sequence.Curve {
	if pb == nil {
		return sequence.Curve{}
	}
	if len(pb.Widths) == 0 {
		return sequence.Curve{}
	}

	// PBY defaults to 0 for every segment
	for len(pb.Ys) < len(pb.Widths) {
		pb.Ys = append(pb.Ys, 0)
	}

	// PBM defaults to sine
	for len(pb.Modes) < len(pb.Widths) {
		pb.Modes = append(pb.Modes, PitchBendModeSine)
	}

	points := make(sequence.Curve, 0, len(pb.Widths)+1)

	// Mode1
	if !mode2 {
		x := pb.Start.X
		y := pb.Start.Y
		for i := range pb.Widths {
			points = append(points, sequence.CurvePoint{
				X:      x,
				Y:      y,
				Interp: convertPBM(pb.Modes[i]),
			})

			x += pb.Widths[i]
			y = pb.Ys[i]
		}

		// final point
		points = append(points, sequence.CurvePoint{
			X: x,
			Y: y,
		})
		return points
	}

	// Mode2
	x := pb.Start.X
	baseY := pb.Start.Y
	// PBW is normalized segments over note length
	// convert to absolute ms spans
	totalW := 0.0
	for _, w := range pb.Widths {
		totalW += w
	}
	if totalW == 0 {
		return sequence.Curve{}
	}
	scale := noteDurMs / totalW
	y := baseY // PBYs are offsets, accumulate
	for i := range pb.Widths {
		dx := pb.Widths[i] * scale
		x += dx
		y += pb.Ys[i]
		points = append(points, sequence.CurvePoint{
			X:      x,
			Y:      y * 10, // convert deci-semitones to cents
			Interp: convertPBM(pb.Modes[i]),
		})
	}

	// final point
	points = append(points, sequence.CurvePoint{
		X:      pb.Start.X + noteDurMs,
		Y:      y,
		Interp: convertPBM(pb.Modes[len(pb.Modes)-1]),
	})
	return points
}

func convertPBM(mode PitchBendMode) sequence.CurveInterpolation {
	return sequence.CurveInterpolation(mode) // this is enough for now
}
