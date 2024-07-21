package world

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_WorldRegression1(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20240714-120933.mln006"))
	expected := string(ReadFile("playthroughs/20240714-120933.mln006-hash"))

	// Run the playthrough.
	w := NewWorld(playthrough.Seed)
	for _, input := range playthrough.History {
		w.Step(input)
	}

	println(w.RegressionId())

	assert.Equal(t, expected, w.RegressionId())
}

func RunPlaythrough(p Playthrough) {
	w := NewWorld(p.Seed)
	for _, input := range p.History {
		w.Step(input)
	}
}

func BenchmarkPlaythroughSpeed(b *testing.B) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20240714-120933.mln006"))
	for n := 0; n < b.N; n++ {
		for i := 1; i < 100; i++ {
			RunPlaythrough(playthrough)
		}
	}
}

func Test_SaveCache(t *testing.T) {
	v := NewVision(Pt{I(10), I(10)})
	v.BuildCache()
	file1 := "cache.cache"
	file2 := "cache.cache2"
	v.SaveCache(file1)

	v2 := NewVision(Pt{I(10), I(10)})
	v2.LoadCache(file1)
	v2.SaveCache(file2)

	assert.Equal(t, ReadFile(file1), ReadFile(file2))
}
