// Package umath provides UTAU math primitives.
package umath

// XY is a 2D x and y value.
type XY[T any] struct {
	X, Y T
}

// XYs is a slice of [XY]s.
type XYs[T any] []XY[T]
