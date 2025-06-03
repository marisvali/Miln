package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type PlayedLevel struct {
	Name            string
	PlayIdx         int
	PlaythroughName string
}

func (p *PlayedLevel) IsTrainingLevel() bool {
	return strings.Contains(p.Name, "training")
}

func (p *PlayedLevel) IsTestLevel() bool {
	return strings.Contains(p.Name, "test")
}

func (p *PlayedLevel) ParamsFilename() string {
	return fmt.Sprintf("%s.mln013-params", p.Name)
}

func (p *PlayedLevel) PlayParamsFilename() string {
	return fmt.Sprintf("%02d-%s.mln013-params", p.PlayIdx, p.Name)
}

func (p *PlayedLevel) LevelFilename() string {
	return fmt.Sprintf("%s.mln013-level", p.Name)
}

func (p *PlayedLevel) PlayLevelFilename() string {
	return fmt.Sprintf("%02d-%s.mln013-level", p.PlayIdx, p.Name)
}

func (p *PlayedLevel) SourcePlaythroughFilename() string {
	return fmt.Sprintf("%s.mln013", p.PlaythroughName)
}

func (p *PlayedLevel) PlayPlaythroughFilename() string {
	return fmt.Sprintf("%02d-%s.mln013", p.PlayIdx, p.Name)
}

func (p *PlayedLevel) PlaythroughFilename() string {
	return fmt.Sprintf("%s.mln013", p.Name)
}

// Code used to extract data from playthroughs for experiment performed using
// miln version 13. This code is meant to be used once, for generating the data
// and is very specific to this version, on purpose. In the future it can be
// copy-pasted and modified as needed.
func TestExtract(t *testing.T) {
	// dirs
	playDir := "../data/levels"
	playthroughsDir := "d:\\Miln\\playthroughs\\gabi-experiment1"

	inputPlayDataDir := "d:\\Miln\\stored\\experiment2\\ai-input\\play-data"
	inputTrainingDataDir := "d:\\Miln\\stored\\experiment2\\ai-input\\training-data"
	inputTestDataDir := "d:\\Miln\\stored\\experiment2\\ai-input\\test-data"

	outputPlayDataDir := "d:\\Miln\\stored\\experiment2\\ai-output\\play-data"
	outputTrainingDataDir := "d:\\Miln\\stored\\experiment2\\ai-output\\training-data"
	outputTestDataDir := "d:\\Miln\\stored\\experiment2\\ai-output\\test-data"

	DeleteDir(outputPlayDataDir)
	MakeDir(outputPlayDataDir)

	DeleteDir(outputTrainingDataDir)
	MakeDir(outputTrainingDataDir)

	DeleteDir(outputTestDataDir)
	MakeDir(outputTestDataDir)

	// Initialize list of played levels.
	playFiles := GetFiles(os.DirFS(playDir).(FS), ".", "*")
	playthroughFiles := GetFiles(os.DirFS(playthroughsDir).(FS), ".", "*")
	playedLevels := make([]PlayedLevel, len(playFiles))
	for idx := range playFiles {
		playFilename := filepath.Base(playFiles[idx])
		idxStr := playFilename[:2]
		playIdx, err := strconv.Atoi(idxStr)
		Check(err)
		assert.Equal(t, playIdx, idx+1)
		playedLevels[idx].PlayIdx = playIdx
		playedLevels[idx].Name = strings.Split(playFilename, ".")[0][3:]
		playthroughFilename := filepath.Base(playthroughFiles[idx])
		playedLevels[idx].PlaythroughName = strings.Split(playthroughFilename, ".")[0]
	}

	// Copy files to output dirs.
	var source, dest string
	for idx := range playedLevels {
		// Fill play data dir.
		source = inputPlayDataDir + "/" + playedLevels[idx].PlayLevelFilename()
		dest = outputPlayDataDir + "/" + playedLevels[idx].PlayLevelFilename()
		CopyFile(source, dest)

		if playedLevels[idx].IsTrainingLevel() {
			source = inputTrainingDataDir + "/" + playedLevels[idx].ParamsFilename()
		} else {
			source = inputTestDataDir + "/" + playedLevels[idx].ParamsFilename()
		}
		dest = outputPlayDataDir + "/" + playedLevels[idx].PlayParamsFilename()
		CopyFile(source, dest)

		source = playthroughsDir + "/" + playedLevels[idx].SourcePlaythroughFilename()
		dest = outputPlayDataDir + "/" + playedLevels[idx].PlayPlaythroughFilename()
		CopyFile(source, dest)

		if playedLevels[idx].IsTrainingLevel() {
			// Fill training data dir.
			source = inputTrainingDataDir + "/" + playedLevels[idx].LevelFilename()
			dest = outputTrainingDataDir + "/" + playedLevels[idx].LevelFilename()
			CopyFile(source, dest)

			source = inputTrainingDataDir + "/" + playedLevels[idx].ParamsFilename()
			dest = outputTrainingDataDir + "/" + playedLevels[idx].ParamsFilename()
			CopyFile(source, dest)

			source = playthroughsDir + "/" + playedLevels[idx].SourcePlaythroughFilename()
			dest = outputTrainingDataDir + "/" + playedLevels[idx].PlaythroughFilename()
			CopyFile(source, dest)
		} else {
			// Fill test data dir.
			source = inputTestDataDir + "/" + playedLevels[idx].LevelFilename()
			dest = outputTestDataDir + "/" + playedLevels[idx].LevelFilename()
			CopyFile(source, dest)

			source = inputTestDataDir + "/" + playedLevels[idx].ParamsFilename()
			dest = outputTestDataDir + "/" + playedLevels[idx].ParamsFilename()
			CopyFile(source, dest)

			source = playthroughsDir + "/" + playedLevels[idx].SourcePlaythroughFilename()
			dest = outputTestDataDir + "/" + playedLevels[idx].PlaythroughFilename()
			CopyFile(source, dest)
		}
	}

	// Extract data from playthroughs.
	dirs := []string{outputPlayDataDir, outputTrainingDataDir, outputTestDataDir}
	for _, dir := range dirs {
		playthroughFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
		paramsFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013-params")

		outputFile, err := os.Create(dir + "/output.csv")
		Check(err)
		for i := range playthroughFiles {
			var params Param
			LoadYAML(os.DirFS(dir).(FS), paramsFiles[i][2:], &params)

			playthrough := DeserializePlaythrough(ReadFile(dir + "/" + playthroughFiles[i]))
			world := NewWorld(playthrough.Seed, playthrough.Level)
			for _, input := range playthrough.History {
				world.Step(input)
			}

			// Playthrough	NEnemies	EnemySpeed	NObstacles	NFlames	PlayerHealth
			_, err := outputFile.WriteString(fmt.Sprintf("%s,%d,%d,%d,%d,%d\n",
				filepath.Base(playthroughFiles[i]),
				params.NEnemies.ToInt(),
				params.EnemySpeed.ToInt(),
				params.NObstacles.ToInt(),
				params.NFlames.ToInt(),
				world.Player.Health.ToInt()))
			Check(err)
		}
		Check(outputFile.Close())
	}
}

// Gets a median of 21 frames between actions.
func TestGetReactionSpeed(t *testing.T) {
	dir := "d:\\Miln\\stored\\experiment2\\ai-output\\training-data"
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln013")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	diffs := []int{}
	for _, inputFile := range inputFiles {
		playthrough := DeserializePlaythrough(ReadFile(inputFile))
		framesWithActions := GetFramesWithActions(playthrough)
		for i := 1; i < len(framesWithActions); i++ {
			diff := framesWithActions[i] - framesWithActions[i-1]
			diffs = append(diffs, diff)
		}
	}

	histogram := map[int]int{}
	for _, diff := range diffs {
		histogram[diff]++
	}
	fmt.Println(histogram)

	// Collect all the keys.
	keys := make([]int, 0)
	for k := range histogram {
		keys = append(keys, k)
	}

	// Sort the keys.
	slices.Sort(keys)

	// Print out the map, sorted by keys.
	outputFile, err := os.Create("outputs/reaction-speed.csv")
	Check(err)
	_, err = outputFile.WriteString(fmt.Sprintf("n_frames_between_actions,n_occurrences\n"))
	Check(err)
	for _, k := range keys {
		_, err = outputFile.WriteString(fmt.Sprintf("%d,%d\n", k, histogram[k]))
		Check(err)
	}
	Check(outputFile.Close())

	slices.Sort(diffs)
	median := diffs[len(diffs)/2]
	fmt.Println(median)
}
