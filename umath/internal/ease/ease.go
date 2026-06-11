package ease

import "math"

// t = current time
// b = start value
// c = change from start value
// d = duration

func Linear(t, b, c, d float64) float64 {
	return c*t/d + b
}

func SineIn(t, b, c, d float64) float64 {
	return -c*math.Cos(t/d*(math.Pi/2)) + c + b
}

func SineOut(t, b, c, d float64) float64 {
	return c*math.Sin(t/d*(math.Pi/2)) + b
}

func SineInOut(t, b, c, d float64) float64 {
	return -c/2*(math.Cos(math.Pi*t/d)-1) + b
}

func SineOutIn(t, b, c, d float64) float64 {
	if t < d/2 {
		return SineOut(t*2, b, c/2, d)
	}
	return SineIn((t*2)-d, b+c/2, c/2, d)
}
