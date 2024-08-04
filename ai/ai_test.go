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

func RunLevelWithAI(seed Int, targetDifficulty Int) (playerHealth Int) {
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

func RunPlaythrough(p Playthrough) (playerHealth Int) {
	w := NewWorld(p.Seed, p.TargetDifficulty)
	for _, input := range p.History {
		w.Step(input)
	}
	playerHealth = w.Player.Health
	return
}

func TestAI_Step(t *testing.T) {
	dir := "d:\\gms\\Miln\\analysis\\2024-07-29 - set benchmark for AI\\data-set-1\\playthroughs"
	entries, err := os.ReadDir(dir)
	Check(err)

	expectedOutcomes := []Int{}
	actualOutcomes := []Int{}
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		data := ReadFile(fullPath)
		playthrough := DeserializePlaythrough(data)
		expectedOutcome := RunPlaythrough(playthrough)
		expectedOutcomes = append(expectedOutcomes, expectedOutcome)
		actualOutcome := RunLevelWithAI(playthrough.Seed, playthrough.TargetDifficulty)
		actualOutcomes = append(actualOutcomes, actualOutcome)
	}

	meanSquaredError := ComputeMeanSquaredError(expectedOutcomes, actualOutcomes)
	fmt.Printf("mean squared error: %f", meanSquaredError)
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
