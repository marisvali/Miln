package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

// NumFramesUntilAttacked returns the number of frames until an enemy attacks pos.
// 0 - there's an enemy at this position right now
// 1 - after 1 frame, the player will be attacked
// ..
// -1 - the position will never be attacked (e.g. there are no more enemies)
func NumFramesUntilAttacked(w World, pos Pt) int64 {
	// So far the algorithms calling this function all have reasons to assume
	// that the player will get hit. If it's not the case
	// I should be warned.
	if w.Enemies.N == 0 {
		Check(fmt.Errorf("something went wrong"))
		return -1
	}

	// Put the player at pos so that enemies will come to it.
	w.Player.SetPos(pos)
	w.Player.OnMap = true
	w.Player.JustHit = false

	// If an enemy doesn't hit within 1000 frames, it's not happening.
	input := PlayerInput{} // Don't move, don't attack.
	for frameIdx := int64(0); frameIdx < 1000; frameIdx++ {
		w.Step(input)
		if w.Player.JustHit {
			return frameIdx
		}
	}

	// No algorithm I want to run will want to take the hit of running 100k
	// frames for nothing. So far the algorithms calling this function all
	// have reasons to assume that the player will get hit. If it's not the case
	// I should be warned.
	// Check(fmt.Errorf("something went wrong"))
	// return -1
	return 1000
}
func ValidMove(world *World, pos Pt) bool {
	freePositions := world.Player.ComputeFreePositions(world)
	return freePositions.At(pos)
}

func ValidAttack(world *World, pos Pt) bool {
	if !world.Player.OnMap || world.Player.AmmoCount == ZERO {
		return false
	}
	attackablePositions := world.VulnerableEnemyPositions()
	attackablePositions.IntersectWith(world.VisibleTiles)
	return attackablePositions.At(pos)
}

// NumMovesToAmmo returns the minimum number of move actions required to end up
// on a tile with ammo on it.
func NumMovesToAmmo(world *World, pos Pt) int {
	ammos := MatBool{}
	for i := range world.Ammos.N {
		ammos.Set(world.Ammos.V[i].Pos)
	}
	if ammos.Get(pos) {
		return 0
	}

	seeker := NewTargetSeeker(world)
	return seeker.NumMovesUntilTargetVisible(pos, ammos) + 1
}

// NumMovesToEnemy returns the minimum number of move actions required to end up
// on a tile from which a vulnerable enemy is visible.
func NumMovesToEnemy(world *World, pos Pt) int {
	seeker := NewTargetSeeker(world)
	return seeker.NumMovesUntilTargetVisible(pos, world.VulnerableEnemyPositions())
}
func AmmoAtPos(world *World, pos Pt) bool {
	for i := range world.Ammos.N {
		if world.Ammos.V[i].Pos == pos {
			return true
		}
	}
	return false
}

type TargetSeeker struct {
	vision    Vision
	world     *World
	obstacles MatBool
}

func NewTargetSeeker(world *World) (h TargetSeeker) {
	h.world = world
	h.vision = NewVision()
	// For the purposes of this calculator, both terrain obstacles and enemies
	// constitute obstacles for vision.
	h.obstacles = world.Obstacles
	h.obstacles.Add(h.world.EnemyPositions())
	return h
}

// Get all the positions that are visible from the positions indicated by the
// "startPositions" matrix.
func (h *TargetSeeker) computeVisiblePositions(startPositions MatBool) MatBool {
	positions := startPositions.ToArray()
	allVisible := MatBool{}
	for i := range positions.N {
		visible := h.vision.Compute(positions.V[i], h.obstacles)
		allVisible.Add(visible)
	}
	return allVisible
}

// NumMovesUntilTargetVisible computes the minimum number of moves required to see
// one of the positions in "targets". We consider "startPos" to be the
// first position from which we start looking/moving. Here we are talking about
// moves in the sense of a player move: if the player is at position X, in 1
// move he can change his position to any of the positions visible from X that
// are not occupied by an enemy or a terrain obstacle.
// 0 - a target is visible from startPos already and no moves are required
// 1 - a target will become visible after making 1 move from startPos
// 2 - a target will become visible after making 2 moves from startPos
// ..
// -1 - no target is reachable from startPos, no matter the number of moves
func (h *TargetSeeker) NumMovesUntilTargetVisible(startPos Pt, targets MatBool) int {
	// lookoutPositions - the positions from which we will look out and check
	// if we see a target.
	lookoutPositions := MatBool{}
	lookoutPositions.Set(startPos)

	// Try a maximum of 10 moves for now, even though normally there should be
	// a better stop condition like "all possible positions have been visited".
	for iMove := 0; iMove < 10; iMove++ {
		visiblePositions := h.computeVisiblePositions(lookoutPositions)
		// Check if any targets are visible from the current start positions.
		visibleTargets := visiblePositions
		visibleTargets.IntersectWith(targets)
		if visibleTargets.ToArray().N > 0 {
			// A target is visible.
			return iMove
		}
		// Compute the lookout positions for the next move.
		// They are all the positions visible at this move, minus the obstacles
		// and minus the current lookoutPositions (no sense recomputing for the
		// positions in lookoutPositions).
		newLookoutPositions := visiblePositions
		newLookoutPositions.Subtract(h.obstacles)
		newLookoutPositions.Subtract(lookoutPositions)
		lookoutPositions = newLookoutPositions
	}
	return -1 // Haven't reached any target.
}
