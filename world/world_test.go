package world

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_WorldRegression1(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20240709-150738.mln002"))
	expected := string(ReadFile("playthroughs/20240709-150738.mln002-hash"))

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
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20240709-150738.mln002"))
	for n := 0; n < b.N; n++ {
		RunPlaythrough(playthrough)
	}
}
