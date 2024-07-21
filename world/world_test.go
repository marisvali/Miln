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
