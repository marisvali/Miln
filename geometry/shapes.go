package geometry

import (
	. "playful-patterns.com/miln/ints"
	. "playful-patterns.com/miln/point"
)

type Line struct {
	Start Pt
	End   Pt
}

type Circle struct {
	Center   Pt
	Diameter Int
}

type Square struct {
	Center Pt
	Size   Int
}
