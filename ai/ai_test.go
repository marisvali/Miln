package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func IsGameWon(w *World) bool {
	for _, enemy := range w.Enemies {
		if enemy.Alive() {
			return false
		}
	}
	for _, portal := range w.SpawnPortals {
		if portal.Active() {
			return false
		}
	}
	return true
}

func IsGameLost(w *World) bool {
	return w.Player.Health.Leq(ZERO)
}

func IsGameOver(w *World) bool {
	return IsGameWon(w) || IsGameLost(w)
}

func ComputeMeanSquaredError(expectedOutcome []float64, actualOutcome []float64) (error float64) {
	if len(expectedOutcome) != len(actualOutcome) {
		Check(fmt.Errorf("expected equal lengths, got %d %d",
			len(expectedOutcome), len(actualOutcome)))
	}

	sum := float64(0)
	for i := range expectedOutcome {
		dif := expectedOutcome[i] - actualOutcome[i]
		sum += dif * dif
		// fmt.Printf("%d %d %d\n", expectedOutcome[i].ToInt(), actualOutcome[i].ToInt(), sum.ToInt())
	}
	error = sum / float64(len(expectedOutcome))
	return
}

func RunLevelWithAI(seed, targetDifficulty Int) (playerHealth Int) {
	w := NewWorld(seed, targetDifficulty)
	ai := AI{}
	for {
		input := ai.Step(&w)
		w.Step(input)
		if IsGameOver(&w) {
			break
		}
	}
	playerHealth = w.Player.Health
	return
}

func RunPlaythrough(p Playthrough) (playerHealth Int, isGameOver bool) {
	w := NewWorld(p.Seed, p.TargetDifficulty)
	for _, input := range p.History {
		w.Step(input)
	}
	playerHealth = w.Player.Health
	isGameOver = IsGameOver(&w)
	return
}

func ComputeMeanSquaredErrorOnDataset2(dir string, files []string) (actualOutcomes []float64) {
	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		data := ReadFile(fullPath)
		playthrough := DeserializePlaythrough(data)
		_, isGameOver := RunPlaythrough(playthrough)
		if !isGameOver {
			Check(fmt.Errorf("not cool"))
		}
		fmt.Printf("%d", playthrough.TargetDifficulty.ToInt())
		actualOutcome := RunLevelWithAI(playthrough.Seed, playthrough.TargetDifficulty)
		fmt.Printf(", %d\n", actualOutcome.ToInt())
		actualOutcomes = append(actualOutcomes, actualOutcome.ToFloat64())
	}

	return actualOutcomes
}

func TestAI_MeanSquaredError(t *testing.T) {
	dir := "d:\\gms\\Miln\\analysis\\2024-08-04 - 4. how do AI and statistics perform when the same levels are played multiple times\\data-set-6\\playthroughs"
	// level difficulty, file
	// 52, 20240804-064317.mln007
	// 54, 20240804-064006.mln007
	// 56, 20240804-063847.mln007
	// 58, 20240804-063509.mln007
	// 60, 20240804-064120.mln007
	// 62, 20240804-063624.mln007
	// 64, 20240804-064422.mln007
	// 66, 20240804-064228.mln007
	// 68, 20240804-063754.mln007
	// 70, 20240804-063920.mln007

	all := []string{
		"20240804-064317.mln007",
		"20240804-064006.mln007",
		"20240804-063847.mln007",
		"20240804-063509.mln007",
		"20240804-064120.mln007",
		"20240804-063624.mln007",
		"20240804-064422.mln007",
		"20240804-064228.mln007",
		"20240804-063754.mln007",
		"20240804-063920.mln007"}

	allExpectedOutcomes := []float64{
		0.666666667,
		2.6,
		1.333333333,
		0.8,
		1.333333333,
		1,
		0.333333333,
		0.2,
		0,
		0.2}

	allActualOutcomes := [][]float64{}
	for i := 25; i <= 31; i++ {
		MinFramesBetweenActions = i
		actualOutcomes := ComputeMeanSquaredErrorOnDataset2(dir, all)
		allActualOutcomes = append(allActualOutcomes, actualOutcomes)
	}

	sumsOutcomes := make([]float64, len(allActualOutcomes[0]))
	for i := range allActualOutcomes {
		for j := range allActualOutcomes[i] {
			sumsOutcomes[j] += allActualOutcomes[i][j]
		}
	}
	avgOutcomes := make([]float64, len(sumsOutcomes))
	for i := range sumsOutcomes {
		avgOutcomes[i] = sumsOutcomes[i] / float64(len(allActualOutcomes))
	}
	for i := range avgOutcomes {
		fmt.Printf("%f\n", avgOutcomes[i])
	}

	meanSquaredError := ComputeMeanSquaredError(allExpectedOutcomes, avgOutcomes)
	fmt.Printf("%f\n", meanSquaredError)

	assert.True(t, true)
}

func BoolToInt(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

func TestAI_PlayerStats(t *testing.T) {
	inputFilename := "d:\\gms\\Miln\\analysis\\2024-07-29 - set benchmark for AI\\data-set-1\\playthroughs\\20240709-112511.mln002"

	playthrough := DeserializePlaythrough(ReadFile(inputFilename))
	// Create a new CSV file
	outFile, err := os.Create("output.csv")
	Check(err)
	defer CloseFile(outFile)

	_, err = outFile.WriteString("frame_idx,moved,shot\n")
	Check(err)
	for frameIdx, input := range playthrough.History {
		if input.Move || input.Shoot {

			_, err = outFile.WriteString(fmt.Sprintf("%d,%d,%d\n", frameIdx, BoolToInt(input.Move), BoolToInt(input.Shoot)))
			Check(err)
		}
	}
	assert.True(t, true)
}

func TestAI_GeneratePlaySequence(t *testing.T) {
	originalSequence := []int{}
	for i := 52; i <= 70; i = i + 2 {
		originalSequence = append(originalSequence, i)
	}

	finalSequence := []int{}
	for i := 0; i < 10; i++ {
		s := slices.Clone(originalSequence)
		rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
		finalSequence = append(finalSequence, s...)
	}

	// seed := 15
	content := ""
	for i := range finalSequence {
		line := fmt.Sprintf("%d %d\n", finalSequence[i]*3+7, finalSequence[i])
		content = content + line
	}
	filename := "play-sequence.txt"
	WriteFile(filename, []byte(content))
	assert.True(t, true)
}
