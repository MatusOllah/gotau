package ust

// Note represents a note in an UST file.
type Note struct {
	Length       int
	Lyric        string
	NoteNum      int
	PreUtterance float64
	VoiceOverlap float64
	Velocity     float64
	StartPoint   float64
	Intensity    int
	Modulation   int
	Flags        string
	PBType       int
	PBStart      float64
	PBS          string
	PBW          string
	PBY          string
	Envelope     string
	VBR          string
}
