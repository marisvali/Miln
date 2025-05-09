package main

import (
	"fmt"
	"testing"
)

// Question: When I benchmark TimeConsumingFunction(), do I have to make sure
// it has a side effect visible at the end of the benchmark, to get a reliable
// result?
// Answer: Yes.
//
// The benchmarks below all run TimeConsumingFunction() and report the following
// durations for it (all tests run in release mode):
// BenchmarkNoOptimization: 7.8 ms (print sum of results at the end)
// BenchmarkPossibleOptimization1: 2.6 ms (do not print anything)
// BenchmarkPossibleOptimization2: 2.6 ms (compute the sum of results but assign
// a constant value to the variable before printing it at the end)

func TimeConsumingFunction() int {
	ni := 1000
	nj := 10000
	var v int
	for i := 1; i < ni; i++ {
		for j := 1; j < nj; j++ {
			v += i*j + i - j/2
		}
	}
	return v
}

// Typical result:
// BenchmarkNoOptimization-12           152           7863170 ns/op
func BenchmarkNoOptimization(b *testing.B) {
	result := 0
	for n := 0; n < b.N; n++ {
		result += TimeConsumingFunction()
	}
	fmt.Println(result)
}

// Typical result:
// BenchmarkPossibleOptimization1-12            446           2651248 ns/op
func BenchmarkPossibleOptimization1(b *testing.B) {
	result := 0
	for n := 0; n < b.N; n++ {
		result += TimeConsumingFunction()
	}
	// Don't print result.
	// fmt.Println(result)
}

// Typical result:
// BenchmarkPossibleOptimization2-12            447           2655188 ns/op
func BenchmarkPossibleOptimization2(b *testing.B) {
	result := 0
	for n := 0; n < b.N; n++ {
		result += TimeConsumingFunction()
	}
	// Make computation irrelevant for the final output of the code.
	result = 17
	fmt.Println(result)
}

// Typical result:
// BenchmarkNoOptimizationWithBLoop-12    	     142	   8356441 ns/op
func BenchmarkNoOptimizationWithBLoop(b *testing.B) {
	result := 0
	for b.Loop() {
		result += TimeConsumingFunction()
	}
	fmt.Println(result)
}

// Typical result:
// BenchmarkPossibleOptimization1WithBLoop-12    	     140	   8392801 ns/op
func BenchmarkPossibleOptimization1WithBLoop(b *testing.B) {
	result := 0
	for b.Loop() {
		result += TimeConsumingFunction()
	}
	// Don't print result.
	// fmt.Println(result)
}

// Typical result:
// BenchmarkPossibleOptimization2WithBLoop-12    	     142	   8312732 ns/op
func BenchmarkPossibleOptimization2WithBLoop(b *testing.B) {
	result := 0
	for b.Loop() {
		result += TimeConsumingFunction()
	}
	// Make computation irrelevant for the final output of the code.
	result = 17
	fmt.Println(result)
}
