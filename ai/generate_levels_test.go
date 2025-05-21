package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

type Param struct {
	NEnemies   Int `yaml:"NEnemies"`
	EnemySpeed Int `yaml:"EnemySpeed"`
	NObstacles Int `yaml:"NObstacles"`
	NFlames    Int `yaml:"NFlames"`
}

type TrainingData struct {
	Params                  []Param `yaml:"Params"`
	NumInstancesPerParamSet Int     `yaml:"NumInstancesPerParamSet"`
}

type TestData struct {
	NumTestInstances Int `yaml:"NumTestInstances"`
	NObstaclesMin    Int `yaml:"NObstaclesMin"`
	NObstaclesMax    Int `yaml:"NObstaclesMax"`
	NEnemiesMin      Int `yaml:"NEnemiesMin"`
	NEnemiesMax      Int `yaml:"NEnemiesMax"`
	EnemySpeedMin    Int `yaml:"EnemySpeedMin"`
	EnemySpeedMax    Int `yaml:"EnemySpeedMax"`
	NFlamesMin       Int `yaml:"NFlamesMin"`
	NFlamesMax       Int `yaml:"NFlamesMax"`
}

type AiInput struct {
	TrainingData `yaml:"TrainingData"`
	TestData     `yaml:"TestData"`
}

func SpeedToCooldown(speed Int) Int {
	// speed = 10 + (500-cooldown)/490*90
	// (speed - 10) / 90 * 490 = 500 - cooldown
	// cooldown = 500 - (speed - 10) / 90 * 490
	// cooldown = 500 - (speed - 10) * 490 / 90
	cooldown := I(300).Minus(speed.Minus(I(10)).Times(I(290)).DivBy(I(90)))
	return cooldown
}

func NFlamesToAmmoLimit(nFlames Int) Int {
	ammoLimit := nFlames.Minus(I(1)).Times(I(3)).Plus(I(1))
	return ammoLimit
}

func GenerateTrainingLevel(p Param) Level {
	var l Level
	l.Boardgame = false
	l.UseAmmo = true
	l.AmmoLimit = NFlamesToAmmoLimit(p.NFlames)
	l.EnemyMoveCooldownDuration = SpeedToCooldown(p.EnemySpeed)
	l.EnemiesAggroWhenVisible = true
	l.SpawnPortalCooldownMin = I(100)
	l.SpawnPortalCooldownMax = I(100)
	l.HoundMaxHealth = I(3)
	l.HoundMoveCooldownMultiplier = I(1)
	l.HoundPreparingToAttackCooldown = SpeedToCooldown(p.EnemySpeed)
	l.HoundAttackCooldownMultiplier = I(1)
	l.HoundHitCooldownDuration = SpeedToCooldown(p.EnemySpeed)
	l.HoundHitsPlayer = true
	l.HoundAggroDistance = ZERO
	l.Obstacles = ValidRandomLevel(p.NObstacles, I(8), I(8))
	occ := l.Obstacles.Clone()
	var sps []SpawnPortalParams
	for i := 0; i < p.NEnemies.ToInt(); i++ {
		var sp SpawnPortalParams
		sp.Pos = occ.OccupyRandomPos(&DefaultRand)
		sp.SpawnPortalCooldown = I(100)
		wave := Wave{}
		wave.SecondsAfterLastWave = I(0)
		wave.NHounds = I(1)
		sp.Waves = []Wave{wave}
		sps = append(sps, sp)
	}
	l.SpawnPortalsParams = sps
	return l
}

func (a *AiInput) GenerateParam() Param {
	var p Param
	p.NEnemies = RInt(a.NEnemiesMin, a.NEnemiesMax)
	p.NObstacles = RInt(a.NObstaclesMin, a.NObstaclesMax)
	p.EnemySpeed = RInt(a.EnemySpeedMin, a.EnemySpeedMax)
	p.NFlames = RInt(a.NFlamesMin, a.NFlamesMax)
	return p
}

func Test_GenerateInputData(t *testing.T) {
	workDir := "d:\\Miln\\stored\\experiment2\\ai-input"
	fsys := os.DirFS(workDir).(FS)
	var input AiInput
	LoadYAML(fsys, "ai-input.yaml", &input)

	// Generate training data.
	trainingDir := workDir + "\\training-data"
	DeleteDir(trainingDir)
	MakeDir(trainingDir)
	ChDir(trainingDir)

	for paramIdx, params := range input.Params {
		for instanceIdx := range input.NumInstancesPerParamSet.ToInt() {
			levelS := fmt.Sprintf("training-params-set-%02d-instance-%02d", paramIdx+1, instanceIdx+1)
			SaveYAML(fmt.Sprintf("%s.mln999-params", levelS), params)
			l := GenerateTrainingLevel(params)
			l.SaveToYAML(RInt63(), fmt.Sprintf("%s.mln999-level", levelS))
		}
	}

	// Generate test data.
	testDir := workDir + "\\test-data"
	DeleteDir(testDir)
	MakeDir(testDir)
	ChDir(testDir)

	for instanceIdx := range input.NumTestInstances.ToInt() {
		levelS := fmt.Sprintf("test-%02d", instanceIdx+1)
		params := input.GenerateParam()
		SaveYAML(fmt.Sprintf("%s.mln999-params", levelS), params)
		l := GenerateTrainingLevel(params)
		l.SaveToYAML(RInt63(), fmt.Sprintf("%s.mln999-level", levelS))
	}

	// Generate play data.
	trainingFiles := GetFiles(fsys, "training-data", "*-level")
	Shuffle(&DefaultRand, trainingFiles)
	testFiles := GetFiles(fsys, "test-data", "*-level")
	Shuffle(&DefaultRand, testFiles)

	playDir := workDir + "\\play-data"
	DeleteDir(playDir)
	MakeDir(playDir)
	ChDir(playDir)

	files := []string{}
	minLen := min(len(trainingFiles), len(testFiles))
	for i := range minLen {
		files = append(files, trainingFiles[i])
		files = append(files, testFiles[i])
	}
	for i := minLen; i < len(trainingFiles); i++ {
		files = append(files, trainingFiles[i])
	}
	for i := minLen; i < len(testFiles); i++ {
		files = append(files, testFiles[i])
	}

	for i, filename := range files {
		sourceFilename := workDir + "/" + filename
		destFilename := playDir + fmt.Sprintf("/%02d-", i+1) + filepath.Base(filename)
		CopyFile(sourceFilename, destFilename)
	}
}

func Test_Func(t *testing.T) {
	fmt.Println(I(10), SpeedToCooldown(I(10)))
	fmt.Println(I(30), SpeedToCooldown(I(30)))
	fmt.Println(I(50), SpeedToCooldown(I(50)))
	fmt.Println(I(70), SpeedToCooldown(I(70)))
	fmt.Println(I(80), SpeedToCooldown(I(80)))
	fmt.Println(I(90), SpeedToCooldown(I(90)))
	fmt.Println(I(95), SpeedToCooldown(I(95)))
	fmt.Println(I(99), SpeedToCooldown(I(99)))
	fmt.Println(I(100), SpeedToCooldown(I(100)))
}
