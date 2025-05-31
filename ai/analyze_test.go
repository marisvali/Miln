package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"os"
	"strings"
	"testing"
)

func CooldownToSpeed(cooldown Int) Int {
	x := I(300).Minus(cooldown)
	speed := x.Times(I(90)).DivBy(I(290)).Plus(I(10))
	return speed
}

func AmmoLimitToNFlames(ammoLimit Int) Int {
	nFlames := ammoLimit.Minus(I(1)).DivBy(I(3)).Plus(I(1))
	return nFlames
}

func TestExtract(t *testing.T) {
	// playthroughs
	dir := os.DirFS("d:\\Miln\\playthroughs\\gabi-experiment1").(FS)
	files := GetFiles(dir, ".", "*")

	// list of levels with which the version was delivered, in order, with their
	// names used when generated
	files2 := GetFiles(os.DirFS("..").(FS), "data/levels", "*")

	// list of trainings
	dirTrainingData := os.DirFS("d:\\Miln\\stored\\experiment2\\ai-input\\training-data").(FS)
	filesTrainingParams := GetFiles(dirTrainingData, ".", "*.mln013-params")

	// list of test
	// dirTestData := os.DirFS("d:\\Miln\\stored\\experiment2\\ai-input\\test-data").(FS)
	// filesTestParams := GetFiles(dirTestData, ".", "*.mln013-params")

	for idx, file := range files {
		if !strings.Contains(files2[idx], "training") {
			continue
		}

		file = file[2:]
		data, err := dir.ReadFile(file)
		Check(err)
		playthrough := DeserializePlaythrough(data)

		var fileParams string
		for i := range filesTrainingParams {
			fileId := strings.Split(filesTrainingParams[i][2:], ".")[0]
			if strings.Contains(files2[idx], fileId) {
				fileParams = filesTrainingParams[i]
			}
		}
		var params Param
		LoadYAML(dirTrainingData, fileParams[2:], &params)

		world := NewWorld(playthrough.Seed, playthrough.Level)
		for _, input := range playthrough.History {
			world.Step(input)
		}

		var result string
		if world.Player.Health.IsPositive() {
			result = "win"
		} else {
			result = "loss"
		}

		// Playthrough	NEnemies	EnemySpeed	NObstacles	NFlames	Result
		fmt.Printf("%s,%d,%d,%d,%d,%s,%d\n",
			file,
			params.NEnemies.ToInt(),
			params.EnemySpeed.ToInt(),
			params.NObstacles.ToInt(),
			params.NFlames.ToInt(),
			result,
			world.Player.Health.ToInt())
	}
	assert.True(t, true)
}
