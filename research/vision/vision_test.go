package vision

import (
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelevantPts(t *testing.T) {
	obstacles := NewMatBool(Pt{I(50), I(50)})
	obstacles.Set(Pt{I(1), I(1)})

	// PrintMatBool(obstacles)
	// println()
	// println()

	a := NewVisibleTiles()

	dummy := ZERO
	sz := obstacles.Size()
	for y := ZERO; y.Lt(sz.Y); y.Inc() {
		for x := ZERO; x.Lt(sz.X); x.Inc() {
			visible := a.Compute(Pt{x, y}, obstacles)
			if visible.At(Pt{I(2), I(2)}) {
				dummy.Inc()
			}
		}
	}
	// visible := a.Compute(Pt{ZERO, ZERO}, obstacles)
	// if visible.At(Pt{I(2), I(2)}) {
	// 	dummy.Inc()
	// }

	println(dummy.ToInt())

	assert.True(t, true)
}
