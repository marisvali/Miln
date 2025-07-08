package main

import (
	"fmt"
	. "github.com/marisvali/miln/ai"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"sync"
	"time"
)

func DoIt(inputFile string, nPlaysPerLevel int, randomness RandomnessInPlay, idx int) (health float64) {
	playthrough := DeserializePlaythrough(ReadFile(inputFile))
	totalHealth := 0
	// debug := idx == 13
	debug := false
	for i := 0; i < nPlaysPerLevel; i++ {
		randomness.RSeed(I(i))
		world := PlayLevel(playthrough.Level, playthrough.Seed, randomness, idx, i, debug)
		// WriteFile(fmt.Sprintf("outputs/ai-play-opt-%02d-%02d.mln016-opt", idx, i), world.SerializedPlaythrough())
		if world.Status() == Won {
			totalHealth += world.Player.Health.ToInt()
			// fmt.Printf("win ")
		} else {
			// fmt.Printf("loss ")
		}
	}
	health = float64(totalHealth) / float64(nPlaysPerLevel)
	// fmt.Printf("health: %f\n", health)
	return
}

func RunBatch(inputFiles []string, nPlaysPerLevel int, randomness RandomnessInPlay) (healths []float64) {
	healths = make([]float64, len(inputFiles))
	for i := range healths {
		healths[i] = -1
	}

	var wg sync.WaitGroup
	for idx, inputFile := range inputFiles {
		wg.Add(1)
		go func() {
			defer wg.Done()
			healths[idx] = DoIt(inputFile, nPlaysPerLevel, randomness, idx)
		}()
	}
	wg.Wait()
	return
}

func Fitness(expected []float64, result []float64) float64 {
	sum := float64(0)
	for i := range expected {
		diff := expected[i] - result[i]
		sum += diff * diff
	}
	fitness := sum / float64(len(expected))
	return fitness
}

func DoItAll() {
	// expected := []float64{0, 1, 0, 3, 3, 1, 2, 2, 2, 3, 2, 1, 0, 2, 2, 2, 2, 3,
	// 	3, 3, 1, 2, 3, 1, 2, 0, 3, 2, 2, 0, 2, 0, 2, 2, 3, 3, 1, 0, 3, 0, 3, 2,
	// 	3, 2, 1, 1, 0, 0, 1, 1, 3, 3, 1, 0, 3, 2, 0, 3, 2, 3, 3, 1, 3, 2, 0, 1,
	// 	2, 1, 3, 2, 1, 2, 2, 3, 2, 1, 3, 0, 3, 3, 0, 1, 3, 3, 2, 2, 2, 0, 2, 1,
	// 	1, 2, 0, 1, 1, 1, 3, 3, 2, 1}

	// debug.SetGCPercent(-1)
	// RSeed(I(0))
	// nPlaysPerLevel := 10

	// dir := "d:\\Miln\\stored\\experiment3\\output\\"
	// inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln017-017")
	// for idx := range inputFiles {
	// 	inputFiles[idx] = dir + inputFiles[idx][1:]
	// }

	// inputFilesSubset := make([]string, len(expected)/2)
	// expectedSubset := make([]float64, len(expected)/2)
	// for i := 0; i < len(expected)/2; i++ {
	// 	inputFilesSubset[i] = inputFiles[i*2]
	// 	expectedSubset[i] = expected[i*2]
	// }

	// // 0.9396
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 3, 1})))
	//
	// // 0.9658
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 100, 1})))
	//
	// // 0.8548
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	// // 0.9743
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 5,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 1, 1})))

	// 0.9023
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 5,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	// // 0.7481
	// fmt.Println(Fitness(expected, RunBatch(inputFiles, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 3, 1})))

	// // 1.0094
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 1, 1})))

	// // 1.5180
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 60, 2, 1})))

	// // 0.8662
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 2, 1})))

	// // 1.1128
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 10, 30, 2, 1})))

	// // 1.0100
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 40, 5, 3})))

	// // 0.9435
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 25, 35, 5, 3})))

	// // 1.0008
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 5, 3})))

	// // 0.8072
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 20,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	// // 0.9341
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 20,
	// 	RandomnessInPlay{NewRand(I(0)), 10, 50, 2, 1})))

	// // 0.8369
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 30,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	// // 0.9907
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 20,
	// 	RandomnessInPlay{NewRand(I(0)), 15, 30, 2, 1})))

	// // 0.8951
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 50,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	// // 0.8652
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 100,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	// // 1.08
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 1,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 2, 1})))

	// // 0.8288
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 3,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 2, 1})))

	// 30, 30, 10, 1
	// -------------
	// // 1.36
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 1,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))

	// // 1.225
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 2,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))

	// // 1.0422
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 3,
	// 	RandomnessInPlay{NewRand(I(3)), 30, 30, 10, 1})))

	// // 0.985
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 4,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))
	//
	// // 1.0199
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 5,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))

	// // 0.9544
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(3)), 30, 30, 10, 1})))

	// // 0.9551
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 20,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))

	// // 0.9670
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 30,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))

	// // 0.9751
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 40,
	// 	RandomnessInPlay{NewRand(I(0)), 30, 30, 10, 1})))

	// // 0.985016
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 50,
	// 	RandomnessInPlay{NewRand(I(3)), 30, 30, 10, 1})))

	// // 0.9764
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 100,
	// 	RandomnessInPlay{NewRand(I(3)), 30, 30, 10, 1})))

	// -------------------------------------------------------------------------
	// 1	1.080000
	// 2	0.900000
	// 3	0.828889
	// 4	0.935000
	// 5	0.961600
	// 10	0.866200
	// 20	0.940650
	// 30	0.927067
	// 40	0.908713
	// 50	0.912816
	// 100	0.926368
	// for _, nPlaysPerLevel := range []int{1, 2, 3, 4, 5, 10, 20, 30, 40, 50, 100} {
	// 	fmt.Printf("%f ", Fitness(expectedSubset, RunBatch(inputFilesSubset,
	// 		nPlaysPerLevel, RandomnessInPlay{NewRand(I(0)), 30, 30, 2, 1})))
	// }

	// 1	1.340000
	// 2	1.290000
	// 3	1.060000
	// 4	0.993750
	// 5	0.902400
	// 10	0.854800
	// 20	0.807250
	// 30 	0.836956
	// 40	0.884763
	// 50	0.895120
	// 100	0.865230
	// for _, nPlaysPerLevel := range []int{1, 2, 3, 4, 5, 10, 20, 30, 40, 50, 100} {
	// 	fmt.Printf("%f\n", Fitness(expectedSubset, RunBatch(inputFilesSubset,
	// 		nPlaysPerLevel, RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))
	// }

	// // 0.9396
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 3, 1})))

	// 0.8548
	// fmt.Println(Fitness(expectedSubset, RunBatch(inputFilesSubset, 10,
	// 	RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1})))

	RSeed(I(0))

	dir := "d:\\Miln\\stored\\experiment3\\output\\"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln017-017")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	inputFilesTraining := make([]string, len(inputFiles)/2)
	inputFilesTest := make([]string, len(inputFiles)/2)
	for i := 0; i < len(inputFiles); i++ {
		if i%2 == 0 {
			inputFilesTraining[i/2] = inputFiles[i]
		} else {
			inputFilesTest[i/2] = inputFiles[i]
		}
	}

	nPlaysPerLevel := 10
	r := RandomnessInPlay{NewRand(I(0)), 20, 40, 2, 1}

	healthsTraining := RunBatch(inputFilesTraining, nPlaysPerLevel, r)

	f, err := os.Create("outputs/ai-plays-training.csv")
	Check(err)
	_, err = f.WriteString(fmt.Sprintf("health\n"))
	Check(err)
	for i := range healthsTraining {
		_, err = f.WriteString(fmt.Sprintf("%s,%f\n", inputFilesTraining[i], healthsTraining[i]))
		Check(err)
	}
	Check(f.Close())

	healthsTest := RunBatch(inputFilesTest, nPlaysPerLevel, r)

	f, err = os.Create("outputs/ai-plays-test.csv")
	Check(err)
	_, err = f.WriteString(fmt.Sprintf("health\n"))
	Check(err)
	for i := range healthsTest {
		_, err = f.WriteString(fmt.Sprintf("%s,%f\n", inputFilesTest[i], healthsTest[i]))
		Check(err)
	}
	Check(f.Close())
}

// As of 2025-06-04 it takes 2515.15s to run for 30 levels with
// nPlaysPerLevel := 10. That's 42 min and is a problem.
// As of 2025-07-07 it takes 50.27s to run 100 levels with
// nPlaysPerLevel := 10. That's 15.08s per 30 levels. That's 166 times faster.
// It is still annoying but I can work with it.
func main() {
	start := time.Now()
	DoItAll()
	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("Duration: %.2f seconds\n", duration.Seconds())
}
