package main

import (
	"fmt"
	. "github.com/marisvali/miln/ai"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"sync"
)

func DoIt(inputFile string, nPlaysPerLevel int, randomness RandomnessInPlay, idx int) (health float64) {
	playthrough := DeserializePlaythroughFromOld(ReadFile(inputFile))
	totalHealth := 0
	// debug := idx == 13
	debug := false
	for i := 0; i < nPlaysPerLevel; i++ {
		randomness.RSeed(I(i))
		world := PlayLevel(playthrough.Level, playthrough.Seed, randomness, idx, i, debug)
		// WriteFile(fmt.Sprintf("outputs/ai-play-opt-%02d-%02d.mln016-opt", idx, i), world.SerializedPlaythrough())
		if world.Status() == Won {
			totalHealth += world.Player.Health.ToInt()
			fmt.Printf("win ")
		} else {
			fmt.Printf("loss ")
		}
	}
	health = float64(totalHealth) / float64(nPlaysPerLevel)
	fmt.Printf("health: %f\n", health)
	return
}

func DoItAll() {
	// debug.SetGCPercent(-1)
	RSeed(I(0))
	nPlaysPerLevel := 10
	randomness := RandomnessInPlay{NewRand(I(0)), 20, 40, 3, 1}
	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\test-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	healths := make([]float64, len(inputFiles))

	var wg sync.WaitGroup
	for idx, inputFile := range inputFiles {
		wg.Add(1)
		go func() {
			defer wg.Done()
			healths[idx] = DoIt(inputFile, nPlaysPerLevel, randomness, idx)
		}()
	}
	wg.Wait()

	f, err := os.Create("outputs/ai-plays.csv")
	Check(err)
	_, err = f.WriteString(fmt.Sprintf("health\n"))
	Check(err)
	for _, health := range healths {
		_, err = f.WriteString(fmt.Sprintf("%f\n", health))
		Check(err)
	}
	Check(f.Close())
}

// As of 2025-06-04 it takes 2515.15s to run for 30 levels with
// nPlaysPerLevel := 10. That's 42 min and is a problem.
func main() {
	DoItAll()
}
