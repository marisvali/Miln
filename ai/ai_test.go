package ai

import (
	"bufio"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"os"
	"slices"
	"testing"
)

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

func TestStuff(t *testing.T) {
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
	for _, inputFile := range inputFiles {
		Check(consoleWriter.Flush())
		playthrough := DeserializePlaythroughFromOld(ReadFile(inputFile))

		totalHealth := 0
		for i := 0; i < nPlaysPerLevel; i++ {
			world := PlayLevel(playthrough.Level, playthrough.Seed, randomness)
			// WriteFile(fmt.Sprintf("outputs/ai-play-opt-%02d-%02d.mln016-opt", idx, i), world.SerializedPlaythrough())
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
