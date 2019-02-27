package api

import "math"

// Line is a pair of coordinates forming a line.
type Line struct {
	A, B Coord
}

// Distance is the net distance from A to B.
func (l Line) Distance() float64 {
	return math.Sqrt(
		math.Pow(float64(l.B.X - l.A.X), 2) + math.Pow(float64(l.B.Y - l.A.Y), 2),
	)
}

// Slope is the angle of the line in degrees.
func (l Line) Slope() float64 {
	return radiansToDegrees(math.Atan(float64(
		float64(l.B.Y - l.A.Y) / float64(l.B.X - l.A.X),
	)))
}

func radiansToDegrees(radians float64) (degrees float64) {
	return (radians * 180) / math.Pi
}
