//go:build !int_overflow_checks_enabled

package gamelib

/*
This is the variant of Int that has no checks. Useful for running expensive
computations (e.g. AI algorithms) when the correctness of the simulation code
is well established and the simulation execution time is a bottleneck.
*/

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Int struct {
	// This is made public for the sake of serializing and deserializing
	// using the encoding/binary package.
	// Don't access it otherwise.
	Val int64
}

var ZERO = I(0)
var ONE = I(1)
var TWO = I(2)

func I(val int) Int {
	return Int{int64(val)}
}

func I64(val int64) Int {
	return Int{val}
}

func (a Int) ToInt64() int64 {
	return a.Val
}

func (a Int) ToInt() int {
	return int(a.Val)
}

func (a Int) ToFloat64() float64 {
	return float64(a.Val)
}

func (a Int) Lt(b Int) bool {
	return a.Val < b.Val
}

func (a Int) Leq(b Int) bool {
	return a.Val <= b.Val
}

func (a Int) Eq(b Int) bool {
	return a.Val == b.Val
}

func (a Int) Neq(b Int) bool {
	return a.Val != b.Val
}

func (a Int) Gt(b Int) bool {
	return a.Val > b.Val
}

func (a Int) Geq(b Int) bool {
	return a.Val >= b.Val
}

func (a Int) IsPositive() bool {
	return a.Val > 0
}

func (a Int) IsZero() bool {
	return a.Val == 0
}

func (a Int) IsNegative() bool {
	return a.Val < 0
}

func (a Int) IsNonNegative() bool {
	return a.Val >= 0
}

func (a Int) IsNonPositive() bool {
	return a.Val <= 0
}

func (a Int) Between(val1, val2 Int) bool {
	valMin, valMax := MinMax(val1, val2)
	return a.Geq(valMin) && a.Leq(valMax)
}

func (a *Int) Inc() {
	a.Val++
}

func (a *Int) Dec() {
	a.Val--
}

// EnlargeByOne does Inc() when a > 0, Dec() when a < 0
func (a *Int) EnlargeByOne() {
	if a.Val >= 0 {
		a.Inc()
	} else {
		a.Dec()
	}
}

func (a Int) EnlargedByOne() Int {
	if a.Val >= 0 {
		a.Inc()
	} else {
		a.Dec()
	}
	return a
}

func (a Int) Abs() Int {
	if a.Val < 0 {
		return Int{-a.Val}
	} else {
		return a
	}
}

func (a Int) Negative() Int {
	return Int{-a.Val}
}

func (a Int) Plus(b Int) Int {
	return Int{a.Val + b.Val}
}

func (a *Int) Add(b Int) {
	a.Val += b.Val
}

func (a Int) Minus(b Int) Int {
	return Int{a.Val - b.Val}
}

func (a *Int) Subtract(b Int) {
	a.Val -= b.Val
}

func (a Int) Times(b Int) Int {
	return Int{a.Val * b.Val}
}

func (a Int) DivBy(b Int) Int {
	return Int{a.Val / b.Val}
}

func (a Int) Mod(b Int) Int {
	return Int{a.Val % b.Val}
}

func (a Int) Sqr() Int {
	return a.Times(a)
}

/**
 * Fast Square root algorithm
 *
 * Fractional parts of the answer are discarded. That is:
 *      - SquareRoot(3) --> 1
 *      - SquareRoot(4) --> 2
 *      - SquareRoot(5) --> 2
 *      - SquareRoot(8) --> 2
 *      - SquareRoot(9) --> 3
 */
func sqrt(a uint64) uint32 {
	op := a
	res := uint64(0)
	// The second-to-top bit is set: use 1 << 14 for uint16; use 1 << 30 for
	// uint32.
	one := uint64(1) << 62

	// "one" starts at the highest power of four <= than the argument.
	for one > op {
		one >>= 2
	}

	for one != 0 {
		if op >= res+one {
			op = op - (res + one)
			res = res + 2*one
		}
		res >>= 1
		one >>= 2
	}
	return uint32(res)
}

func (a Int) Sqrt() Int {
	// float square root - faster but (potential) non-deterministic
	// res := math.Sqrt(float64(a.Val))
	// if math.IsNaN(res) {
	//	panic(fmt.Errorf("sqrt failed (got NaN) for: %d", a.Val))
	// }
	// return Int{int64(res)}

	// int square root - 5 times slower than floats, but deterministic
	return Int{int64(sqrt(uint64(a.Val)))}
}

func MinMax(a, b Int) (Int, Int) {
	if a.Lt(b) {
		return a, b
	} else {
		return b, a
	}
}

func Min(a, b Int) Int {
	if a.Lt(b) {
		return a
	} else {
		return b
	}
}

func Max(a, b Int) Int {
	if a.Lt(b) {
		return b
	} else {
		return a
	}
}

func (a *Int) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &a.Val)
}

func (a Int) MarshalYAML() ([]byte, error) {
	s := strconv.FormatInt(a.Val, 10)
	return []byte(s), nil
}

func (a *Int) UnmarshalYAML(b []byte) error {
	s := string(b)
	n, err := fmt.Sscan(s, &a.Val)
	if n != 1 {
		Check(fmt.Errorf("failed to get exactly 1 int64 from string %s", s))
	}
	return err
}
