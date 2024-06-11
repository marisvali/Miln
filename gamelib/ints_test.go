package gamelib

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestInt_Add(t *testing.T) {
	a := I(3)
	a.Add(I(4))
	assert.Equal(t, a, I(7))
	a = I(math.MaxInt64 - 4)
	a.Add(I(4))
	assert.Equal(t, a, I(math.MaxInt64))
	a = I(math.MinInt64)
	a.Add(I(1))
	assert.Equal(t, a, I(math.MinInt64+1))
	a = I(math.MaxInt64)
	assert.Panics(t, func() { a.Add(I(1)) })
	a = I(math.MaxInt64)
	assert.Panics(t, func() { a.Add(I(4)) })
	a = I(math.MaxInt64 - 100)
	assert.Panics(t, func() { a.Add(I(101)) })
	a = I(134)
	assert.Panics(t, func() { a.Add(I(math.MaxInt64)) })
}

func TestInt_Plus(t *testing.T) {
	assert.Equal(t, I(5).Plus(I(1)), I(6))
	assert.Equal(t, I(-4).Plus(I(1)), I(-3))
	assert.Equal(t, I(math.MaxInt64-1).Plus(I(1)), I(math.MaxInt64))
	assert.Panics(t, func() { I(math.MaxInt64).Plus(I(1)) })
	assert.Equal(t, I(0).Plus(I(math.MinInt64)), I(math.MinInt64))
	assert.Panics(t, func() { I(-1).Plus(I(math.MinInt64)) })
	assert.NotPanics(t, func() { I(math.MaxInt64 - 1000).Plus(I(1000)) })
	assert.Panics(t, func() { I(math.MaxInt64 - 1000).Plus(I(1001)) })
}

func TestInt_Minus(t *testing.T) {
	assert.Equal(t, I(5).Minus(I(1)), I(4))
	assert.Equal(t, I(-4).Minus(I(1)), I(-5))
	assert.Equal(t, I(math.MinInt64+1).Minus(I(1)), I(math.MinInt64))
	assert.Panics(t, func() { I(math.MinInt64).Minus(I(1)) })
	assert.Equal(t, I(0).Minus(I(math.MaxInt64)), I(math.MinInt64+1))
	assert.Equal(t, I(-1).Minus(I(math.MaxInt64)), I(math.MinInt64))
	assert.Panics(t, func() { I(-2).Minus(I(math.MaxInt64)) })
	assert.Panics(t, func() { I(0).Minus(I(math.MinInt64)) })
	assert.NotPanics(t, func() { I(math.MinInt64 + 1000).Minus(I(1000)) })
	assert.Panics(t, func() { I(math.MinInt64 + 1000).Minus(I(1001)) })
}

func TestI(t *testing.T) {
	assert.Equal(t, I(123), Int{123})
	assert.NotEqual(t, I(124), Int{123})
}

func TestInt_Dec(t *testing.T) {
	a := I(2)
	a.Dec()
	assert.Equal(t, a, I(1))
	a = I(math.MaxInt64)
	a.Dec()
	assert.Equal(t, a, I(math.MaxInt64-1))
	a = I(math.MinInt64 + 1)
	a.Dec()
	assert.Equal(t, a, I(math.MinInt64))
	a = I(math.MinInt64)
	assert.Panics(t, func() { a.Dec() })
}

func TestInt_DivBy(t *testing.T) {
	assert.Equal(t, I(4).DivBy(I(2)), I(2))
	assert.Equal(t, I(3).DivBy(I(2)), I(1))
	assert.Equal(t, I(2).DivBy(I(2)), I(1))
	assert.Equal(t, I(1).DivBy(I(2)), I(0))
	assert.Equal(t, I(0).DivBy(I(2)), I(0))
	assert.Equal(t, I(math.MaxInt64).DivBy(I(2)), I(math.MaxInt64/2))
	assert.Panics(t, func() { I(123).DivBy(I(0)) })
	assert.Panics(t, func() { I(math.MinInt64).DivBy(I(-1)) })
}

func TestInt_Mod(t *testing.T) {
	assert.Equal(t, I(4).Mod(I(2)), I(0))
	assert.Equal(t, I(3).Mod(I(2)), I(1))
	assert.Equal(t, I(2).Mod(I(2)), I(0))
	assert.Equal(t, I(1).Mod(I(2)), I(1))
	assert.Equal(t, I(0).Mod(I(2)), I(0))
	assert.Equal(t, I(17).Mod(I(5)), I(2))
	assert.Equal(t, I(17).Mod(I(4)), I(1))
	assert.Equal(t, I(29).Mod(I(5)), I(4))
	assert.Equal(t, I(math.MaxInt64).Mod(I(2)), I(1))
	assert.Equal(t, I(math.MaxInt64-1).Mod(I(math.MaxInt64)), I(math.MaxInt64-1))
	assert.Panics(t, func() { I(123).Mod(I(0)) })
}

func TestInt_Abs(t *testing.T) {
	assert.Equal(t, I(17).Abs(), I(17))
	assert.Equal(t, I(-17).Abs(), I(17))
	assert.Equal(t, I(0).Abs(), I(0))
}

func TestInt_Eq(t *testing.T) {
	assert.True(t, I(123).Eq(I(123)))
	assert.False(t, I(123).Eq(I(124)))
}

func TestInt_Geq(t *testing.T) {
	assert.True(t, I(124).Geq(I(123)))
	assert.True(t, I(123).Geq(I(123)))
	assert.False(t, I(122).Geq(I(123)))
}

func TestInt_Gt(t *testing.T) {
	assert.True(t, I(124).Gt(I(123)))
	assert.False(t, I(123).Gt(I(123)))
	assert.False(t, I(122).Geq(I(123)))
}

func TestInt_Inc(t *testing.T) {
	a := I(-2)
	a.Inc()
	assert.Equal(t, a, I(-1))
	a = I(math.MinInt64)
	a.Inc()
	assert.Equal(t, a, I(math.MinInt64+1))
	a = I(math.MaxInt64 - 1)
	a.Inc()
	assert.Equal(t, a, I(math.MaxInt64))
	a = I(math.MaxInt64)
	assert.Panics(t, func() { a.Inc() })
}

func TestInt_Leq(t *testing.T) {
	assert.False(t, I(124).Leq(I(123)))
	assert.True(t, I(123).Leq(I(123)))
	assert.True(t, I(122).Leq(I(123)))
}

func TestInt_Lt(t *testing.T) {
	assert.False(t, I(124).Lt(I(123)))
	assert.False(t, I(123).Lt(I(123)))
	assert.True(t, I(122).Lt(I(123)))
}

func TestInt_Neq(t *testing.T) {
	assert.False(t, I(123).Neq(I(123)))
	assert.True(t, I(123).Neq(I(124)))
}

func TestInt_Sqr(t *testing.T) {
	assert.Equal(t, I(9).Sqr(), I(81))
	assert.Equal(t, I(-2).Sqr(), I(4))
	assert.Panics(t, func() { I(math.MaxInt64).Sqr() })
	assert.Panics(t, func() { I(math.MaxInt64 / 2).Sqr() })
	assert.Panics(t, func() { I(math.MinInt64).Sqr() })
}

func TestInt_Sqrt(t *testing.T) {
	assert.Equal(t, I(9).Sqrt(), I(3))
	assert.Equal(t, I(8).Sqrt(), I(2))
	assert.Equal(t, I(10).Sqrt(), I(3))
	assert.Equal(t, I(0).Sqrt(), I(0))
	assert.Panics(t, func() { I(-1).Sqrt() })
	assert.Panics(t, func() { I(-100).Sqrt() })

	// Test what happens near the top of the max value.
	// A square root algorithm might do intermediary computations which are over
	// the number for which it has to find the square root.
	// This is not an exhaustive check but it's good enough to provide some
	// confidence.
	for i := int64(0); i < 20; i++ {
		nr := math.MaxInt64 - i
		res1 := I64(nr).Sqrt()
		res2 := I64(int64(math.Sqrt(float64(nr))))
		// we have to allow for a difference of 1 because the float sqrt has
		// some rounding errors and sometimes reports the integer above instead
		// of the one below
		assert.True(t, res1.Minus(res2).Abs().Lt(I(1)))
	}
	for i := int64(0); i < 100; i++ {
		nr := math.MaxInt64 - i*1000000
		res1 := I64(nr).Sqrt()
		res2 := I64(int64(math.Sqrt(float64(nr))))
		assert.True(t, res1.Minus(res2).Abs().Lt(I(1)))
	}
}

func TestInt_Subtract(t *testing.T) {
	a := I(3)
	a.Subtract(I(4))
	assert.Equal(t, a, I(-1))
	a = I(math.MaxInt64)
	a.Subtract(I(4))
	assert.Equal(t, a, I(math.MaxInt64-4))
	a = I(math.MinInt64 + 1)
	a.Subtract(I(1))
	assert.Equal(t, a, I(math.MinInt64))
	a = I(math.MinInt64)
	assert.Panics(t, func() { a.Subtract(I(1)) })
	a = I(math.MinInt64)
	assert.Panics(t, func() { a.Subtract(I(4)) })
	a = I(math.MinInt64 + 100)
	assert.Panics(t, func() { a.Subtract(I(101)) })
	a = I(-134)
	assert.Panics(t, func() { a.Subtract(I(math.MaxInt64)) })
}

func TestInt_Times(t *testing.T) {
	assert.Equal(t, I(0).Times(I(0)), I(0))
	assert.Equal(t, I(0).Times(I(123)), I(0))
	assert.Equal(t, I(123).Times(I(0)), I(0))
	assert.Equal(t, I(3).Times(I(4)), I(12))
	assert.Equal(t, I(math.MaxInt64).Times(I(-1)), I(-math.MaxInt64))
	assert.Equal(t, I(math.MinInt64+1).Times(I(-1)), I(math.MaxInt64))
	assert.Panics(t, func() { I(math.MinInt64).Times(I(-1)) })
	assert.Panics(t, func() { I(1234567890123).Times(I(1234567890123)) })
	assert.Panics(t, func() { I(math.MinInt64).Times(I(math.MinInt64)) })
	assert.Panics(t, func() { I(math.MaxInt64).Times(I(math.MaxInt64)) })
	assert.Panics(t, func() { I(math.MaxInt64 / 2).Times(I(math.MaxInt64 / 2)) })
	edgeVal := int64(math.Sqrt(math.MaxInt64))
	assert.Equal(t, I64(edgeVal).Times(I64(edgeVal)), I64(edgeVal*edgeVal))
	assert.Panics(t, func() { I64(edgeVal + 100).Times(I64(edgeVal + 102)) })
	assert.Panics(t, func() { I64(edgeVal - 100).Times(I64(edgeVal + 10023)) })
	assert.Equal(t, I64(-edgeVal).Times(I64(edgeVal)), I64(-edgeVal*edgeVal))
	assert.Equal(t, I64(-edgeVal).Times(I64(-edgeVal)), I64(edgeVal*edgeVal))
	assert.Panics(t, func() { I64(-edgeVal).Times(I64(edgeVal + 1340)) })
	assert.Panics(t, func() { I64(-edgeVal).Times(I64(-edgeVal - 1340)) })
}

func TestInt_ToFloat64(t *testing.T) {
	assert.Equal(t, I(0).ToFloat64(), float64(0))
	assert.Equal(t, I(123).ToFloat64(), float64(123))
	assert.Equal(t, I(-123).ToFloat64(), -123.0)
	assert.Equal(t, I(math.MaxInt64).ToFloat64(), float64(math.MaxInt64))
	assert.Equal(t, I(math.MinInt64).ToFloat64(), float64(math.MinInt64))
}

func TestInt_ToInt64(t *testing.T) {
	assert.Equal(t, I(0).ToInt64(), int64(0))
	assert.Equal(t, I(123).ToInt64(), int64(123))
	assert.Equal(t, I(-123).ToInt64(), int64(-123))
	assert.Equal(t, I(math.MaxInt64).ToInt64(), int64(math.MaxInt64))
	assert.Equal(t, I(math.MinInt64).ToInt64(), int64(math.MinInt64))
}
