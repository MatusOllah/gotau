package gotau

import (
	"iter"
	"sort"

	"github.com/SladkyCitron/gotau/internal/timeutil"
	"github.com/SladkyCitron/gotau/sequence"
)

type scheduler struct {
	queue   []sequence.Note
	tpqn    int
	bpm     float32
	tickPos int
}

func (s *scheduler) enqueue(notes ...sequence.Note) {
	s.queue = append(s.queue, notes...)
	sort.Slice(s.queue, func(i, j int) bool {
		return s.queue[i].Position < s.queue[j].Position
	})
}

// popSeq returns and dequeues all notes that are ready to be rendered up to seconds.
func (s *scheduler) popSeq(seconds float64) iter.Seq[sequence.Note] {
	return func(yield func(sequence.Note) bool) {
		ticks := s.secondsToTicks(seconds)
		i := 0
		for i < len(s.queue) && ticks > 0 {
			note := s.queue[i]
			ticks -= note.Duration
			if !yield(note) {
				break
			}
			i++
		}
		s.queue = s.queue[i:]
	}
}

func (s *scheduler) secondsToTicks(seconds float64) int {
	return timeutil.SecondsToTicks(seconds, s.tpqn, float64(s.bpm))
}

func (s *scheduler) ticksToSeconds(ticks int) float64 {
	return timeutil.TicksToSeconds(ticks, s.tpqn, float64(s.bpm))
}
