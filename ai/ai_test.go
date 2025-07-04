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
	playthrough := PlayLevelForAtLeastNFrames(level, I(0), 18000)
	fmt.Println(len(playthrough.History))
	WriteFile("outputs/large-playthrough.mln17-17", playthrough.Serialize())
}

func TestGenerateAveragePlaythrough(t *testing.T) {
	level := GenerateLevelFromParams(Param{I(5), I(90), I(8), I(4)})
	playthrough := PlayLevelForAtLeastNFrames(level, I(0), 2000)
	fmt.Println(len(playthrough.History))
	WriteFile("outputs/average-playthrough.mln17-17", playthrough.Serialize())
}
