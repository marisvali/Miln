package gamelib

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

type Rectangle struct {
	Corner1 Pt
	Corner2 Pt
}

func (r *Rectangle) Width() Int {
	return r.Corner1.X.Minus(r.Corner2.X).Abs()
}

func (r *Rectangle) Height() Int {
	return r.Corner1.Y.Minus(r.Corner2.Y).Abs()
}

func (r *Rectangle) Min() Pt {
	return Pt{Min(r.Corner1.X, r.Corner2.X), Min(r.Corner1.Y, r.Corner2.Y)}
}

func (r *Rectangle) Max() Pt {
	return Pt{Max(r.Corner1.X, r.Corner2.X), Max(r.Corner1.Y, r.Corner2.Y)}
}

func (r *Rectangle) ContainsPt(pt Pt) bool {
	minX, maxX := MinMax(r.Corner1.X, r.Corner2.X)
	minY, maxY := MinMax(r.Corner1.Y, r.Corner2.Y)
	return pt.X.Geq(minX) && pt.X.Leq(maxX) && pt.Y.Geq(minY) && pt.Y.Leq(maxY)
}

func LineVerticalLineIntersection(l, vert Line) (bool, Pt) {
	// Check if the Lines even intersect.

	// Check if l's min X is at the right of vertX.
	minX, maxX := MinMax(l.Start.X, l.End.X)
	vertX := vert.Start.X // we assume vert.Start.X == vert.End.X

	if minX.Gt(vertX) {
		return false, Pt{}
	}

	// Or if l's max X is at the left of vertX.
	if maxX.Lt(vertX) {
		return false, Pt{}
	}

	//// Check if l's minY is under the vertMaxY.
	//minY, maxY := MinMax(l.Start.Y, l.End.Y)
	//vertMinY, vertMaxY := MinMax(vert.Start.Y, vert.End.Y)
	//
	//if minY.Gt(vertMaxY) {
	//	return false, Pt{}
	//}
	//
	//// Or if l's max Y is above vertMinY.
	//if maxY.Lt(vertMinY) {
	//	return false, Pt{}
	//}

	vertMinY, vertMaxY := MinMax(vert.Start.Y, vert.End.Y)

	// We know the intersection point will have the X coordinate equal to vertX.
	// We just need to compute the Y coordinate.
	// We have to move along the Y axis the same proportion that we moved along
	// the X axis in order to get to the intersection

	//factor := (vertX - l.Start.X) / (l.End.X - l.Start.X) // will always be positive
	//y := l.Start.Y + factor * (l.End.Y - l.Start.Y) // l.End.Y - l.Start.Y will
	// have the proper sign so that Y gets updated in the right direction
	//y := l.Start.Y + (vertX - l.Start.X) / (l.End.X - l.Start.X) * (l.End.Y - l.Start.Y)
	//y := l.Start.Y + (vertX - l.Start.X) * (l.End.Y - l.Start.Y) / (l.End.X - l.Start.X)
	var y Int
	if l.End.X.Eq(l.Start.X) {
		y = l.Start.Y
	} else {
		y = l.Start.Y.Plus((vertX.Minus(l.Start.X)).Times(l.End.Y.Minus(l.Start.Y)).DivBy(l.End.X.Minus(l.Start.X)))
	}

	if y.Lt(vertMinY) || y.Gt(vertMaxY) {
		return false, Pt{}
	} else {
		return true, Pt{vertX, y}
	}
}

func LineHorizontalLineIntersection(l, horiz Line) (bool, Pt) {
	// Check if the Lines even intersect.

	// Check if l's minY is under the vertY.
	minY, maxY := MinMax(l.Start.Y, l.End.Y)
	vertY := horiz.Start.Y // we assume vert.Start.Y == vert.End.Y

	if minY.Gt(vertY) {
		return false, Pt{}
	}

	// Or if l's max Y is above vertY.
	if maxY.Lt(vertY) {
		return false, Pt{}
	}

	//// Check if l's min X is at the right of vertMaxX.
	//minX, maxX := MinMax(l.Start.X, l.End.X)
	//vertMinX, vertMaxX := MinMax(horiz.Start.X, horiz.End.X)
	//
	//if minX.Gt(vertMaxX) {
	//	return false, Pt{}
	//}
	//
	//// Or if l's max X is at the left of vertMinX.
	//if maxX.Lt(vertMinX) {
	//	return false, Pt{}
	//}

	vertMinX, vertMaxX := MinMax(horiz.Start.X, horiz.End.X)

	// We know the intersection point will have the Y coordinate equal to vertY.
	// We just need to compute the X coordinate.
	// We have to move along the X axis the same proportion that we moved along
	// the Y axis in order to get to the intersection

	//factor := (vertY - l.Start.Y) / (l.End.Y - l.Start.Y) // will always be positive
	//x := l.Start.X + factor * (l.End.X - l.Start.X) // l.End.X - l.Start.X will
	// have the proper sign so that Y gets updated in the right direction
	//x := l.Start.X + (vertY - l.Start.Y) / (l.End.Y - l.Start.Y) * (l.End.X - l.Start.X)
	//x := l.Start.X + (vertY - l.Start.Y) * (l.End.X - l.Start.X) / (l.End.Y - l.Start.Y)
	var x Int
	if l.End.Y.Eq(l.Start.Y) {
		x = l.Start.X
	} else {
		x = l.Start.X.Plus((vertY.Minus(l.Start.Y)).Times(l.End.X.Minus(l.Start.X)).DivBy(l.End.Y.Minus(l.Start.Y)))
	}

	if x.Lt(vertMinX) || x.Gt(vertMaxX) {
		return false, Pt{}
	} else {
		return true, Pt{x, vertY}
	}
}

func LineSquareIntersection(l Line, s Square) (bool, Pt) {
	half := s.Size.DivBy(TWO)
	p1 := Pt{s.Center.X.Minus(half), s.Center.Y.Minus(half)}
	p2 := Pt{s.Center.X.Plus(half), s.Center.Y.Minus(half)}
	p3 := Pt{s.Center.X.Plus(half), s.Center.Y.Plus(half)}
	p4 := Pt{s.Center.X.Minus(half), s.Center.Y.Plus(half)}

	l1 := Line{p1, p2}
	l2 := Line{p2, p3}
	l3 := Line{p3, p4}
	l4 := Line{p4, p1}

	ipts := []Pt{}
	if intersects, ipt := LineHorizontalLineIntersection(l, l1); intersects {
		ipts = append(ipts, ipt)
	}
	if intersects, ipt := LineVerticalLineIntersection(l, l2); intersects {
		ipts = append(ipts, ipt)
	}
	if intersects, ipt := LineHorizontalLineIntersection(l, l3); intersects {
		ipts = append(ipts, ipt)
	}
	if intersects, ipt := LineVerticalLineIntersection(l, l4); intersects {
		ipts = append(ipts, ipt)
	}

	return GetClosestPoint(ipts, l.Start)
}

func LineSquaresIntersection(l Line, squares []Square) (bool, Pt) {
	ipts := []Pt{}
	for _, s := range squares {
		if intersects, ipt := LineSquareIntersection(l, s); intersects {
			ipts = append(ipts, ipt)
		}
	}

	return GetClosestPoint(ipts, l.Start)
}

func GetClosestPoint(pts []Pt, refPt Pt) (bool, Pt) {
	if len(pts) == 0 {
		return false, Pt{}
	}

	// Find the point closest to the reference
	minDist := refPt.SquaredDistTo(pts[0])
	minIdx := 0
	for idx := 1; idx < len(pts); idx++ {
		dist := refPt.SquaredDistTo(pts[idx])
		if dist.Lt(minDist) {
			minDist = dist
			minIdx = idx
		}
	}
	return true, pts[minIdx]
}
