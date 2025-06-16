package main

import (
	"bufio"
	"fmt"
	. "github.com/marisvali/miln/ai"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"runtime/debug"
)

// As of 2025-06-04 it takes 2515.15s to run for 30 levels with
// nPlaysPerLevel := 10. That's 42 min and is a problem.
func main() {
	debug.SetGCPercent(-1)
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
