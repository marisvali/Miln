package gamelib

import (
	"fmt"
	"math/rand"
	"time"
)

var randomGenerator *rand.Rand

func init() {
	// randomGenerator = rand.New(rand.NewSource(0))
	randomGenerator = rand.New(rand.NewSource(time.Now().Unix()))
}

func RSeed(seed Int) {
	randomGenerator = rand.New(rand.NewSource(seed.ToInt64()))
}

// RInt returns a random number in the interval [min, max].
// min must be smaller than max.
// The difference between min and max must be at most max.MaxInt64 - 1.
func RInt(min Int, max Int) Int {
	if max.Lt(min) {
		panic(fmt.Errorf("min larger than max: %d %d", min, max))
	}

	dif := max.Minus(min).Plus(I(1)) // this will panic if the difference
	// between min and max is greater than max.MaxInt64 - 1

	randomValue := I64(randomGenerator.Int63())
	return randomValue.Mod(dif).Plus(min)
}

// RElem returns a random element from a slice.
func RElem[T any](s []T) T {
	return s[RInt(I(0), I(len(s)-1)).ToInt()]
}
