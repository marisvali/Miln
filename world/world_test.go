package world

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWorld_Regression1(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln17-17"))
	expected := string(ReadFile("playthroughs/large-playthrough.mln17-17-hash"))
	actual := RegressionId(&playthrough)
	println(actual)
	assert.Equal(t, expected, actual)
}

func BenchmarkWorldSpeed(b *testing.B) {
	p := DeserializePlaythrough(ReadFile("playthroughs/average-playthrough.mln17-17"))
	for b.Loop() {
		w := NewWorldFromPlaythrough(p)
		for i := range p.History {
			w.Step(p.History[i])
		}
	}
}

func TestWorld_PredictableRandomness(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln17-17"))

	// Run the playthrough halfway through.
	w1 := NewWorldFromPlaythrough(playthrough)
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
