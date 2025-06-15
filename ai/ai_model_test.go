package ai

import (
	"bufio"
	"cmp"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"slices"
	"testing"
)

// NumFramesUntilAttacked returns the number of frames until an enemy attacks pos.
// 0 - there's an enemy at this position right now
// 1 - after 1 frame, the player will be attacked
// ..
// -1 - the position will never be attacked (e.g. there are no more enemies)
func NumFramesUntilAttacked(world *World, pos Pt) int64 {
	// So far the algorithms calling this function all have reasons to assume
	// that the player will get hit. If it's not the case
	// I should be warned.
	if world.Enemies.N == 0 {
		Check(fmt.Errorf("something went wrong"))
		return -1
	}

	// Clone the world as we're going to modify it.
	w := *world

	// Put the player at pos so that enemies will come to it.
	w.Player.SetPos(pos)
	w.Player.OnMap = true
	w.Player.JustHit = false

	// If an enemy doesn't hit within 100k frames, it's not happening.
	input := PlayerInput{} // Don't move, don't attack.
	for frameIdx := int64(0); frameIdx < 100000; frameIdx++ {
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

type RandomnessInPlay struct {
	minNFramesBetweenActions int
	maxNFramesBetweenActions int
	weightOfRank1Action      int
	weightOfRank2Action      int
}

// FitnessOfMoveAction returns the fitness of a specific move action at a
// certain moment. The action is "the player moves to pos".
func FitnessOfMoveAction(world *World, pos Pt) int64 {
	if !ValidMove(world, pos) {
		// If the action isn't even valid, the fitness of the action is zero.
		return 0
	}
	// The action is valid so other factors come into play.

	// Compute how safe the position is.
	safetyFitness := int64(0)
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

	ammoFitness := int64(0)
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

	enemyFitness := int64(0)
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
func FitnessOfAttackAction(world *World, pos Pt) int64 {
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
	w := *world
	w.Step(input)

	// The player might have just been attacked and hit.
	if w.Player.JustHit {
		return 0
	}

	if w.Enemies.N == 0 {
		// Just won the game.
		return 1000
	}

	// Compute how safe the current position is.
	safetyFitness := int64(0)
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
		safetyFitness = 20 + int64(timeUntilAttacked*3)
	}

	if safetyFitness < 10 {
		// If the current position is too dangerous, the fitness is zero.
		return 0
	}

	// Prefer attacking to moving, if it's safe and possible.
	biasForAttacking := int64(20)

	return safetyFitness + biasForAttacking
}

type ActionsArray struct {
	N int64
	V [NRows * NCols * 2]Action
}

var rankedActions ActionsArray

func CurrentRankedActions(world World) {
	rankedActions.N = 0

	// Compute the fitness of every move action.
	for y := 0; y < NRows; y++ {
		for x := 0; x < NCols; x++ {
			fitness := FitnessOfMoveAction(&world, IPt(x, y))
			action := Action{}
			action.Move = true
			action.Pos = IPt(x, y)
			action.Fitness = fitness
			rankedActions.V[rankedActions.N] = action
			rankedActions.N++
		}
	}

	// Compute the fitness of every attack action.
	for y := 0; y < NRows; y++ {
		for x := 0; x < NCols; x++ {
			fitness := FitnessOfAttackAction(&world, IPt(x, y))
			action := Action{}
			action.Move = false
			action.Pos = IPt(x, y)
			action.Fitness = fitness
			rankedActions.V[rankedActions.N] = action
			rankedActions.N++
		}
	}

	// Sort.
	// It's important to use a stable sort in order to get repeatable results,
	// especially for regression purposes.
	cmpActions := func(a, b Action) int {
		return cmp.Compare(b.Fitness, a.Fitness)
	}
	slices.SortStableFunc(rankedActions.V[0:rankedActions.N], cmpActions)

	// Rank.
	rankedActions.V[0].Rank = 1
	for i := int64(1); i < rankedActions.N; i++ {
		if rankedActions.V[i].Fitness == rankedActions.V[i-1].Fitness {
			rankedActions.V[i].Rank = rankedActions.V[i-1].Rank
		} else {
			rankedActions.V[i].Rank = rankedActions.V[i-1].Rank + 1
		}
	}
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
			CurrentRankedActions(world)

			action := rankedActions.V[0]
			// There is a random chance to degrade the quality of the
			// action based on weights.
			totalWeight := r.weightOfRank1Action + r.weightOfRank2Action
			randomNumber := RInt(I(1), I(totalWeight)).ToInt()
			if randomNumber > r.weightOfRank1Action {
				action = rankedActions.V[1]
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
	RSeed(I(0))
	randomness := RandomnessInPlay{20, 40, 3, 1}
	nPlaysPerLevel := 10

	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\test-data"
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
		Check(consoleWriter.Flush())
		playthrough := DeserializePlaythroughFromOld(ReadFile(inputFile))

		totalHealth := 0
		for i := 0; i < nPlaysPerLevel; i++ {
			world := PlayLevel(playthrough.Level, playthrough.Seed, randomness)
			WriteFile(fmt.Sprintf("outputs/ai-play-opt-%02d-%02d.mln016-opt", idx, i), world.SerializedPlaythrough())
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
