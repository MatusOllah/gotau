package gotau

import (
	"cmp"
	"iter"
	"slices"

	"github.com/SladkyCitron/gotau/internal/timeutil"
	"github.com/SladkyCitron/gotau/sequence"
)

type scheduler struct {
	queue   []sequence.Note
	tpqn    int
	bpm     float64
	tickPos int
}

func (s *scheduler) enqueue(notes ...sequence.Note) {
	s.queue = append(s.queue, notes...)
	s.ensureQueueSorted()
}

var sortFn = func(a, b sequence.Note) int { return cmp.Compare(a.Position, b.Position) }

func (s *scheduler) ensureQueueSorted() {
	if slices.IsSortedFunc(s.queue, sortFn) {
		return
	}
	slices.SortFunc(s.queue, sortFn)
}

// popSeq returns and dequeues all notes that are ready to be rendered up to seconds.
func (s *scheduler) popSeq(seconds float64) iter.Seq[sequence.Note] {
	return func(yield func(sequence.Note) bool) {
		ticks := s.secondsToTicks(seconds)
		for len(s.queue) > 0 && ticks > 0 {
			note := s.queue[0]
			ticks -= note.Duration
			if !yield(note) {
				return
			}
			s.queue = s.queue[1:]
		}
	}
}

func (s *scheduler) peek() (sequence.Note, bool) {
	if len(s.queue) == 0 {
		return sequence.Note{}, false
	}
	return s.queue[0], true
}

func (s *scheduler) secondsToTicks(seconds float64) int {
	return timeutil.SecondsToTicks(seconds, s.tpqn, s.bpm)
}

func (s *scheduler) ticksToSeconds(ticks int) float64 {
	return timeutil.TicksToSeconds(ticks, s.tpqn, s.bpm)
}
