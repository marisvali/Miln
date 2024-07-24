package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"math"
)

type AI struct {
}

func ClosestEnemy(pt Pt, w *World) Enemy {
	minDist := I(math.MaxInt64)
	minI := -1
	for i, e := range w.Enemies {
		dist := e.Pos().Minus(pt).SquaredLen()
		if dist.Lt(minDist) {
			minDist = dist
			minI = i
		}
	}
	return w.Enemies[minI]
}

func (a *AI) Step(w *World) (input PlayerInput) {
	// Move and shoot randomly.
	input.Move = RInt(I(0), I(1)) == I(0)
	input.MovePt = Pt{RInt(I(0), w.Obstacles.Size().X.Minus(ONE)),
		RInt(I(0), w.Obstacles.Size().Y.Minus(ONE))}
	input.Shoot = !input.Move
	input.ShootPt = Pt{RInt(I(0), w.Obstacles.Size().X.Minus(ONE)),
		RInt(I(0), w.Obstacles.Size().Y.Minus(ONE))}

	// Move to a random attackable tile.
	var pts []Pt
	if !w.Player.OnMap {
		// Move to random non-obstacle tile.
		free := w.Obstacles.Clone()
		free.Negate()
		pts = free.ToSlice()
		input.MovePt = pts[RInt(I(0), I(len(pts)-1)).ToInt()]
	} else {
		pts = w.AttackableTiles.ToSlice()
	}
	input.MovePt = pts[RInt(I(0), I(len(pts)-1)).ToInt()]

	// Move to the tile that's furthest away from everyone.
	// For each attackable tile compute how far it is from all other enemies.
	if len(w.Enemies) > 0 {
		maxDist := I(0)
		maxPt := input.MovePt
		for _, pt := range pts {
			e := ClosestEnemy(pt, w)
			dist := pt.Minus(e.Pos()).SquaredLen()
			if dist.Gt(maxDist) {
				maxDist = dist
				maxPt = pt
			}
		}
		input.MovePt = maxPt
	}

	// Shoot at the closest guy that isn't frozen.
	minDist := I(math.MaxInt64)
	minI := -1
	for i, e := range w.Enemies {
		if e.FreezeCooldownIdx().IsZero() {
			dist := e.Pos().Minus(w.Player.Pos).SquaredLen()
			if dist.Lt(minDist) {
				minDist = dist
				minI = i
			}
		}
	}
	if minI > 0 {
		input.ShootPt = w.Enemies[minI].Pos()
	}
	return
}
