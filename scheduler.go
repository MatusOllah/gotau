package gotau

import (
	"cmp"
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

var cmpFn = func(a, b sequence.Note) int { return cmp.Compare(a.Position, b.Position) }

func (s *scheduler) ensureQueueSorted() {
	if slices.IsSortedFunc(s.queue, cmpFn) {
		return
	}
	slices.SortFunc(s.queue, cmpFn)
}

func (s *scheduler) pop() (sequence.Note, bool) {
	if len(s.queue) == 0 {
		return sequence.Note{}, false
	}
	note := s.queue[0]
	s.queue = s.queue[1:]
	return note, true
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
