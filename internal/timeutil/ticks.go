package timeutil

import "time"

func SecondsToTicks(seconds float64, tpqn int, bpm float64) int {
	return int(seconds * float64(tpqn) * float64(bpm) / 60)
}

func TicksToSeconds(ticks int, tpqn int, bpm float64) float64 {
	return float64(ticks) / (float64(tpqn) * float64(bpm) / 60)
}

func TicksToDuration(ticks, tpqn int, bpm float64) time.Duration {
	seconds := TicksToSeconds(ticks, tpqn, bpm)
	return time.Duration(seconds * 1e9)
}
