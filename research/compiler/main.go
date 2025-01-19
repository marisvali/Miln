package main

import (
	"fmt"
	"time"
)

// Question: does the Go compiler perform optimizations that remove code when it
// deems it unnecessary?
//
// Answer: yes, the Go compiler does do optimizations that modify the code if
// it detects that the code has no impact on the output of the program.
//

// Background: this is important for benchmarks. If I benchmark a piece of
// code, do I need to make sure that the code has some effect that I accumulate
// and print at the end, otherwise my code will be simplified by the compiler,
// because it doesn't 'contribute' anything?

// Typical output (release mode):
// Result: 249952502699955000
// Duration (no optimization): 0.799730
// Duration (possible optimization 1): 0.530237
// Result: 17
// Duration (possible optimization 2): 0.264161

// Another thing. If I don't overwrite v3's value with 17, I get this output
// (release mode):
// Result: 249952502699955000
// Duration (no optimization): 0.810908
// Duration (possible optimization 1): 0.265336
// Result: 249952502699955000
// Duration (possible optimization 2): 0.793616

// This makes no sense to me at first sight. How can the computation I do in
// loop 3 affect how long loop 2 takes?
// Without analyzing assembly I won't know what computations are actually
// performed. And without reading in detail about compiler optimizations I won't
// understand how and why it reached that assembly.
// Both of these directions are too involved for my current purposes.

// Conclusion: the Go compiler does do optimizations that modify the code if
// it detects that the code has no impact on the output of the program.

func main() {
	// Control the duration of the loops with these 2 variables.
	ni := 10000
	nj := 100000

	// No optimization
	// -------------------------------------------------------------------------
	start := time.Now()
	// Do some time-consuming computation.
	var v int
	for i := 1; i < ni; i++ {
		for j := 1; j < nj; j++ {
			v += i*j + i - j/2
		}
	}
	duration := time.Now().Sub(start).Seconds()
	fmt.Printf("Result: %d\n", v)
	fmt.Printf("Duration (no optimization): %f\n", duration)

	// Possible optimization 1 (result not used)
	// -------------------------------------------------------------------------
	start = time.Now()
	// Do some time-consuming computation.
	var v2 int
	for i := 1; i < ni; i++ {
		for j := 1; j < nj; j++ {
			v2 += i*j + i - j/2
		}
	}
	// Don't do anything with the result (v2).
	duration = time.Now().Sub(start).Seconds()
	fmt.Printf("Duration (possible optimization 1): %f\n", duration)

	// Possible optimization 2 (loop not relevant)
	// -------------------------------------------------------------------------
	start = time.Now()
	// Do some time-consuming computation.
	var v3 int
	for i := 1; i < ni; i++ {
		for j := 1; j < nj; j++ {
			v3 += i*j + i - j/2
		}
	}
	// v3 = 17 // Make the previous computation useless.
	fmt.Printf("Result: %d\n", v3)
	duration = time.Now().Sub(start).Seconds()
	fmt.Printf("Duration (possible optimization 2): %f\n", duration)
}
