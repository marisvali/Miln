package world

import (
	"bytes"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestWorld_Regression1(t *testing.T) {
	playthrough := DeserializePlaythroughFromOld(ReadFile("playthroughs/large-playthrough.mln016"))
	expected := string(ReadFile("playthroughs/large-playthrough.mln016-hash"))
	actual := RegressionId(&playthrough)
	println(actual)
	assert.Equal(t, expected, actual)
}

func RunPlaythrough(p Playthrough) {
	w := NewWorld(p.Seed, p.Level)
	for i := range p.History {
		w.Step(p.History[i])
	}
}

func BenchmarkPlaythroughSpeed(b *testing.B) {
	playthrough := DeserializePlaythroughFromOld(ReadFile("playthroughs/average-playthrough.mln016"))
	for b.Loop() {
		RunPlaythrough(playthrough)
	}
}

// The benchmarks below are relevant mostly in relation to each other, to
// answer the question: how long does world serialization take and can it be
// performed every frame without impacting the FPS?

// Typical output:
// BenchmarkSerializedPlaythrough_WithoutCompression-12    	 124       9.553545 ms/op
// BenchmarkSerializedPlaythrough_Compression-12    	      48	  22.284935 ms/op
// BenchmarkWorldClone-12    	    						6948	   0.170807 ms/op
// Conclusion: serializing is too expensive to perform every frame, better to
// just clone the world and send it to a go routine once every K frames, where
// it can then be serialized and saved to a file or uploaded to a database.

// Get large world (18520 frames, 308 sec).
// Note: I don't really care about the final state of the world, I just need a
// world that ran for a relatively long time.

func GetLargeWorld() World {
	// Load the playthrough.
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln016"))

	// Run the playthrough.
	w := NewWorld(playthrough.Seed, playthrough.Level)
	for i := range playthrough.History {
		w.Step(playthrough.History[i])
	}
	return w
}

// Check how much time it takes to serialize a world (without compressing it).
func BenchmarkSerializedPlaythrough_WithoutCompression(b *testing.B) {
	// Initialize, get large playthrough.
	p := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln016"))

	// Run benchmark loop.
	for b.Loop() {
		buf := new(bytes.Buffer)
		Serialize(buf, int64(Version))
		Serialize(buf, p)
	}
}

// Check how much time it takes to compress a serialized world.
func BenchmarkSerializedPlaythrough_Compression(b *testing.B) {
	// Initialize, get large playthrough.
	p := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln016"))

	// Serialize the world to buf.
	buf := new(bytes.Buffer)
	Serialize(buf, int64(Version))
	Serialize(buf, p)

	// Run benchmark loop.
	for b.Loop() {
		Zip(buf.Bytes())
	}
}

func BenchmarkWorldClone(b *testing.B) {
	// Initialize, get large playthrough.
	p := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln016"))

	// Run benchmark loop.
	res := 0
	for b.Loop() {
		p2 := p
		res += len(p2.History)
	}
}

func TestWorld_PredictableRandomness(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln016"))

	// Run the playthrough halfway through.
	w1 := NewWorld(playthrough.Seed, playthrough.Level)
	for i := 0; i < len(playthrough.History)/2; i++ {
		w1.Step(playthrough.History[i])
	}

	w2 := w1

	// Finish running the playthrough.
	// Intersperse global randoms to show that each world behaves predictably:
	// - its randomness is independent of global randomness
	// - clones have the same randomness
	noise := ZERO
	for i := len(playthrough.History) / 2; i < len(playthrough.History); i++ {
		w1.Step(playthrough.History[i])
		noise.Add(RInt(I(0), I(10)))
		w2.Step(playthrough.History[i])
	}

	println(noise.ToInt64())
	println(w1.State())
	println(w2.State())

	assert.Equal(t, w1.State(), w2.State())
}

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
