package gamelib

import (
	"github.com/stretchr/testify/assert"
	"math"
	"slices"
	"testing"
)

func TestPt_Add(t *testing.T) {
	p := IPt(1, 2)
	p.Add(IPt(3, 4))
	assert.Equal(t, p, IPt(4, 6))
	p = IPt(1, 2)
	p.Add(IPt(-3, -4))
	assert.Equal(t, p, IPt(-2, -2))
}

func TestPt_AddLenSimple(t *testing.T) {
	p := IPt(1, 1)
	p.AddLen(I(1))
	assert.Equal(t, IPt(2, 2), p)

	p = IPt(100, 0)
	p.AddLen(I(20))
	assert.Equal(t, IPt(120, 0), p)

	p = IPt(0, 13)
	p.AddLen(I(20))
	assert.Equal(t, IPt(0, 33), p)

	p = IPt(800, 130)
	oldLen := p.Len()
	extraLen := I(54)
	p.AddLen(extraLen)
	assert.Equal(t, oldLen.Plus(extraLen), p.Len())
}

func TestPt_AddLen_TooSmallToHaveEffect(t *testing.T) {
	p := IPt(800, 130)
	oldLen := p.Len()
	extraLen := I(1)
	p.AddLen(extraLen)
	// Because we're working with integers, we won't ever get the exact length
	// we want. Sometimes it's not possible to keep a vector's direction the
	// same, expressed in integers and extend its length by 1.
	// For vector (800, 130) the length is 810 (real length 810.493...).
	// In order to extend the length by 1 I would need to get length 811.
	// Which means I need a vector with the same direction but with length 811
	// or 811.999...
	// The vector with the same direction and length 811 has coordinates ~=
	// (800.499..., 130.081...).
	// The vector with the same direction and length 811.99 has coordinates ~=
	// (801.486..., 130.241...).
	// So there's no way to get what we want.
	assert.NotEqual(t, oldLen.Plus(extraLen), p.Len())
	assert.Equal(t, oldLen, p.Len())
}

// TODO: study what tolerance I can guarantee and how to test this better
func TestPt_AddLen_ErrorTolerance(t *testing.T) {
	// The reasoning for this test is very similar to the one for SetLen.
	// See that one first.

	// extraLen := I(int64(math.Sqrt(float64(math.MaxInt64)) / 100))
	extraLen := I(100)

	// Get maximum error for adding length to vectors with coordinates
	// between 10 and 100 (exhaustive search).
	{
		var diffs []float64
		for x := 10; x < 100; x++ {
			for y := 10; y < 100; y++ {
				diff := AddLenGetDif(Pt{I(x), I(y)}, extraLen)
				diffs = append(diffs, diff)
			}
		}
		// fmt.Println("maximum diff percentage-wise:", slices.Max(diffs))
		maxDiff := slices.Max(diffs)
		assert.Less(t, maxDiff, 8.0) // < 8%
	}

	// Get maximum error for adding length to vectors with coordinates
	// between 100 and 1000 (exhaustive search).
	{
		var diffs []float64
		for x := 100; x < 1000; x++ {
			for y := 100; y < 1000; y++ {
				diff := AddLenGetDif(Pt{I(x), I(y)}, extraLen)
				diffs = append(diffs, diff)
			}
		}
		// fmt.Println("maximum diff percentage-wise:", slices.Max(diffs))
		maxDiff := slices.Max(diffs)
		assert.Less(t, maxDiff, 3.0) // < 3%
	}

	// Get maximum error for adding length to vectors with coordinates
	// between 1000 and 10000 (random samples)
	{
		var diffs []float64
		RSeed(I(0))
		for i := 1; i < 1000000; i++ {
			randomPt := Pt{RInt(I(1000), I(9999)), RInt(I(1000), I(9999))}
			diff := AddLenGetDif(randomPt, extraLen)
			diffs = append(diffs, diff)
		}
		// fmt.Println("maximum diff percentage-wise:", slices.Max(diffs))
		maxDiff := slices.Max(diffs)
		assert.Less(t, maxDiff, 3.0) // < 3%
	}
}

func AddLenGetDif(a Pt, extraLen Int) float64 {
	oldLen := a.Len()
	a.AddLen(extraLen)
	newLen := a.Len()
	// difference between the desired extraLen and actual extraLen
	dif := extraLen.Minus(newLen.Minus(oldLen)).Abs()
	difPercent := dif.ToFloat64() / extraLen.ToFloat64() * 100
	// truncate it to 3 decimals so it's a little nicer when printed
	difPercent = float64(int64(difPercent*1000)) / 1000.0
	return difPercent
}

func TestPt_Len(t *testing.T) {
	assert.Equal(t, IPt(0, 0).Len(), I(0))                    // real length: 0.000..
	assert.Equal(t, IPt(1, 0).Len(), I(1))                    // real length: 1.000..
	assert.Equal(t, IPt(1, 1).Len(), I(1))                    // real length: 1.414..
	assert.Equal(t, IPt(5, 5).Len(), I(7))                    // real length: 7.071..
	assert.Equal(t, IPt(-13, -5).Len(), I(13))                // real length: 13.928..
	assert.Equal(t, IPt(-13, 5).Len(), I(13))                 // real length: 13.928..
	assert.Equal(t, IPt(13, -5).Len(), I(13))                 // real length: 13.928..
	assert.Equal(t, IPt(130034, 23458883).Len(), I(23459243)) // real length: 23,459,243.390..
	assert.Panics(t, func() { IPt(math.MaxInt64, math.MaxInt64).Len() })
	assert.Panics(t, func() { IPt(math.MinInt64, 34).Len() })
}

func TestPt_Scale(t *testing.T) {
	a := IPt(-13, 5)
	a.Scale(I(2), I(1))
	assert.Equal(t, a, IPt(-26, 10))

	a = IPt(-100, 25)
	a.Scale(I(1), I(2))
	assert.Equal(t, a, IPt(-50, 12))

	a = IPt(14037623, -3212809)
	a.Scale(I(9), I(4))
	assert.Equal(t, a, IPt(31584651, -7228820))

	a = IPt(1, 5)
	a.Scale(I(2), I(11))
	assert.Equal(t, a, IPt(0, 0))

	a = IPt(0, 0)
	a.Scale(I(10), I(1))
	assert.Equal(t, a, IPt(0, 0))

	a = IPt(14037623, -3212809)
	assert.Panics(t, func() { a.Scale(I(math.MaxInt64), I(math.MaxInt64)) })
}

func TestPt_SetLen(t *testing.T) {
	// When we set the length, we will get a vector with a length that's
	// a little different from the target. This is due to the fact that I'm
	// using integer operations instead of floats.
	// The question now is to see what "a little different" means.
	// I've deduced analytically that the difference is proportional to the
	// original size of the vector and there should be some thresholds.
	// I want to find these thresholds numerically.
	// I'll use these categories of vectors:
	// - coordinates in the tens (between 10 and 100)
	// - coordinates in the hundreds (between 100 and 1000)
	// - coordinates in the thousands (between 1000 and 10000)

	// Set the target length to something large.
	// We need to be able to calculate the final length without overflowing.
	// That requires computing the squared length, which means we must be
	// under sqrt(MaxInt64) otherwise we will get an overflow when trying
	// to compute the final actual length.
	targetLen := I(int(math.Sqrt(float64(math.MaxInt64)) / 10))

	// Get maximum error for setting the length of vectors with coordinates
	// between 10 and 100 (exhaustive search).
	{
		var diffs []float64
		for x := 10; x < 100; x++ {
			for y := 10; y < 100; y++ {
				diff := SetLenGetDif(Pt{I(x), I(y)}, targetLen)
				diffs = append(diffs, diff)
			}
		}
		// fmt.Println("maximum diff percentage-wise:", slices.Max(diffs))
		maxDiff := slices.Max(diffs)
		assert.Less(t, maxDiff, 7.0) // < 7%
	}

	// Get maximum error for setting the length of vectors with coordinates
	// between 100 and 1000 (exhaustive search).
	{
		var diffs []float64
		for x := 100; x < 1000; x++ {
			for y := 100; y < 1000; y++ {
				diff := SetLenGetDif(Pt{I(x), I(y)}, targetLen)
				diffs = append(diffs, diff)
			}
		}
		// fmt.Println("maximum diff percentage-wise:", slices.Max(diffs))
		maxDiff := slices.Max(diffs)
		assert.Less(t, maxDiff, 0.7) // < 0.7%
	}

	// Get maximum error for setting the length of vectors with coordinates
	// between 1000 and 10000 (random samples)
	{
		var diffs []float64
		RSeed(I(0))
		for i := 1; i < 1000000; i++ {
			randomPt := Pt{RInt(I(1000), I(9999)), RInt(I(1000), I(9999))}
			diff := SetLenGetDif(randomPt, targetLen)
			diffs = append(diffs, diff)
		}
		// fmt.Println("maximum diff percentage-wise:", slices.Max(diffs))
		maxDiff := slices.Max(diffs)
		assert.Less(t, maxDiff, 0.07) // < 0.07%
	}
}

func SetLenGetDif(a Pt, targetLen Int) float64 {
	a.SetLen(targetLen)
	dif := a.Len().Minus(targetLen).Abs()
	difPercent := dif.ToFloat64() / targetLen.ToFloat64() * 100
	// truncate it to 3 decimals so it's a little nicer when printed
	difPercent = float64(int64(difPercent*1000)) / 1000.0
	return difPercent
}

func TestPt_SquaredDistTo(t *testing.T) {
	assert.Equal(t, IPt(0, 0).SquaredDistTo(IPt(1, 1)), I(2))
	assert.Equal(t, IPt(1, 1).SquaredDistTo(IPt(0, 0)), I(2))
	assert.Equal(t, IPt(6, -8).SquaredDistTo(IPt(4, 77)), I(7229))
	assert.Equal(t, IPt(130034, 23458883).SquaredDistTo(IPt(0, 0)), I(550336100448845))
	assert.Panics(t, func() { IPt(3424543543, -943242123433).SquaredDistTo(IPt(-3424543543, 943242123433)) })
}

func TestPt_SquaredLen(t *testing.T) {
	assert.Equal(t, IPt(0, 0).SquaredLen(), I(0))
	assert.Equal(t, IPt(1, 0).SquaredLen(), I(1))
	assert.Equal(t, IPt(1, 1).SquaredLen(), I(2))
	assert.Equal(t, IPt(5, 5).SquaredLen(), I(50))
	assert.Equal(t, IPt(-13, -5).SquaredLen(), I(194))
	assert.Equal(t, IPt(-13, 5).SquaredLen(), I(194))
	assert.Equal(t, IPt(13, -5).SquaredLen(), I(194))
	assert.Equal(t, IPt(130034, 23458883).SquaredLen(), I(550336100448845))
	assert.Panics(t, func() { IPt(math.MaxInt64, math.MaxInt64).SquaredLen() })
	assert.Panics(t, func() { IPt(math.MinInt64, 34).SquaredLen() })
}

func TestPt_To(t *testing.T) {
	assert.Equal(t, IPt(0, 0).To(IPt(1, 1)), IPt(1, 1))
	assert.Equal(t, IPt(1, 1).To(IPt(0, 0)), IPt(-1, -1))
	assert.Equal(t, IPt(6, -8).To(IPt(4, 77)), IPt(-2, 85))
	assert.Equal(t, IPt(123, 0).To(IPt(122, 1)), IPt(-1, 1))
	assert.Equal(t, IPt(3424543543, -943242123433).To(IPt(-3424543543, 943242123433)), IPt(-6849087086, 1886484246866))
	assert.Panics(t, func() { IPt(math.MinInt64, math.MinInt64).To(IPt(math.MaxInt64, math.MaxInt64)) })
	assert.Panics(t, func() { IPt(math.MaxInt64, math.MaxInt64).To(IPt(math.MinInt64, math.MinInt64)) })
}
