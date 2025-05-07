package gamelib

import (
	"image/color"
	"math/rand"
	"testing"
)

// BenchmarkCol tests how fast Col() is, so we can compare with color.NRGBA{}.
// 2.056 ns/op
func BenchmarkCol(b *testing.B) {
	// Initialize.
	rr := uint8(rand.Intn(100))
	rg := uint8(rand.Intn(100))
	rb := uint8(rand.Intn(100))
	ra := uint8(rand.Intn(100)) + 100
	c := Col(rr, rg, rb, ra)

	// Run benchmark loop.
	for b.Loop() {
		c.RGBA()
	}
}

// BenchmarkNRGBA tests how fast color.NRGBA{} is, so we can compare with Col().
// 3.271 ns/op
// Compiler optimizations can mess up the benchmarks. If the conditions are
// right, the benchmark will show 0.2667 ns/op. But that only happens in
// conditions that almost never hold in regular code, which makes that benchmark
// useless. We need to make sure these conditions don't hold in the benchmark
// otherwise we don't get the actual execution time of color.NRGBA.RGBA()
// function, we get the execution time of a pre-calculated version of it.
// The conditions are:
// 1. Constant values. If the compiler can guess what the color will be in
// advance, it can optimize the call to color.RGBA(). So creating a color object
// with constant values and using it in the same function is not indicative of
// real performance. In real code you usually pass colors to functions who then
// use the color, and they can't predict and optimize ahead of time.
// 2. No interfaces. If the variable is of type color.Color and you assign a
// constant value of color.NRGBA type, the compiler optimization goes away.
func BenchmarkNRGBA(b *testing.B) {
	// Initialize.
	rr := uint8(rand.Intn(100))
	rg := uint8(rand.Intn(100))
	rb := uint8(rand.Intn(100))
	ra := uint8(rand.Intn(100)) + 100
	var c color.Color
	c = color.NRGBA{rr, rg, rb, ra}

	// Run benchmark loop.
	for b.Loop() {
		c.RGBA()
	}
}
