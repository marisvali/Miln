package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"path/filepath"
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
// miln version 17-17. This code is meant to be used once, for generating the
// data and is very specific to this version, on purpose. In the future it can
// be copy-pasted and modified as needed.
func TestExtract(t *testing.T) {
	// dirs
	dir := "d:\\Miln\\stored\\experiment3\\ai-output\\test-data"

	// Extract data from playthroughs.
	playthroughFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln017-017")
	paramsFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln017-017-params")
	outputFile, err := os.Create(dir + "/output.csv")
	Check(err)
	for i := range playthroughFiles {
		var params Param
		LoadYAML(os.DirFS(dir).(FS), paramsFiles[i][2:], &params)

		playthrough := DeserializePlaythrough(ReadFile(dir + "/" + playthroughFiles[i]))
		world := NewWorldFromPlaythrough(playthrough)
		for j := range playthrough.History {
			world.Step(playthrough.History[j])
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
