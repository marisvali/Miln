package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"math"
)

type AI struct {
	frameIdx          Int
	lastRandomMoveIdx Int
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

func CanAttackEnemy(w *World, e Enemy) bool {
	_, isUltraHound := e.(*UltraHound)
	return (w.Player.HitPermissions.CanHitUltraHound || !isUltraHound) &&
		e.FreezeCooldownIdx().IsZero() &&
		w.AttackableTiles.At(e.Pos())
}

func InDangerZone(w *World, pos Pt) bool {
	for _, e := range w.Enemies {
		if e.Pos().Minus(pos).Len().Lt(TWO) {
			return true
		}
	}
	return false
}

func OnlyUltrahoundsLeft(w *World) bool {
	for _, e := range w.Enemies {
		_, isUltraHound := e.(*UltraHound)
		if !isUltraHound {
			return false
		}
	}
	return true
}

func (a *AI) Step(w *World) (input PlayerInput) {
	a.frameIdx.Inc()

	// Move and shoot randomly.
	input.Move = a.frameIdx.Mod(TWO).Eq(ZERO)
	input.MovePt = Pt{RInt(I(0), w.Obstacles.Size().X.Minus(ONE)),
		RInt(I(0), w.Obstacles.Size().Y.Minus(ONE))}
	input.Shoot = !input.Move
	input.ShootPt = Pt{RInt(I(0), w.Obstacles.Size().X.Minus(ONE)),
		RInt(I(0), w.Obstacles.Size().Y.Minus(ONE))}

	// Move to a random attackable tile.
	freePos := w.Player.ComputeFreePositions(w)
	freePts := freePos.ToSlice()
	if len(freePts) > 0 {
		input.MovePt = RElem(freePts)
	} else {
		input.Move = false
	}

	// Move to the tile that's furthest away from everyone.
	// For each attackable tile compute how far it is from all other enemies.
	if len(w.Enemies) > 0 {
		maxDist := I(0)
		maxPt := input.MovePt
		for _, pt := range freePts {
			e := ClosestEnemy(pt, w)
			dist := pt.Minus(e.Pos()).SquaredLen()
			if dist.Gt(maxDist) {
				maxDist = dist
				maxPt = pt
			}
		}
		input.MovePt = maxPt
	}

	// Move to the tile that's furthest away from everyone, from which an enemy
	// can be shot.
	// Compute from which free points an enemy can be shot.
	if a.frameIdx.Mod(I(50)).Eq(ZERO) {
		attackPts := []Pt{}
		for _, pt := range freePts {
			// Move to this point and see if anything is attackable after moving
			// there.
			cloneW := w.Clone()
			var moveInput PlayerInput
			moveInput.Move = true
			moveInput.MovePt = pt
			cloneW.Step(moveInput)

			for _, e := range cloneW.Enemies {
				if CanAttackEnemy(&cloneW, e) {
					attackPts = append(attackPts, pt)
				}
			}
		}

		// For each tile that permits an attack compute how far it is from all
		// other enemies.
		if len(w.Enemies) > 0 {
			maxDist := I(0)
			maxPt := input.MovePt
			for _, pt := range attackPts {
				e := ClosestEnemy(pt, w)
				dist := pt.Minus(e.Pos()).Len()
				if dist.Gt(maxDist) {
					maxDist = dist
					maxPt = pt
				}
			}
			if maxDist.Gt(I(1)) {
				input.MovePt = maxPt
			}
		}
	}

	// Shoot at the closest guy that isn't frozen.
	minDist := I(math.MaxInt64)
	minI := -1
	for i, e := range w.Enemies {
		if CanAttackEnemy(w, e) {
			dist := e.Pos().Minus(w.Player.Pos).SquaredLen()
			if dist.Lt(minDist) {
				minDist = dist
				minI = i
			}
		}
	}
	if minI >= 0 {
		input.ShootPt = w.Enemies[minI].Pos()
	} else {
		input.Shoot = false
	}

	// Move to get key if it exists and is reachable.
	if len(w.Keys) > 0 && freePos.At(w.Keys[0].Pos) {
		tooClose := false
		for _, e := range w.Enemies {
			if e.Pos().Minus(w.Keys[0].Pos).Len().Lt(TWO) {
				tooClose = true
				break
			}
		}
		if !tooClose {
			input.MovePt = w.Keys[0].Pos
		}
	}

	// We may be in the situation that only ultra hounds are left and we keep
	// dodging them, but not getting the key.
	// In this case, do a random safe move in an effort to be within key
	// range at some point.
	// But don't do this every frame, otherwise it might keep ultrahounds almost
	// in the same place near the key and we keep jumping around them.
	if len(w.Keys) > 0 && !freePos.At(w.Keys[0].Pos) && OnlyUltrahoundsLeft(w) {
		if a.frameIdx.Minus(a.lastRandomMoveIdx).Gt(I(120)) {
			// Only try a maximum number of times, because such a position might
			// not exist and I don't want to get stuck in an infinite loop.
			for i := 0; i < 10; i++ {
				pt := RElem(freePts)
				if !InDangerZone(w, pt) {
					input.Move = true
					input.MovePt = pt
					a.lastRandomMoveIdx = a.frameIdx
					break
				}
			}
		} else {
			// Just sit in the random position for a while.
			if !InDangerZone(w, w.Player.Pos) {
				input.Move = false
			}
		}
	}
	return
}
