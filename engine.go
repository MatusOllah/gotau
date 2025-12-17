package gotau

type Engine struct {
	Scheduler  *Scheduler
	OnProgress func(Progress)
}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) ReadSamples(p []float32) (int, error) {
	//TODO: this
	return 0, nil
}
