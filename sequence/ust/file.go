package ust

import (
	"github.com/SladkyCitron/gotau/sequence"
	"github.com/SladkyCitron/gotau/umath"
	"gopkg.in/ini.v1"
)

// File represents a parsed UST file.
type File struct {
	Version  Version  // Version is the UST file format version.
	Settings Settings // Settings represents the settings of the UST file.
	Notes    []Note   // Notes hold the notes.

	iniFile *ini.File // iniFile holds the raw parsed INI file structure (used internally).
}

func (f *File) Sequence() sequence.Sequence {
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
	for _, note := range f.Notes {
		if IsLyricRest(note.Lyric) {
			position += note.Length
			continue
		}

		msPerTick := 60000 / (f.Settings.Tempo * float32(seq.Metadata.Resolution))

		seq.Notes = append(seq.Notes, sequence.Note{
			Position:     position,
			Duration:     note.Length,
			Lyric:        note.Lyric,
			Note:         note.NoteNum,
			Intensity:    note.Intensity,
			PreUtterance: note.PreUtterance,
			VoiceOverlap: note.VoiceOverlap,
			StartPoint:   note.StartPoint,
			Envelope:     envelopeToCurve(note.Envelope, msPerTick*float32(note.Length)),
			PitchBend:    pitchBendToCurve(note.PitchBend, msPerTick),
		})
		position += note.Length
	}
	return seq
}

func envelopeToCurve(env *Envelope, noteDurMs float32) sequence.Curve {
	points := make(sequence.Curve, 0, 5)

	if env == nil {
		return points
	}

	add := func(t, v float32) {
		// t = milliseconds
		// v = percentage
		points = append(points, sequence.CurvePoint{
			XY: umath.XY[float32]{
				X: t,
				Y: v / 100,
			},
			Interp: sequence.CurveInterpolationLinear,
		})
	}

	resolveV := func(value EnvelopeValue) float32 {
		if value.Auto {
			return 0
		} else {
			return value.Value
		}
	}

	add(env.P1.Value, resolveV(env.V1))
	add(env.P2.Value, resolveV(env.V2))
	add(env.P3.Value, resolveV(env.V3))

	p4 := env.P4.Value
	if env.P4.Auto {
		p4 = noteDurMs
	}
	add(p4, resolveV(env.V4))

	if !env.P5.Auto || !env.V5.Auto {
		p5 := env.P5.Value
		if env.P5.Auto {
			p5 = noteDurMs
		}
		v5 := env.V5.Value
		if env.V5.Auto {
			v5 = 0
		}
		add(p5, v5)
	}

	return points
}

func pitchBendToCurve(pb *PitchBend, msPerTick float32) sequence.Curve {
	if pb == nil {
		return sequence.Curve{}
	}
	if (pb.Start.X == 0 && pb.Start.Y == 0) || len(pb.Widths) == 0 {
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

	points := sequence.Curve{}

	x := pb.Start.X
	y := pb.Start.Y
	for i := range pb.Widths {
		points = append(points, sequence.CurvePoint{
			XY:     umath.XY[float32]{X: x * msPerTick, Y: y},
			Interp: convertPBM(pb.Modes[i]),
		})
		x += pb.Widths[i]
		y = pb.Ys[i]
	}

	return points
}

func convertPBM(mode PitchBendMode) sequence.CurveInterpolation {
	return sequence.CurveInterpolation(mode) // this is enough for now
}
