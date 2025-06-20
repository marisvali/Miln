package ai

import (
	"cmp"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"slices"
)

type Action struct {
	Move    bool
	Pos     Pt
	Fitness int64
	Rank    int64
}

func ModelFitness(ranksOfPlayerActions []int64) (modelFitness int64) {
	for i := 0; i < len(ranksOfPlayerActions); i++ {
		diff := ranksOfPlayerActions[i] - 1
		modelFitness += diff * diff
	}
	return modelFitness
}

func ModelFitnessForPlaythrough(inputFile string) int64 {
	playthrough := DeserializePlaythrough(ReadFile(inputFile))
	framesWithActions := GetFramesWithActions(playthrough)
	decisionFrames := GetDecisionFrames(framesWithActions)
	ranksOfPlayerActions := GetRanksOfPlayerActions(playthrough, framesWithActions, decisionFrames)
	modelFitness := ModelFitness(ranksOfPlayerActions)
	return modelFitness
}

type ActionsArray struct {
	N int64
	V [NRows * NCols * 2]Action
}

func ComputeRankedActions(world World, rankedActions *ActionsArray) {
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
			fitness := FitnessOfAttackAction(world, IPt(x, y))
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

func GoToFrame(playthrough Playthrough, frameIdx int64) World {
	world := NewWorld(playthrough.Seed, playthrough.Level)
	for i := int64(0); i < frameIdx; i++ {
		input := playthrough.History[i]
		world.Step(input)
	}
	return world
}

func (a Action) String() string {
	if a.Move {
		return fmt.Sprintf("rank: %3d fitness: %3d move:  %d %d", a.Rank, a.Fitness, a.Pos.X.ToInt(), a.Pos.Y.ToInt())
	} else {
		return fmt.Sprintf("rank: %3d fitness: %3d shoot: %d %d", a.Rank, a.Fitness, a.Pos.X.ToInt(), a.Pos.Y.ToInt())
	}
}

func ComputeRankedActionsPerFrame(playthrough Playthrough, frameIdx int64, rankedActions *ActionsArray) {
	world := GoToFrame(playthrough, frameIdx)
	ComputeRankedActions(world, rankedActions)
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

func FindActionRank(action Action, rankedActions *ActionsArray) int64 {
	for i := range rankedActions.N {
		if action.Pos == rankedActions.V[i].Pos &&
			action.Move == rankedActions.V[i].Move {
			return rankedActions.V[i].Rank
		}
	}
	panic(fmt.Errorf("bad"))
}

func DebugRank(framesWithActions []int64, decisionFrames []int64,
	ranksOfPlayerActions []int64, playthrough Playthrough, actionIdx int64) {
	fmt.Printf("frame with player action: %d\n", framesWithActions[actionIdx])
	fmt.Printf("decision frame: %d\n", decisionFrames[actionIdx])
	println(ranksOfPlayerActions[actionIdx])
	world := GoToFrame(playthrough, decisionFrames[actionIdx])
	fmt.Printf("%+v\n", InputToAction(playthrough.History[framesWithActions[actionIdx]]))

	println("debugging now")
	println(FitnessOfAttackAction(world, IPt(7, 4)))

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
			fitness := FitnessOfAttackAction(world, IPt(x, y))
			fmt.Printf("%3d ", fitness)
		}
		fmt.Printf("\n")
	}
}

func GetRanksOfPlayerActions(playthrough Playthrough, framesWithActions []int64, decisionFrames []int64) (ranksOfPlayerActions []int64) {
	var rankedActions ActionsArray
	for actionIdx := range framesWithActions {
		ComputeRankedActionsPerFrame(playthrough, decisionFrames[actionIdx], &rankedActions)
		playerAction := InputToAction(playthrough.History[framesWithActions[actionIdx]])
		rank := FindActionRank(playerAction, &rankedActions)
		ranksOfPlayerActions = append(ranksOfPlayerActions, rank)
	}
	return
}
