package vision

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelevantPts3(t *testing.T) {
	obstacles := NewMatBool(Pt{I(50), I(50)})
	obstacles.Set(Pt{I(1), I(1)})

	// PrintMatBool(obstacles)
	// println()
	// println()

	a := NewAttackableTiles3(obstacles.Size())

	dummy := ZERO
	sz := obstacles.Size()
	for y := ZERO; y.Lt(sz.Y); y.Inc() {
		for x := ZERO; x.Lt(sz.X); x.Inc() {
			attackable := a.Compute(Pt{x, y}, obstacles)
			if attackable.At(Pt{I(2), I(2)}) {
				dummy.Inc()
			}
		}
	}
	// attackable := a.Compute(Pt{ZERO, ZERO}, obstacles)
	// if attackable.At(Pt{I(2), I(2)}) {
	// 	dummy.Inc()
	// }

	println(dummy.ToInt())

	assert.True(t, true)
}
