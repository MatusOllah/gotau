package concat

// Append is the simplest [Concatenator] possible.
// It just appends the note samples to the tail without doing any processing. May cause clicks between notes.
type Append struct{}

func (c *Append) Concatenate(tail []float32, note []float32, _ ConcatenateConfig) ([]float32, error) {
	return append(tail, note...), nil
}
