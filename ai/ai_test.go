package ai

import (
	"cmp"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"slices"
	"testing"
)

func GoToFrame(playthrough Playthrough, frameIdx int) World {
	world := NewWorld(playthrough.Seed, playthrough.Level)
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
	if world.Player.AmmoCount == ZERO {
		return false
	}
	attackablePositions := world.VulnerableEnemyPositions()
	attackablePositions.IntersectWith(world.VisibleTiles)
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

// FitnessOfMoveAction returns the fitness of a specific move action at a
// certain moment. The action is "the player moves to pos".
func FitnessOfMoveAction(world *World, pos Pt) int {
	if !ValidMove(world, pos) {
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
	if timeUntilAttacked <= 0.05 {
		safetyFitness = 0
	}
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

// FitnessOfAttackAction returns the fitness of a specific attack action at a
// certain moment. The action is "the player attacks pos".
func FitnessOfAttackAction(world *World, pos Pt) int {
	if !ValidAttack(world, pos) {
		// If the action isn't even valid, the fitness of the action is zero.
		return 0
	}
	// The action is valid so other factors come into play.

	// there's some extra usefulness if the enemy will finally die with this shot
	// also there's some usefulness in paralyzing the enemy for a while
	// weeurjiuefiuergeugriuegr etc etc cleanup comments etc
	input := PlayerInput{}
	input.Shoot = true
	input.ShootPt = pos
	w := world.Clone()
	w.Step(input)

	if len(w.Enemies) == 0 {
		// Just won the game.
		return 1000
	}

	// Compute how safe the current position is.
	safetyFitness := 0
	framesUntilAttacked := NumFramesUntilAttacked(&w, w.Player.Pos())
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
		safetyFitness = 20 + int(timeUntilAttacked*3)
	}

	if safetyFitness < 10 {
		// If the current position is too dangerous, the fitness is zero.
		return 0
	}

	// Prefer attacking to moving, if it's safe and possible.
	biasForAttacking := 20

	return safetyFitness + biasForAttacking
}

type Action struct {
	Move    bool
	Pos     Pt
	Fitness int
	Rank    int
}

func (a Action) String() string {
	if a.Move {
		return fmt.Sprintf("rank: %3d fitness: %3d move:  %d %d", a.Rank, a.Fitness, a.Pos.X.ToInt(), a.Pos.Y.ToInt())
	} else {
		return fmt.Sprintf("rank: %3d fitness: %3d shoot: %d %d", a.Rank, a.Fitness, a.Pos.X.ToInt(), a.Pos.Y.ToInt())
	}
}

func RankedActionsPerFrame(playthrough Playthrough, frameIdx int) (actions []Action) {
	world := GoToFrame(playthrough, frameIdx)

	// Compute the fitness of every move action.
	worldSize := world.Obstacles.Size()
	for y := 0; y < worldSize.Y.ToInt(); y++ {
		for x := 0; x < worldSize.X.ToInt(); x++ {
			fitness := FitnessOfMoveAction(&world, IPt(x, y))
			action := Action{}
			action.Move = true
			action.Pos = IPt(x, y)
			action.Fitness = fitness
			actions = append(actions, action)
		}
	}

	// Compute the fitness of every attack action.
	for y := 0; y < worldSize.Y.ToInt(); y++ {
		for x := 0; x < worldSize.X.ToInt(); x++ {
			fitness := FitnessOfAttackAction(&world, IPt(x, y))
			action := Action{}
			action.Move = false
			action.Pos = IPt(x, y)
			action.Fitness = fitness
			actions = append(actions, action)
		}
	}

	// Sort.
	cmpActions := func(a, b Action) int {
		return cmp.Compare(b.Fitness, a.Fitness)
	}
	slices.SortFunc(actions, cmpActions)

	// Rank.
	actions[0].Rank = 1
	for i := 1; i < len(actions); i++ {
		if actions[i].Fitness == actions[i-1].Fitness {
			actions[i].Rank = actions[i-1].Rank
		} else {
			actions[i].Rank = actions[i-1].Rank + 1
		}
	}

	return actions
}

// func GetPlayerActions(playthrough Playthrough, frameIdx int) (actions []Action) {
//
// }

func InputToAction(input PlayerInput) (action Action) {
	if input.Move {
		action.Move = true
		action.Pos = input.MovePt
	} else if input.Shoot {
		action.Move = false
		action.Pos = input.ShootPt
	} else {
		panic(fmt.Errorf("bad"))
	}
	return action
}

func FindActionRank(action Action, actions []Action) int {
	for _, a := range actions {
		if action.Pos == a.Pos && action.Move == a.Move {
			return a.Rank
		}
	}
	panic(fmt.Errorf("bad"))
}

func GetFramesWithActions(playthrough Playthrough) (framesWithActions []int) {
	for i := 0; i < len(playthrough.History); i++ {
		input := playthrough.History[i]
		if input.Move || input.Shoot {
			framesWithActions = append(framesWithActions, i)
		}
	}
	return
}

func GetDecisionFrames(framesWithActions []int) (decisionFrames []int) {
	// Assume player decides about 5 frames after the last action.
	// decisionFrames = append(decisionFrames, framesWithActions[0]-15)
	// for i := 0; i < len(framesWithActions)-1; i++ {
	// 	decisionFrames = append(decisionFrames, framesWithActions[i]+5)
	// }

	// Since the player can only do valid actions and he is protected by auto
	// aims, it's possible that a previously invalid action became valid by the
	// time the player acted.
	// For example the player might decide to shoot a guy, but by the time he
	// got around to clicking, the guy moved (real case). Now, the previously
	// best course of action became impossible, technically, because the guy
	// moved and attacking the previous position is useless. But, the auto-aim
	// will save you from this and right-clicking on the previous position will
	// result in an attack on the enemy's new position.
	// So, just.. screw it for now. Have the algorithm compute the best decision
	// on exactly the frame when the player acts.
	decisionFrames = slices.Clone(framesWithActions)
	return
}

func GetRanksOfPlayerActions(playthrough Playthrough, framesWithActions []int, decisionFrames []int) (ranksOfPlayerActions []int) {
	for actionIdx := 0; actionIdx < len(framesWithActions); actionIdx++ {
		rankedActions := RankedActionsPerFrame(playthrough, decisionFrames[actionIdx])
		playerAction := InputToAction(playthrough.History[framesWithActions[actionIdx]])
		rank := FindActionRank(playerAction, rankedActions)
		ranksOfPlayerActions = append(ranksOfPlayerActions, rank)
	}
	return
}

func ModelFitness(ranksOfPlayerActions []int) (modelFitness int) {
	for i := 0; i < len(ranksOfPlayerActions); i++ {
		diff := ranksOfPlayerActions[i] - 1
		modelFitness += diff * diff
	}
	return modelFitness
}

func DebugRank(framesWithActions []int, decisionFrames []int,
	ranksOfPlayerActions []int, playthrough Playthrough, actionIdx int) {
	fmt.Printf("frame with player action: %d\n", framesWithActions[actionIdx])
	fmt.Printf("decision frame: %d\n", decisionFrames[actionIdx])
	println(ranksOfPlayerActions[actionIdx])
	world := GoToFrame(playthrough, decisionFrames[actionIdx])
	fmt.Printf("%+v\n", InputToAction(playthrough.History[framesWithActions[actionIdx]]))

	println("debugging now")
	println(FitnessOfAttackAction(&world, IPt(7, 4)))

	// Compute the fitness of every move action.
	worldSize := world.Obstacles.Size()
	for y := 0; y < worldSize.Y.ToInt(); y++ {
		for x := 0; x < worldSize.X.ToInt(); x++ {
			fitness := FitnessOfMoveAction(&world, IPt(x, y))
			fmt.Printf("%3d ", fitness)
		}
		fmt.Printf("\n")
	}

	// Compute the fitness of every attack action.
	fmt.Printf("\n")
	for y := 0; y < worldSize.Y.ToInt(); y++ {
		for x := 0; x < worldSize.X.ToInt(); x++ {
			fitness := FitnessOfAttackAction(&world, IPt(x, y))
			fmt.Printf("%3d ", fitness)
		}
		fmt.Printf("\n")
	}
}

func ModelFitnessForPlaythrough(inputFile string) int {
	playthrough := DeserializePlaythrough(ReadFile(inputFile))
	framesWithActions := GetFramesWithActions(playthrough)
	decisionFrames := GetDecisionFrames(framesWithActions)
	ranksOfPlayerActions := GetRanksOfPlayerActions(playthrough, framesWithActions, decisionFrames)
	modelFitness := ModelFitness(ranksOfPlayerActions)
	return modelFitness
}

func TestAI(t *testing.T) {
	inputFiles := []string{
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171419.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171357.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171333.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171313.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171250.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171229.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171159.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171136.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171111.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171054.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171030.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-171010.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170947.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170924.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170900.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170838.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170817.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170755.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170735.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170716.mln010",
		"d:\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170648.mln010"}

	for _, inputFile := range inputFiles {
		modelFitness := ModelFitnessForPlaythrough(inputFile)
		fmt.Printf("modelFitness: %d\n", modelFitness)
	}

	// fmt.Printf("%v\n", ranksOfPlayerActions)
	// for actionIdx := range ranksOfPlayerActions {
	// 	fmt.Printf("%2d ", ranksOfPlayerActions[actionIdx])
	// }
	// println()
	// for actionIdx := range ranksOfPlayerActions {
	// 	fmt.Printf("%2d ", actionIdx)
	// }
	// println()
	//
	// DebugRank(framesWithActions, decisionFrames, ranksOfPlayerActions, playthrough,
	// 	16)
	assert.True(t, true)
}
