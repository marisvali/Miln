package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Vision struct {
	previousStart        Pt
	previousObstacles    MatBool
	previousVisibleTiles MatBool
}

var relativeRelevantPtsQ1 Matrix[[]Pt]
var relativeRelevantPtsQ2 Matrix[[]Pt]
var relativeRelevantPtsQ3 Matrix[[]Pt]
var relativeRelevantPtsQ4 Matrix[[]Pt]
var blockSize Int = I(1000)

func init() {

	const matSizeX = 8
	const matSizeY = 8
	p := Pt{}

	relativeRelevantPtsQ1 = NewMatrix[[]Pt](IPt(matSizeX, matSizeY))
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			relativeRelevantPtsQ1.Set(p, computeRelativeRelevantPts(p))
		}
	}

	// Quick and dirty way to get a clone of relativeRelevantPtsQ1.
	relativeRelevantPtsQ2 = NewMatrix[[]Pt](IPt(matSizeX, matSizeY))
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			relativeRelevantPtsQ2.Set(p, computeRelativeRelevantPts(p))
		}
	}

	// Quick and dirty way to get a clone of relativeRelevantPtsQ1.
	relativeRelevantPtsQ3 = NewMatrix[[]Pt](IPt(matSizeX, matSizeY))
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			relativeRelevantPtsQ3.Set(p, computeRelativeRelevantPts(p))
		}
	}

	// Quick and dirty way to get a clone of relativeRelevantPtsQ1.
	relativeRelevantPtsQ4 = NewMatrix[[]Pt](IPt(matSizeX, matSizeY))
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			relativeRelevantPtsQ4.Set(p, computeRelativeRelevantPts(p))
		}
	}

	// If the end is actually to the left of start, just flip all the X for all
	// relative relevant points.
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			pts2 := relativeRelevantPtsQ2.Get(p)
			for i := range pts2 {
				pts2[i].X = pts2[i].X.Negative()
			}
		}
	}

	// If the end is actually below start, just flip all the X for all relative
	// relevant points.
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			pts2 := relativeRelevantPtsQ3.Get(p)
			for i := range pts2 {
				pts2[i].Y = pts2[i].Y.Negative()
			}
		}
	}

	// If the end is both to the left and below start, flip both X and Y.
	for p.Y = ZERO; p.Y.Lt(I(matSizeX)); p.Y.Inc() {
		for p.X = ZERO; p.X.Lt(I(matSizeY)); p.X.Inc() {
			pts2 := relativeRelevantPtsQ4.Get(p)
			for i := range pts2 {
				pts2[i].X = pts2[i].X.Negative()
				pts2[i].Y = pts2[i].Y.Negative()
			}
		}
	}
}

func NewVision(size Pt) (v Vision) {
	return Vision{}
}

func tileToWorldPos(pt Pt) Pt {
	half := blockSize.DivBy(TWO)
	offset := Pt{half, half}
	return pt.Times(blockSize).Plus(offset)
}

func worldPosToTile(pt Pt) Pt {
	return pt.DivBy(blockSize)
}

func computeRelativeRelevantPts(dif Pt) (pts []Pt) {
	// Dif is the difference between v start and an end.
	start := Pt{ZERO, ZERO}
	end := dif

	// Put v square in each position between start and end and check if it
	// blocks the line going from start to end.
	lineStart := tileToWorldPos(Pt{ZERO, ZERO})
	lineEnd := tileToWorldPos(end)
	l := Line{lineStart, lineEnd}

	for y := start.X; y.Leq(end.Y); y.Inc() {
		for x := start.Y; x.Leq(end.X); x.Inc() {
			pt := Pt{x, y}
			if pt == start || pt == end {
				// The start and end points will of course always block v path
				// between start and end. But we consider relevant points to be
				// those points BETWEEN the start and end points which block v
				// path.
				continue
			}

			center := tileToWorldPos(pt)
			size := blockSize.Times(I(98)).DivBy(I(100))
			square := Square{center, size}

			if intersects, _ := LineSquareIntersection(l, square); intersects {
				pts = append(pts, pt)
			}
		}
	}
	return
}

func (v *Vision) isPathClear(start, end Pt, obstacles MatBool) bool {
	if start == end {
		return true
	}

	// For every two points, there are only some points in-between which can
	// block a line going from the center of start to the center of end.
	// I call those points the 'relevant points'. If any relevant point is an
	// obstacle, then we have no clear path between start and end.

	// We need the relevant points between start and end, relevantPts.

	// However, I notice the following. For the following:
	// relevantPts1 = relevant points between start1 = (10, 10) and end (20, 30)
	// relevantPts2 = relevant points between start2 = (20, 15) and end (30, 35)
	// The following is true:
	// relevantPts1[i] - start1 == relevantPts2[i] - start2
	// for every i.
	// I now define relativeRelevantPts[i] = relevantPts[i] - start1.
	// So the relative relevant points are the same between all start and end
	// points where the difference is exactly the same.

	// This tells me I only need to compute the relevant points for all
	// differences which are distinct, not all combinations of (start, end).
	// But, I have to be careful about cases where end.X < start.X for example,
	// or end.Y < start.Y. But I can see a symmetry there as well.
	// If end.X < start.X, I can compute the relevant points as if
	// end.X > start.X, then negate the X of each relevant point. The same with
	// Y.

	// I will describe things as if the (0, 0) point is at lower-left, not
	// upper-left like some other coordinate systems.

	// The relative relevant points only need to be computed once for each
	// difference. I pre-compute all of the relative relevant points in init()
	// and use them here.

	m := &relativeRelevantPtsQ1

	if end.X.Lt(start.X) && end.Y.Geq(start.Y) {
		m = &relativeRelevantPtsQ2
	}

	if end.X.Geq(start.X) && end.Y.Lt(start.Y) {
		m = &relativeRelevantPtsQ3
	}

	if end.X.Lt(start.X) && end.Y.Lt(start.Y) {
		m = &relativeRelevantPtsQ4
	}

	dif := end.Minus(start)
	dif.X = dif.X.Abs()
	dif.Y = dif.Y.Abs()
	relativeRelevantPts := m.Get(dif)

	// Check if any of the relevant points have an obstacle.
	for i := range relativeRelevantPts {
		// Compute relevant point from relative relevant point.
		relevantPt := start.Plus(relativeRelevantPts[i])
		if obstacles.At(relevantPt) {
			return false
		}
	}

	return true
}

func (v *Vision) Compute(start Pt, obstacles MatBool) (visibleTiles MatBool) {
	if start == v.previousStart && obstacles.Equal(v.previousObstacles) {
		visibleTiles = v.previousVisibleTiles
		return
	}

	visibleTiles = NewMatBool(obstacles.Size())

	sz := obstacles.Size()
	for y := ZERO; y.Lt(sz.Y); y.Inc() {
		for x := ZERO; x.Lt(sz.X); x.Inc() {
			end := Pt{x, y}
			if v.isPathClear(start, end, obstacles) {
				visibleTiles.Set(end)
			}
		}
	}

	// Get all visible tiles connected to the player's pos.
	connectedTiles := visibleTiles.ConnectedPositions(start)

	// Eliminate tiles which were marked as visible but are disconnected from
	// the visible region that contains the player's position.
	// This is needed in order to eliminate tiles which are technically
	// reachable if you respect the math, but which just look weird to people.
	// Do the elimination by intersecting sets.
	visibleTiles.IntersectWith(connectedTiles)

	v.previousStart = start
	v.previousObstacles = obstacles.Clone()
	v.previousVisibleTiles = visibleTiles
	return
}
