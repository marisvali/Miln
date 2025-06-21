package world

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_LevelYaml(t *testing.T) {
	fsys := os.DirFS(".").(FS)

	var l Level
	l.Boardgame = false
	l.UseAmmo = true
	l.AmmoLimit = I(10)
	l.EnemyMoveCooldownDuration = I(107)
	l.EnemiesAggroWhenVisible = true
	l.SpawnPortalCooldownMin = I(100)
	l.SpawnPortalCooldownMax = I(100)
	l.HoundMaxHealth = I(3)
	l.HoundMoveCooldownMultiplier = I(1)
	l.HoundPreparingToAttackCooldown = I(107)
	l.HoundAttackCooldownMultiplier = I(1)
	l.HoundHitCooldownDuration = I(107)
	l.HoundHitsPlayer = true
	l.HoundAggroDistance = ZERO
	l.Obstacles = ValidRandomLevel(I(15))
	occ := l.Obstacles
	var sps SpawnPortalParamsArray
	for i := 0; i < 3; i++ {
		var sp SpawnPortalParams
		sp.Pos = occ.OccupyRandomPos(&DefaultRand)
		sp.SpawnPortalCooldown = I(100)
		wave := Wave{}
		wave.SecondsAfterLastWave = I(0)
		wave.NHounds = I(1)
		sp.Waves.Data[0] = wave
		sp.WavesLen++
		sps.Data[i] = sp
	}
	l.SpawnPortalsParams = sps
	l.SpawnPortalsParams.N = 3

	filename := "level.txt"
	SaveYAML(filename, l)
	var l2 Level
	LoadYAML(fsys, filename, &l2)
	DeleteFile(filename)
	assert.Equal(t, l, l2)
	l2.SpawnPortalsParams.Data[1].Waves.Data[0].SecondsAfterLastWave = I(99)
	assert.NotEqual(t, l, l2)

	l2.SaveToYAML(I(10), filename)
	seed, l3 := LoadLevelFromYAML(fsys, filename)
	DeleteFile(filename)
	assert.Equal(t, I(10), seed)
	assert.Equal(t, l2, l3)
}
