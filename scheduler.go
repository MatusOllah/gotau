package gotau

import (
	"sort"

	"github.com/SladkyCitron/gotau/sequence"
)

type scheduler struct {
	queue   []sequence.Note
	tpqn    int
	bpm     float32
	tickPos int
}

func (s *scheduler) enqueue(notes ...sequence.Note) {
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Position < notes[j].Position
	})
	s.queue = append(s.queue, notes...)
}

// pop returns and dequeues all notes that are ready to be rendered up to seconds.
func (s *scheduler) pop(seconds float64) []sequence.Note {
	var ready []sequence.Note
	ticks := s.secondsToTicks(seconds)
	i := 0
	for i < len(s.queue) && ticks > 0 {
		note := s.queue[i]
		ticks -= note.Duration
		ready = append(ready, note)
		i++
	}
	s.queue = s.queue[i:]
	return ready
}

func (s *scheduler) secondsToTicks(seconds float64) int {
	return int(seconds * float64(s.tpqn) * float64(s.bpm) / 60)
}

func (s *scheduler) ticksToSeconds(ticks int) float64 {
	return float64(ticks) / (float64(s.tpqn) * float64(s.bpm) / 60)
}
