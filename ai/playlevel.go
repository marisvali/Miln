package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

func PlayLevel(l Level, seed Int, r RandomnessInPlay, levelIdx int, playIdx int, debug bool) World {
	world := NewWorld(seed, l)

	// Wait some period in the beginning.
	frameIdx := 0
	for ; frameIdx < 100; frameIdx++ {
		input := PlayerInput{}
		world.Step(input)
	}

	// Start moving every 30 frames (0.5 sec).
	getFrameIdxOfNextMove := func(frameIdx int) int {
		return frameIdx + r.RInt(I(r.MinNFramesBetweenActions), I(r.MaxNFramesBetweenActions)).ToInt()
	}

	var rankedActions ActionsArray
	frameIdxOfNextMove := getFrameIdxOfNextMove(frameIdx)
	debugFile := fmt.Sprintf("outputs/ai-debug-%02d-%02d", levelIdx, playIdx)
	if debug {
		DeleteFile(debugFile)
	}
	for {
		input := PlayerInput{}

		if frameIdx == frameIdxOfNextMove {
			ComputeRankedActions(world, &rankedActions)

			action := rankedActions.V[0]
			// There is a random chance to degrade the quality of the
			// action based on weights.
			totalWeight := r.WeightOfRank1Action + r.WeightOfRank2Action
			randomNumber := r.RInt(I(1), I(totalWeight)).ToInt()
			if randomNumber > r.WeightOfRank1Action {
				action = rankedActions.V[1]
			}

			input = ActionToInput(action)
			frameIdxOfNextMove = getFrameIdxOfNextMove(frameIdx)
		}

		world.Step(input)
		if debug {
			str := ""
			if input.Move {
				str += fmt.Sprintf("move  %02d %02d  ",
					input.MovePt.X.ToInt(),
					input.MovePt.X.ToInt())
			} else if input.Shoot {
				str += fmt.Sprintf("shoot %02d %02d  ",
					input.ShootPt.X.ToInt(),
					input.ShootPt.X.ToInt())
			} else {
				str += fmt.Sprintf("             ")
			}
			AppendToFile(debugFile, str)
			AppendToFile(debugFile, world.StateStr()+"\n")
		}
		if world.Status() != Ongoing {
			return world
		}
		frameIdx++
	}
}

type RandomnessInPlay struct {
	Rand
	MinNFramesBetweenActions int
	MaxNFramesBetweenActions int
	WeightOfRank1Action      int
	WeightOfRank2Action      int
}
