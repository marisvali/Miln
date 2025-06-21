package world

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_ArrayYaml(t *testing.T) {
	var a SpawnPortalParamsArray
	a.N = 2
	a.V[0].Pos = IPt(20, 30)
	a.V[0].SpawnPortalCooldown = I(10)
	a.V[0].Waves.N = 2
	a.V[0].Waves.V[0].NHounds = I(2)
	a.V[0].Waves.V[1].NHounds = I(3)

	a.V[1].Pos = IPt(34, 54)
	a.V[1].SpawnPortalCooldown = I(10)
	a.V[1].Waves.N = 3
	a.V[1].Waves.V[0].SecondsAfterLastWave = I(4)
	a.V[1].Waves.V[0].NHounds = I(5)
	a.V[1].Waves.V[1].NHounds = I(6)
	a.V[1].Waves.V[2].NHounds = I(7)
	a.V[1].Waves.V[2].SecondsAfterLastWave = I(404)
	SaveYAML("try.yaml", a)

	var a2 SpawnPortalParamsArray
	LoadYAML(os.DirFS(".").(FS), "try.yaml", &a2)
	assert.Equal(t, a, a2)
}
