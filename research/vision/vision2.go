package vision

import (
	. "github.com/marisvali/miln/gamelib"
	"slices"
)

type PtSlice []Pt

type VisibleTiles2 struct {
	cachedRelativeRelevantPts Matrix[PtSlice]
}

func NewVisibleTiles2(size Pt) (a VisibleTiles2) {
	a.cachedRelativeRelevantPts = NewMatrix[PtSlice](size)
	return
}

func (a *VisibleTiles2) computeRelativeRelevantPts(dif Pt) (pts []Pt) {
	pts = a.cachedRelativeRelevantPts.Get(dif)
	if pts != nil {
		// We have previously computed this.
		// fmt.Printf("cached: %d %d\n", dif.X.ToInt(), dif.Y.ToInt())
		return
	}

	// Dif is the difference between a start and an end.
	start := Pt{ZERO, ZERO}
	end := dif

	// Put a square in each position between start and end and check if it
	// blocks the line going from start to end.
	lineStart := TileToWorldPos(Pt{ZERO, ZERO})
	lineEnd := TileToWorldPos(end)
	l := Line{lineStart, lineEnd}

	for y := start.X; y.Leq(end.Y); y.Inc() {
		for x := start.Y; x.Leq(end.X); x.Inc() {
			pt := Pt{x, y}
			if pt == start || pt == end {
				// The start and end points will of course always block a path
				// between start and end. But we consider relevant points to be
				// those points BETWEEN the start and end points which block a
				// path.
				continue
			}

			center := TileToWorldPos(pt)
			size := BlockSize.Times(I(98)).DivBy(I(100))
			square := Square{center, size}

			if intersects, _ := LineSquareIntersection(l, square); intersects {
				pts = append(pts, pt)
			}
		}
	}

	// Cache this computation for later.
	// fmt.Printf("computed for: %d %d\n", dif.X.ToInt(), dif.Y.ToInt())
	a.cachedRelativeRelevantPts.Set(dif, pts)
	return
}

func (a *VisibleTiles2) isPathClear(start, end Pt, obstacles MatBool) bool {
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
	// First, compute the relevant points as if end is upper-right compared to
	// start.
	dif := end.Minus(start)
	dif.X = dif.X.Abs()
	dif.Y = dif.Y.Abs()

	// This is the expensive function call that does the main job.
	relativeRelevantPts := a.computeRelativeRelevantPts(dif)

	// If the end is actually to the left of start, just flip all the X for all
	// relative relevant points.
	if end.X.Lt(start.X) {
		// Clone as otherwise we would modify cached values.
		relativeRelevantPts = slices.Clone(relativeRelevantPts)
		for i := range relativeRelevantPts {
			relativeRelevantPts[i].X = relativeRelevantPts[i].X.Negative()
		}
	}

	// If the end is actually below start, just flip all the X for all relative
	// relevant points.
	if end.Y.Lt(start.Y) {
		// Clone as otherwise we would modify cached values.
		relativeRelevantPts = slices.Clone(relativeRelevantPts)
		for i := range relativeRelevantPts {
			relativeRelevantPts[i].Y = relativeRelevantPts[i].Y.Negative()
		}
	}

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

func (a *VisibleTiles2) Compute(start Pt, obstacles MatBool) (visibleTiles MatBool) {
	visibleTiles = NewMatBool(obstacles.Size())

	sz := obstacles.Size()
	for y := ZERO; y.Lt(sz.Y); y.Inc() {
		for x := ZERO; x.Lt(sz.X); x.Inc() {
			end := Pt{x, y}
			if a.isPathClear(start, end, obstacles) {
				visibleTiles.Set(end)
			}
		}
	}
	return
}
