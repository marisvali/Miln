package main

import "testing"

func BenchmarkDoItAll(b *testing.B) {
	for b.Loop() {
		DoItAll()
	}
}
