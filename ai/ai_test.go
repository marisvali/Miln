package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"os"
	"slices"
	"testing"
)

func GoToFrame(playthrough Playthrough, frameIdx int64) World {
	world := NewWorld(playthrough.Seed, playthrough.Level)
	for i := int64(0); i < frameIdx; i++ {
		input := playthrough.History.Data[i]
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
	for i := range world.Ammos.N {
		if world.Ammos.Data[i].Pos == pos {
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
	positions := startPositions.ToSlice()
	allVisible := MatBool{}
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
	lookoutPositions := MatBool{}
	lookoutPositions.Set(startPos)

	// Try a maximum of 10 moves for now, even though normally there should be
	// a better stop condition like "all possible positions have been visited".
	for iMove := 0; iMove < 10; iMove++ {
		visiblePositions := h.computeVisiblePositions(lookoutPositions)
		// Check if any targets are visible from the current start positions.
		visibleTargets := visiblePositions
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
	ammos := MatBool{}
	for i := range world.Ammos.N {
		ammos.Set(world.Ammos.Data[i].Pos)
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

type Action struct {
	Move    bool
	Pos     Pt
	Fitness int64
	Rank    int64
}

func (a Action) String() string {
	if a.Move {
		return fmt.Sprintf("rank: %3d fitness: %3d move:  %d %d", a.Rank, a.Fitness, a.Pos.X.ToInt(), a.Pos.Y.ToInt())
	} else {
		return fmt.Sprintf("rank: %3d fitness: %3d shoot: %d %d", a.Rank, a.Fitness, a.Pos.X.ToInt(), a.Pos.Y.ToInt())
	}
}

func RankedActionsPerFrame(playthrough Playthrough, frameIdx int64) {
	world := GoToFrame(playthrough, frameIdx)
	CurrentRankedActions(world)
}

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

func FindActionRank(action Action) int64 {
	for i := range rankedActions.N {
		if action.Pos == rankedActions.V[i].Pos &&
			action.Move == rankedActions.V[i].Move {
			return rankedActions.V[i].Rank
		}
	}
	panic(fmt.Errorf("bad"))
}

func GetFramesWithActions(playthrough Playthrough) (framesWithActions []int64) {
	for i := range playthrough.History.N {
		input := playthrough.History.Data[i]
		if input.Move || input.Shoot {
			framesWithActions = append(framesWithActions, i)
		}
	}
	return
}

func GetDecisionFrames(framesWithActions []int64) (decisionFrames []int64) {
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

func GetRanksOfPlayerActions(playthrough Playthrough, framesWithActions []int64, decisionFrames []int64) (ranksOfPlayerActions []int64) {
	for actionIdx := range framesWithActions {
		RankedActionsPerFrame(playthrough, decisionFrames[actionIdx])
		playerAction := InputToAction(playthrough.History.Data[framesWithActions[actionIdx]])
		rank := FindActionRank(playerAction)
		ranksOfPlayerActions = append(ranksOfPlayerActions, rank)
	}
	return
}

func ModelFitness(ranksOfPlayerActions []int64) (modelFitness int64) {
	for i := 0; i < len(ranksOfPlayerActions); i++ {
		diff := ranksOfPlayerActions[i] - 1
		modelFitness += diff * diff
	}
	return modelFitness
}

func DebugRank(framesWithActions []int64, decisionFrames []int64,
	ranksOfPlayerActions []int64, playthrough Playthrough, actionIdx int64) {
	fmt.Printf("frame with player action: %d\n", framesWithActions[actionIdx])
	fmt.Printf("decision frame: %d\n", decisionFrames[actionIdx])
	println(ranksOfPlayerActions[actionIdx])
	world := GoToFrame(playthrough, decisionFrames[actionIdx])
	fmt.Printf("%+v\n", InputToAction(playthrough.History.Data[framesWithActions[actionIdx]]))

	println("debugging now")
	println(FitnessOfAttackAction(&world, IPt(7, 4)))

	// Compute the fitness of every move action.
	for y := 0; y < NRows; y++ {
		for x := 0; x < NCols; x++ {
			fitness := FitnessOfMoveAction(&world, IPt(x, y))
			fmt.Printf("%3d ", fitness)
		}
		fmt.Printf("\n")
	}

	// Compute the fitness of every attack action.
	fmt.Printf("\n")
	for y := 0; y < NRows; y++ {
		for x := 0; x < NCols; x++ {
			fitness := FitnessOfAttackAction(&world, IPt(x, y))
			fmt.Printf("%3d ", fitness)
		}
		fmt.Printf("\n")
	}
}

func ModelFitnessForPlaythrough(inputFile string) int64 {
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
			CurrentRankedActions(world)
			action := rankedActions.V[0]
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

func GetHistogram(s []int64) map[int64]int64 {
	h := map[int64]int64{}
	for _, v := range s {
		h[v]++
	}
	return h
}

func OutputHistogram(histogram map[int64]int64, outputFile string, header string) {
	// Collect all the keys.
	keys := make([]int64, 0)
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

	diffs := []int64{}
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

	allRanks := []int64{}
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

func WorldState(frameIdx int, w *World) {
	pts := []Pt{}
	pts = append(pts, w.Player.Pos())

	for i := range w.Enemies.N {
		pts = append(pts, w.Enemies.Data[i].Pos())
	}

	fmt.Printf("%04d  ", frameIdx)
	for i := range pts {
		fmt.Printf("%02d %02d  ", pts[i].X.ToInt(), pts[i].Y.ToInt())
	}
	fmt.Println()
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
		CurrentRankedActions(w)
		w.Step(ActionToInput(rankedActions.V[0]))
	}

	// Fight until only 3 enemy is left.
	for {
		if frameIdx%20 == 0 {
			CurrentRankedActions(w)
			w.Step(ActionToInput(rankedActions.V[0]))
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return w
			}
			if w.Enemies.N == 3 {
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
			CurrentRankedActions(w)
			for i := range rankedActions.N {
				if rankedActions.V[i].Move {
					w.Step(ActionToInput(rankedActions.V[i]))
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
			CurrentRankedActions(w)
			w.Step(ActionToInput(rankedActions.V[0]))
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
	fmt.Println(world.History.N)
	WriteFile("outputs/large-playthrough.mln016", world.SerializedPlaythrough())
}

func TestGenerateAveragePlaythrough(t *testing.T) {
	level := GenerateLevelFromParams(Param{I(5), I(90), I(8), I(4)})
	world := PlayLevelForAtLeastNFrames(level, I(0), 2000)
	fmt.Println(world.History.N)
	WriteFile("outputs/average-playthrough.mln016", world.SerializedPlaythrough())
}
