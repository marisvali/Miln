package ai

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
		if e.Pos().Minus(pos).Len().Lt(I(DistanceToDangerZone)) {
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

func (a *AI) MoveRandomly(w *World) (input PlayerInput) {
	// Move and shoot randomly.
	input.Move = a.frameIdx.Mod(TWO).Eq(ZERO)
	input.MovePt = Pt{RInt(I(0), w.Obstacles.Size().X.Minus(ONE)),
		RInt(I(0), w.Obstacles.Size().Y.Minus(ONE))}
	input.Shoot = !input.Move
	input.ShootPt = Pt{RInt(I(0), w.Obstacles.Size().X.Minus(ONE)),
		RInt(I(0), w.Obstacles.Size().Y.Minus(ONE))}
	return
}

func (a *AI) MoveToRandomAttackableTile(freePts []Pt) (input PlayerInput) {
	if len(freePts) > 0 {
		input.MovePt = RElem(freePts)
	} else {
		input.Move = false
	}
	return
}

func PointFurthestFromEnemies(w *World, pts []Pt, defaultPt Pt) (maxPt Pt) {
	maxPt = defaultPt
	maxDist := I(0)
	if len(w.Enemies) == 0 {
		return
	}

	for _, pt := range pts {
		e := ClosestEnemy(pt, w)
		dist := pt.Minus(e.Pos()).SquaredLen()
		if dist.Gt(maxDist) {
			maxDist = dist
			maxPt = pt
		}
	}
	return
}

func GetPointsFromWhichPlayerCanAttack(w *World, freePts []Pt) (attackPts []Pt) {
	for _, pt := range freePts {
		// Move to this point and see if anything is attackable after moving
		// there.
		cloneW := w.Clone()
		var moveInput PlayerInput
		moveInput.Move = true
		moveInput.MovePt = pt
		cloneW.Step(moveInput)
		// Compensate for annoying bug that I can't fix in order to keep
		// perfect reproducibility and step again without doing anything.
		// The attackable tiles are not computed correctly in the world,
		// they are computed before the player moves, so they are always
		// behind with 1 frame.
		// This is not noticeable for a human player, but for AI it is
		// noticeable because the AI looks at the world 1 frame after it
		// makes a move.
		cloneW.Step(PlayerInput{})

		for _, e := range cloneW.Enemies {
			if CanAttackEnemy(&cloneW, e) {
				attackPts = append(attackPts, pt)
			}
		}
	}
	return
}

func ClosestAttackableEnemy(w *World) (closestEnemy Enemy) {
	minDist := I(math.MaxInt64)
	closestEnemy = nil
	for _, e := range w.Enemies {
		if CanAttackEnemy(w, e) {
			dist := e.Pos().Minus(w.Player.Pos()).SquaredLen()
			if dist.Lt(minDist) {
				minDist = dist
				closestEnemy = e
			}
		}
	}
	return
}

var MinFramesBetweenActions = 27
var DistanceToDangerZone = 3

func (a *AI) Step(w *World) (input PlayerInput) {
	a.frameIdx.Inc()
	if a.frameIdx.Mod(I(MinFramesBetweenActions)).Neq(ZERO) {
		return
	}

	freePos := w.Player.ComputeFreePositions(w)
	freePts := freePos.ToSlice()

	// Compute the best point to move to, if we were to move.
	input.Move = true

	// At first, move to the tile that's furthest away from everyone.
	input.MovePt = PointFurthestFromEnemies(w, freePts, w.Player.Pos())

	// Try to find the tile that's furthest from everyone, from which an
	// enemy can be shot.
	attackPts := GetPointsFromWhichPlayerCanAttack(w, freePts)
	maxPt := PointFurthestFromEnemies(w, attackPts, w.Player.Pos())
	if !InDangerZone(w, maxPt) {
		input.MovePt = maxPt
	}

	// Get the key if it exists and is reachable.
	if len(w.Keys) > 0 {
		keyPos := w.Keys[0].Pos
		if freePos.At(keyPos) && !InDangerZone(w, keyPos) {
			input.MovePt = keyPos
		}
	}

	// We may be in the situation that only ultra hounds are left and we keep
	// dodging them, but not getting the key.
	// In this case, do a random safe move in an effort to be within key
	// range at some point.
	// Don't do this often, otherwise it might keep ultrahounds almost
	// in the same place near the key and we keep jumping around them.
	if len(w.Keys) > 0 && !freePos.At(w.Keys[0].Pos) && OnlyUltrahoundsLeft(w) {
		if a.frameIdx.Minus(a.lastRandomMoveIdx).Gt(I(240)) {
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
			if !InDangerZone(w, w.Player.Pos()) {
				input.Move = false
			}
		}
	}

	// Compute the best target to shoot if we were to shoot.
	// Shoot at the closest guy that isn't frozen.
	if closestEnemy := ClosestAttackableEnemy(w); closestEnemy != nil {
		input.Shoot = true
		input.ShootPt = closestEnemy.Pos()
	} else {
		input.Shoot = false
	}

	// Don't shoot king when he's over a spawn portal, because he'll drop the
	// key on a place that we can't reach.
	for _, e := range w.Enemies {
		if _, isKing := e.(*King); isKing {
			if e.Pos() == input.ShootPt {
				for _, sp := range w.SpawnPortals {
					if sp.Pos() == e.Pos() {
						input.Shoot = false
						break
					}
				}
			}
			break
		}
	}

	// Decide if we move or shoot.
	// We should move if we are in a danger zone and we have a better option.
	if !w.Player.OnMap ||
		input.Move &&
			InDangerZone(w, w.Player.Pos()) &&
			!InDangerZone(w, input.MovePt) {
		// Suppress shooting.
		input.Shoot = false
	} else {
		// Shoot if we have something to shoot at.
		// Otherwise, leave move as it is.
		if input.Shoot {
			// Suppress moving.
			input.Move = false
		}
	}

	return
}
