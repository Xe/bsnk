package api

import "math"

// Line is a pair of coordinates forming a line.
type Line struct {
	A, B Coord
}

// Distance is the net distance from A to B.
func (l Line) Distance() float64 {
	return math.Sqrt(
		math.Pow(float64(l.B.X-l.A.X), 2) + math.Pow(float64(l.B.Y-l.A.Y), 2),
	)
}

// The Manhattan distance of a line.
func (l Line) Manhattan() float64 {
	absX := l.B.X - l.A.X
	if absX < 0 {
		absX = -absX
	}

	absY := l.B.Y - l.A.Y
	if absY < 0 {
		absY = -absY
	}

	return float64(absX + absY)
}

// Slope is the angle of the line in degrees.
func (l Line) Slope() float64 {
	return radiansToDegrees(math.Atan(float64(
		float64(l.B.Y-l.A.Y) / float64(l.B.X-l.A.X),
	)))
}

func radiansToDegrees(radians float64) (degrees float64) {
	return (radians * 180) / math.Pi
}
