package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"testing"
)

type Param struct {
	NEnemies   Int `yaml:"NEnemies"`
	EnemySpeed Int `yaml:"EnemySpeed"`
	NObstacles Int `yaml:"NObstacles"`
	NFlames    Int `yaml:"NFlames"`
}

type AiInput struct {
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

func GenerateLevelFromParams(p Param) Level {
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
	l.Obstacles = ValidRandomLevel(p.NObstacles)
	occ := l.Obstacles
	for i := 0; i < p.NEnemies.ToInt(); i++ {
		var sp SpawnPortalParams
		sp.Pos = occ.OccupyRandomPos(&DefaultRand)
		sp.SpawnPortalCooldown = I(100)
		wave := Wave{}
		wave.SecondsAfterLastWave = I(0)
		wave.NHounds = I(1)
		sp.Waves.Data[0] = wave
		sp.WavesLen = 1
		l.SpawnPortalsParams.Data[i] = sp
	}
	l.SpawnPortalsParams.N = p.NEnemies.ToInt64()
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
	workDir := "d:\\Miln\\stored\\experiment4\\ai-input"
	fsys := os.DirFS(workDir).(FS)
	var input AiInput
	LoadYAML(fsys, "ai-input.yaml", &input)

	// Generate test data.
	testDir := workDir + "\\test-data"
	DeleteDir(testDir)
	MakeDir(testDir)
	ChDir(testDir)

	for instanceIdx := range input.NumTestInstances.ToInt() {
		levelS := fmt.Sprintf("test-%03d", instanceIdx+1)
		params := input.GenerateParam()
		SaveYAML(fmt.Sprintf("%s.mln016-params", levelS), params)
		l := GenerateLevelFromParams(params)
		l.SaveToYAML(RInt63(), fmt.Sprintf("%s.mln016-level", levelS))
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

func Test_GeneratePracticeLevels(t *testing.T) {
	workDir := "d:\\Miln\\stored\\experiment2\\ai-input"

	// Generate practice data.
	praticeDir := workDir + "\\practice-data"
	DeleteDir(praticeDir)
	MakeDir(praticeDir)
	ChDir(praticeDir)

	practiceParams := []Param{
		{NEnemies: I(7), EnemySpeed: I(92), NObstacles: I(15), NFlames: I(4)},
		{NEnemies: I(7), EnemySpeed: I(92), NObstacles: I(17), NFlames: I(4)},
		{NEnemies: I(7), EnemySpeed: I(92), NObstacles: I(20), NFlames: I(4)},
		{NEnemies: I(8), EnemySpeed: I(92), NObstacles: I(15), NFlames: I(4)},
		{NEnemies: I(8), EnemySpeed: I(92), NObstacles: I(17), NFlames: I(4)},
		{NEnemies: I(8), EnemySpeed: I(92), NObstacles: I(20), NFlames: I(4)},
		{NEnemies: I(9), EnemySpeed: I(92), NObstacles: I(15), NFlames: I(4)},
		{NEnemies: I(9), EnemySpeed: I(92), NObstacles: I(15), NFlames: I(3)},
		{NEnemies: I(9), EnemySpeed: I(92), NObstacles: I(17), NFlames: I(3)},
		{NEnemies: I(9), EnemySpeed: I(92), NObstacles: I(20), NFlames: I(3)},
	}

	for paramIdx, params := range practiceParams {
		levelS := fmt.Sprintf("practice-%02d", paramIdx+1)
		SaveYAML(fmt.Sprintf("%s.mln013-params", levelS), params)
		l := GenerateLevelFromParams(params)
		l.SaveToYAML(RInt63(), fmt.Sprintf("%s.mln013-level", levelS))
	}
}
