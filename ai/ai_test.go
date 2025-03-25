package ai

import (
	"embed"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"testing"
)

//go:embed data/*
var embeddedFS embed.FS

func GoToFrame(playthrough Playthrough, frameIdx int) World {
	world := NewWorld(playthrough.Seed, playthrough.TargetDifficulty, &embeddedFS)
	for i := 0; i < frameIdx; i++ {
		input := playthrough.History[i]
		world.Step(input)
	}
	return world
}

func ValidMove(world *World, pos Pt) bool {
	freePositions := world.Player.ComputeFreePositions(world)
	return freePositions.At(pos)
}

func ValidAttack(world *World, pos Pt) bool {
	attackablePositions := world.VulnerableEnemyPositions()
	attackablePositions.IntersectWith(world.AttackableTiles)
	return attackablePositions.At(pos)
}

func AmmoAtPos(world *World, pos Pt) bool {
	for _, ammo := range world.Ammos {
		if ammo.Pos == pos {
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
	h.vision = NewVision(world.Obstacles.Size())
	// For the purposes of this calculator, both terrain obstacles and enemies
	// constitute obstacles for vision.
	h.obstacles = world.Obstacles.Clone()
	h.obstacles.Add(h.world.EnemyPositions())
	return h
}

// Get all the positions that are visible from the positions indicated by the
// "startPositions" matrix.
func (h *TargetSeeker) computeVisiblePositions(startPositions MatBool) MatBool {
	positions := startPositions.ToSlice()
	allVisible := NewMatBool(h.obstacles.Size())
	for _, pos := range positions {
		visible := h.vision.Compute(pos, h.obstacles)
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
	lookoutPositions := NewMatBool(h.obstacles.Size())
	lookoutPositions.Set(startPos)

	// Try a maximum of 10 moves for now, even though normally there should be
	// a better stop condition like "all possible positions have been visited".
	for iMove := 0; iMove < 10; iMove++ {
		visiblePositions := h.computeVisiblePositions(lookoutPositions)
		// Check if any targets are visible from the current start positions.
		visibleTargets := visiblePositions.Clone()
		visibleTargets.IntersectWith(targets)
		if len(visibleTargets.ToSlice()) > 0 {
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

// NumMovesToAmmo returns the minimum number of move actions required to end up
// on a tile with ammo on it.
func NumMovesToAmmo(world *World, pos Pt) int {
	ammos := NewMatBool(world.Obstacles.Size())
	for _, ammo := range world.Ammos {
		ammos.Set(ammo.Pos)
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

// NumFramesUntilAttacked returns the number of frames until an enemy attacks pos.
// 0 - there's an enemy at this position right now
// 1 - after 1 frame, the player will be attacked
// ..
// -1 - the position will never be attacked (e.g. there are no more enemies)
func NumFramesUntilAttacked(world *World, pos Pt) int {
	if len(world.Enemies) == 0 {
		return -1
	}

	// Clone the world as we're going to modify it.
	w := world.Clone()

	// Put the player at pos so that enemies will come to it.
	w.Player.SetPos(pos)
	w.Player.OnMap = true
	w.Player.JustHit = false

	// If an enemy doesn't hit within 100k frames, it's not happening.
	input := PlayerInput{} // Don't move, don't attack.
	for frameIdx := 0; frameIdx < 100000; frameIdx++ {
		w.Step(input)
		if w.Player.JustHit {
			return frameIdx
		}
	}
	return -1
}

// FitnessOfMoveAction returns the fitness of a specific move action at a certain moment
// in time, by
func FitnessOfMoveAction(world *World, pos Pt) int {
	if ValidMove(world, pos) {
		// If the action isn't even valid, the fitness of the action is zero.
		return 0
	}
	// The action is valid so other factors come into play.

	// Compute how safe the position is.
	safetyFitness := 0
	framesUntilAttacked := NumFramesUntilAttacked(world, pos)
	// Use time instead of frames because I have an easier time understanding
	// how dangerous a position feels based on how much time it takes for it
	// to be attacked. For example, I know the average for a playthrough was
	// 1 click every 0.6 seconds.
	timeUntilAttacked := float32(framesUntilAttacked) / 60.0
	if timeUntilAttacked <= 0.3 {
		safetyFitness = 0
	}
	if timeUntilAttacked > 0.3 && timeUntilAttacked <= 0.6 {
		safetyFitness = 5
	}
	if timeUntilAttacked > 0.6 && timeUntilAttacked <= 1 {
		safetyFitness = 10
	}
	if timeUntilAttacked > 1 && timeUntilAttacked <= 2 {
		safetyFitness = 15
	}
	if timeUntilAttacked > 2 {
		safetyFitness = 20
	}

	if safetyFitness == 0 {
		// If the position is too dangerous, the fitness is zero.
		return 0
	}

	ammoFitness := 0
	currentAmmo := world.Player.AmmoCount.ToInt()
	numMovesToAmmo := NumMovesToAmmo(world, pos)
	if currentAmmo == 0 {
		if numMovesToAmmo == 0 {
			ammoFitness = 10
		}
		if numMovesToAmmo == 1 {
			ammoFitness = 5
		}
		if numMovesToAmmo == 2 {
			ammoFitness = 2
		}
	}

	if currentAmmo > 1 && currentAmmo <= 3 {
		if numMovesToAmmo == 0 {
			ammoFitness = 6
		}
		if numMovesToAmmo == 1 {
			ammoFitness = 2
		}
		if numMovesToAmmo == 2 {
			ammoFitness = 1
		}
	}

	if currentAmmo > 3 {
		if numMovesToAmmo == 0 {
			ammoFitness = 2
		}
		if numMovesToAmmo == 1 {
			ammoFitness = 1
		}
	}

	enemyFitness := 0
	movesToVisibleEnemy := NumMovesToEnemy(world, pos)
	if currentAmmo > 0 {
		if movesToVisibleEnemy == 0 {
			enemyFitness = 10
		}
		if movesToVisibleEnemy == 1 {
			enemyFitness = 5
		}
		if movesToVisibleEnemy == 2 {
			enemyFitness = 2
		}
	}

	// if there is some safety, then the other factors come into play
	return safetyFitness + ammoFitness + enemyFitness
}

func TestAI(t *testing.T) {
	inputFile := "d:\\gms\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170648.mln010"
	playthrough := DeserializePlaythrough(ReadFile(inputFile))
	// Go to a frame before which the player jumps on an ammo.
	world := GoToFrame(playthrough, 315)

	// numMovesToAmmo := NumMovesToEnemy(&world, world.Player.Pos())
	// fmt.Println(numMovesToAmmo)
	pt := world.Player.Pos()
	pt.X.Add(TWO)
	numFramesUntilAttacked := NumFramesUntilAttacked(&world, pt)
	fmt.Println(numFramesUntilAttacked)
	fmt.Println(float32(numFramesUntilAttacked) / 60.0)

	// worldSize := world.Obstacles.Size()
	// for y := 0; y < worldSize.Y.ToInt(); y++ {
	// 	for x := 0; x < worldSize.X.ToInt(); x++ {
	// 		valid := ValidAttack(&world, IPt(x, y))
	// 		fmt.Printf("%t ", valid)
	// 	}
	// 	fmt.Printf("\n")
	// }

	assert.True(t, true)
}
