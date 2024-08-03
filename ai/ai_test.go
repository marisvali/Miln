package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"os"
	"path/filepath"
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

func ComputeMeanSquaredError(expectedOutcome []Int, actualOutcome []Int) (error float64) {
	if len(expectedOutcome) != len(actualOutcome) {
		Check(fmt.Errorf("expected equal lengths, got %d %d",
			len(expectedOutcome), len(actualOutcome)))
	}

	sum := ZERO
	for i := range expectedOutcome {
		sum.Add(expectedOutcome[i].Minus(actualOutcome[i]).Sqr())
		// fmt.Printf("%d %d %d\n", expectedOutcome[i].ToInt(), actualOutcome[i].ToInt(), sum.ToInt())
	}
	error = sum.ToFloat64() / float64(len(expectedOutcome))
	return
}

func RunLevelWithAI(seed Int) (playerHealth Int) {
	w := NewWorld(seed)
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
	w := NewWorld(p.Seed)
	for _, input := range p.History {
		w.Step(input)
	}
	playerHealth = w.Player.Health
	isGameOver = IsGameOver(&w)
	return
}

func ComputeMeanSquaredErrorOnDataset(dir string, odds bool) float64 {
	entries, err := os.ReadDir(dir)
	Check(err)

	expectedOutcomes := []Int{}
	actualOutcomes := []Int{}
	for i, entry := range entries {
		// Skip evens if we want only odds.
		// Skip odds if we want only evens.
		isOdd := (i+1)%2 == 1
		if odds != isOdd {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		data := ReadFile(fullPath)
		playthrough := DeserializePlaythrough(data)
		expectedOutcome, isGameOver := RunPlaythrough(playthrough)
		if !isGameOver {
			continue
		}
		// fmt.Printf("%d - %s - seed %d", i, entry.Name(), playthrough.Seed)
		// fmt.Printf(" - expected outcome %d", expectedOutcome)
		expectedOutcomes = append(expectedOutcomes, expectedOutcome)
		actualOutcome := RunLevelWithAI(playthrough.Seed)
		// fmt.Printf(" - actual outcome %d\n", actualOutcome)
		actualOutcomes = append(actualOutcomes, actualOutcome)
	}

	meanSquaredError := ComputeMeanSquaredError(expectedOutcomes, actualOutcomes)

	// for _, e := range actualOutcomes {
	// 	fmt.Println(e.ToInt())
	// }
	// fmt.Printf("mean squared error: %f\n", meanSquaredError)
	return meanSquaredError
}

func TestAI_MeanSquaredError(t *testing.T) {
	dir := "d:\\gms\\Miln\\analysis\\2024-08-03 - data\\data-set-5\\playthroughs"
	// for i := 20; i <= 40; i++ {
	// 	MinFramesBetweenActions = i
	// 	meanSquaredError := ComputeMeanSquaredErrorOnDataset(dir, true)
	// 	fmt.Printf("%d\t= %f\n", MinFramesBetweenActions, meanSquaredError)
	// }
	MinFramesBetweenActions = 30
	meanSquaredError := ComputeMeanSquaredErrorOnDataset(dir, false)
	fmt.Printf("%d\t= %f\n", MinFramesBetweenActions, meanSquaredError)

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
