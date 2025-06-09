package ai

import (
	"bufio"
	"cmp"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"os"
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
	if !world.Player.OnMap || world.Player.AmmoCount == ZERO {
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
	// So far the algorithms calling this function all have reasons to assume
	// that the player will get hit. If it's not the case
	// I should be warned.
	if len(world.Enemies) == 0 {
		Check(fmt.Errorf("something went wrong"))
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

	// No algorithm I want to run will want to take the hit of running 100k
	// frames for nothing. So far the algorithms calling this function all
	// have reasons to assume that the player will get hit. If it's not the case
	// I should be warned.
	Check(fmt.Errorf("something went wrong"))
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

	// The player might have just been attacked and hit.
	if w.Player.JustHit {
		return 0
	}

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

func CurrentRankedActions(world World) (actions []Action) {
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

func RankedActionsPerFrame(playthrough Playthrough, frameIdx int) (actions []Action) {
	world := GoToFrame(playthrough, frameIdx)
	return CurrentRankedActions(world)
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

func ActionToInput(action Action) (input PlayerInput) {
	if action.Move {
		input.Move = true
		input.MovePt = action.Pos
	} else if !action.Move {
		input.Shoot = true
		input.ShootPt = action.Pos
	} else {
		panic(fmt.Errorf("bad"))
	}
	return
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
	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\training-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

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

func TestAIPlayer(t *testing.T) {
	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\training-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	inputFile := inputFiles[0]

	playthrough := DeserializePlaythrough(ReadFile(inputFile))

	world := NewWorld(playthrough.Seed, playthrough.Level)

	// Wait some period in the beginning.
	frameIdx := 0
	for ; frameIdx < 100; frameIdx++ {
		input := PlayerInput{}
		world.Step(input)
	}

	// Start moving every 30 frames (0.5 sec).
	for {
		input := PlayerInput{}
		if frameIdx%30 == 0 {
			rankedActions := CurrentRankedActions(world)
			action := rankedActions[0]
			input = ActionToInput(action)
		}

		world.Step(input)
		if world.Status() != Ongoing {
			break
		}
		frameIdx++
	}

	if world.Status() == Won {
		fmt.Println("ai player WON")
	} else {
		fmt.Println("ai player LOST")
	}

	WriteFile("outputs/ai-play.mln013", world.SerializedPlaythrough())

	// rankedActions := CurrentRankedActions(world, 100)
	//
	// for _, inputFile := range inputFiles {
	// 	modelFitness := ModelFitnessForPlaythrough(inputFile)
	// 	fmt.Printf("modelFitness: %d\n", modelFitness)
	// }
}

func GetHistogram(s []int) map[int]int {
	h := map[int]int{}
	for _, v := range s {
		h[v]++
	}
	return h
}

func OutputHistogram(histogram map[int]int, outputFile string, header string) {
	// Collect all the keys.
	keys := make([]int, 0)
	for k := range histogram {
		keys = append(keys, k)
	}

	// Sort the keys.
	slices.Sort(keys)

	// Print out the map, sorted by keys.
	f, err := os.Create(outputFile)
	Check(err)
	_, err = f.WriteString(fmt.Sprintf("%s\n", header))
	Check(err)
	for _, k := range keys {
		_, err = f.WriteString(fmt.Sprintf("%d,%d\n", k, histogram[k]))
		Check(err)
	}
	Check(f.Close())
}

// Gets a median of 21 frames between actions.
func TestGetReactionSpeed(t *testing.T) {
	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\training-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	diffs := []int{}
	for _, inputFile := range inputFiles {
		playthrough := DeserializePlaythrough(ReadFile(inputFile))
		framesWithActions := GetFramesWithActions(playthrough)
		for i := 1; i < len(framesWithActions); i++ {
			diff := framesWithActions[i] - framesWithActions[i-1]
			diffs = append(diffs, diff)
		}
	}

	histogram := GetHistogram(diffs)
	fmt.Println(histogram)
	OutputHistogram(histogram, "outputs/reaction-speed.csv", "n_frames_between_actions,n_occurrences")

	slices.Sort(diffs)
	median := diffs[len(diffs)/2]
	fmt.Println(median)
}

// Output:
// rank_of_action,n_occurrences
// 1,777
// 2,432
// 3,328
// 4,206
// 5,98
// 6,29
// 7,10
// 8,2
// 9,1
// 11,1
func TestGetHumanPlayerActionRanks(t *testing.T) {
	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\training-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	allRanks := []int{}
	for _, inputFile := range inputFiles {
		fmt.Println(inputFile)
		playthrough := DeserializePlaythrough(ReadFile(inputFile))
		framesWithActions := GetFramesWithActions(playthrough)
		decisionFrames := GetDecisionFrames(framesWithActions)
		ranksOfPlayerActions := GetRanksOfPlayerActions(playthrough, framesWithActions, decisionFrames)
		allRanks = append(allRanks, ranksOfPlayerActions...)
	}

	OutputHistogram(GetHistogram(allRanks), "outputs/action-ranks.csv", "rank_of_action,n_occurrences")

	assert.True(t, true)
}

type RandomnessInPlay struct {
	minNFramesBetweenActions int
	maxNFramesBetweenActions int
	weightOfRank1Action      int
	weightOfRank2Action      int
}

func PlayLevel(l Level, seed Int, r RandomnessInPlay) World {
	world := NewWorld(seed, l)

	// Wait some period in the beginning.
	frameIdx := 0
	for ; frameIdx < 100; frameIdx++ {
		input := PlayerInput{}
		world.Step(input)
	}

	// Start moving every 30 frames (0.5 sec).
	getFrameIdxOfNextMove := func(frameIdx int) int {
		return frameIdx + RInt(I(r.minNFramesBetweenActions), I(r.maxNFramesBetweenActions)).ToInt()
	}

	frameIdxOfNextMove := getFrameIdxOfNextMove(frameIdx)
	for {
		input := PlayerInput{}

		if frameIdx == frameIdxOfNextMove {
			rankedActions := CurrentRankedActions(world)

			action := rankedActions[0]
			// There is a random chance to degrade the quality of the
			// action based on weights.
			totalWeight := r.weightOfRank1Action + r.weightOfRank2Action
			randomNumber := RInt(I(1), I(totalWeight)).ToInt()
			if randomNumber > r.weightOfRank1Action {
				action = rankedActions[1]
			}

			input = ActionToInput(action)
			frameIdxOfNextMove = getFrameIdxOfNextMove(frameIdx)
		}

		world.Step(input)
		if world.Status() != Ongoing {
			return world
		}
		frameIdx++
	}
}

// As of 2025-06-04 it takes 2515.15s to run for 30 levels with
// nPlaysPerLevel := 10. That's 42 min and is a problem.
func TestAIPlayerMultiple(t *testing.T) {
	randomness := RandomnessInPlay{20, 40, 3, 1}
	nPlaysPerLevel := 10

	// dir := "d:\\Miln\\stored\\experiment2\\ai-output\\test-data"
	dir := "d:\\Miln\\stored\\experiment3\\ai-output\\test-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	f, err := os.Create("outputs/ai-plays.csv")
	Check(err)
	_, err = f.WriteString(fmt.Sprintf("health\n"))
	Check(err)

	consoleWriter := bufio.NewWriter(os.Stdout)
	for idx, inputFile := range inputFiles {
		fmt.Printf("%02d ", idx)
		Check(consoleWriter.Flush())
		playthrough := DeserializePlaythrough(ReadFile(inputFile))

		totalHealth := 0
		for i := 0; i < nPlaysPerLevel; i++ {
			world := PlayLevel(playthrough.Level, playthrough.Seed, randomness)
			WriteFile(fmt.Sprintf("outputs/ai-play-%02d.mln013", idx), world.SerializedPlaythrough())
			if world.Status() == Won {
				totalHealth += world.Player.Health.ToInt()
				fmt.Printf("win ")
			} else {
				fmt.Printf("loss ")
			}
			Check(consoleWriter.Flush())
		}
		health := float64(totalHealth) / float64(nPlaysPerLevel)
		fmt.Printf("health: %f\n", health)
		_, err = f.WriteString(fmt.Sprintf("%f\n", health))
		Check(err)
	}

	Check(f.Close())
}

// NeutralInput generates an input that doesn't do anything but has some values
// for the positions of the mouse. A simple PlayerInput{} would also be neutral
// but a list containing mostly PlayerInput{} values would zip and unzip very
// quickly and efficiently, and this is not representative of realistic
// conditions.
func NeutralInput() PlayerInput {
	return PlayerInput{
		MousePt:            Pt{RInt(I(0), I(1919)), RInt(I(0), I(1079))},
		LeftButtonPressed:  false,
		RightButtonPressed: false,
		Move:               false,
		MovePt:             Pt{RInt(I(0), I(7)), RInt(I(0), I(7))},
		Shoot:              false,
		ShootPt:            Pt{RInt(I(0), I(7)), RInt(I(0), I(7))},
	}
}

// PlayLevelForAtLeastNFrames will play a level for at least nFrames. It will
// attempt to generate a playthrough as close to nFrames as it can.
// It is useful to generate large playthroughs automatically in order to
// benchmark the performance of the simulation or to use in regression tests.
// However, the playthrough must be "interesting". It's easy to just not do the
// first move for n frames, but that isn't a good representation of the
// simulation, it doesn't trigger enough of the simulation code.
// The player should move around the map, not lose, but not win either.
// The way PlayLevelForAtLeastNFrames does it is that the player:
// - waits a little
// - appears on the map and waits until it gets hit
// - moves and kills enemies until there are only 3 left
// - moves around the map without attacking enemies, until all n frames are done
// - waits until it gets hit
// - moves and kills the rest of the enemies
// This works because the player moves each 20 frames which is quite fast.
// The enemies must move at a reasonable speed and it helps if there are some
// obstacles but not too many.
func PlayLevelForAtLeastNFrames(l Level, seed Int, nFrames int) World {
	w := NewWorld(seed, l)

	// Wait some period in the beginning.
	frameIdx := 0
	for ; frameIdx < 100; frameIdx++ {
		w.Step(NeutralInput())
	}

	// Move on the map.
	{
		rankedActions := CurrentRankedActions(w)
		w.Step(ActionToInput(rankedActions[0]))
	}

	// Fight until only 3 enemy is left.
	for {
		if frameIdx%20 == 0 {
			rankedActions := CurrentRankedActions(w)
			w.Step(ActionToInput(rankedActions[0]))
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return w
			}
			if len(w.Enemies) == 3 {
				break
			}
		} else {
			w.Step(NeutralInput())
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return w
			}
		}
		frameIdx++
	}

	// Wait until getting hit once.
	for {
		w.Step(NeutralInput())
		// After each world step, check if the game is over.
		if w.Status() != Ongoing {
			return w
		}
		if w.Player.JustHit {
			break
		}
	}

	// Move around without attacking.
	for ; frameIdx < nFrames; frameIdx++ {
		if frameIdx%20 == 0 {
			rankedActions := CurrentRankedActions(w)
			for _, r := range rankedActions {
				if r.Move {
					w.Step(ActionToInput(r))
					// After each world step, check if the game is over.
					if w.Status() != Ongoing {
						return w
					}
					break
				}
			}
		} else {
			w.Step(NeutralInput())
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return w
			}
		}
	}

	// Wait until getting hit once.
	for {
		w.Step(NeutralInput())
		// After each world step, check if the game is over.
		if w.Status() != Ongoing {
			return w
		}
		if w.Player.JustHit {
			break
		}
	}

	// Fight until all enemies are killed.
	for {
		if frameIdx%20 == 0 {
			rankedActions := CurrentRankedActions(w)
			w.Step(ActionToInput(rankedActions[0]))
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return w
			}
		} else {
			w.Step(NeutralInput())
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return w
			}
		}
		frameIdx++
	}
}

func TestGenerateLargePlaythrough(t *testing.T) {
	level := GenerateLevelFromParams(Param{I(5), I(90), I(8), I(4)})
	world := PlayLevelForAtLeastNFrames(level, I(0), 18000)
	fmt.Println(len(world.History))
	WriteFile("outputs/large-playthrough.mln016", world.SerializedPlaythrough())
}
