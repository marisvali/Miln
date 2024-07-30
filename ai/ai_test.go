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

func RunPlaythrough(p Playthrough) (playerHealth Int) {
	w := NewWorld(p.Seed)
	for _, input := range p.History {
		w.Step(input)
	}
	playerHealth = w.Player.Health
	return
}

func TestAI_MeanSquaredError(t *testing.T) {
	dir := "d:\\gms\\Miln\\analysis\\2024-07-29 - set benchmark for AI\\data-set-1\\playthroughs"
	entries, err := os.ReadDir(dir)
	Check(err)

	expectedOutcomes := []Int{}
	actualOutcomes := []Int{}
	for i, entry := range entries {
		if i >= 3 {
			break
		}
		fmt.Printf("%d\n", i)
		fullPath := filepath.Join(dir, entry.Name())
		data := ReadFile(fullPath)
		playthrough := DeserializePlaythrough(data)
		expectedOutcome := RunPlaythrough(playthrough)
		expectedOutcomes = append(expectedOutcomes, expectedOutcome)
		fmt.Printf("%d etc\n", i)
		actualOutcome := RunLevelWithAI(playthrough.Seed)
		actualOutcomes = append(actualOutcomes, actualOutcome)
	}

	meanSquaredError := ComputeMeanSquaredError(expectedOutcomes, actualOutcomes)
	fmt.Printf("mean squared error: %f\n", meanSquaredError)
	assert.True(t, true)
}
