package world

import (
	"bytes"
	"embed"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_WorldRegression1(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20250319-170648.mln010"))
	expected := string(ReadFile("playthroughs/20250319-170648.mln010-hash"))

	// Run the playthrough.
	w := NewWorld(playthrough.Seed, playthrough.TargetDifficulty, nil)
	for _, input := range playthrough.History {
		w.Step(input)
	}

	println(w.RegressionId())

	assert.Equal(t, expected, w.RegressionId())
}

func RunPlaythrough(p Playthrough) {
	w := NewWorld(p.Seed, p.TargetDifficulty, nil)
	for _, input := range p.History {
		w.Step(input)
	}
}

func BenchmarkPlaythroughSpeed(b *testing.B) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20240804-073523.mln007"))
	for n := 0; n < b.N; n++ {
		RunPlaythrough(playthrough)
	}
}

// The benchmarks below are relevant mostly in relation to each other, to
// answer the question: how long does world serialization take and can it be
// performed every frame without impacting the FPS?

// Typical output:
// BenchmarkSerializedPlaythrough_WithoutCompression-12		304		3.953169 ms/op
// BenchmarkSerializedPlaythrough_Compression-12            144     8.276954 ms/op
// BenchmarkWorldClone-12              					   6336     0.194385 ms/op
// Conclusion: serializing is too expensive to perform every frame, better to
// just clone the world and send it to a go routine once every K frames, where
// it can then be serialized and saved to a file or uploaded to a database.

// Get large world (18000 frames, 300 sec).
// Note: I don't really care about the final state of the world, I just need a
// world that ran for a relatively long time.
//
//go:embed playthroughs/data/*
var embeddedFiles embed.FS

func GetLargeWorld() World {
	// Load the playthrough.
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20250118-180923.mln"))

	// Run the playthrough.
	etc, _ := os.Getwd()
	fmt.Println(etc)
	// subFs := fs.Sub(embeddedFiles, "playthrough")
	w := NewWorld(playthrough.Seed, playthrough.TargetDifficulty, &embeddedFiles)
	for _, input := range playthrough.History {
		w.Step(input)
	}
	return w
}

// Check how much time it takes to serialize a world (without compressing it).
func BenchmarkSerializedPlaythrough_WithoutCompression(b *testing.B) {
	// Initialize.
	w := GetLargeWorld()

	// Run benchmark loop.
	b.ResetTimer()
	x := byte(0)
	for n := 0; n < b.N; n++ {
		// Serialize.
		buf := new(bytes.Buffer)
		Serialize(buf, int64(Version))
		Serialize(buf, w.Seed.ToInt64())
		Serialize(buf, w.TargetDifficulty.ToInt64())
		SerializeSlice(buf, w.History)

		// Accumulate result for final side effect.
		data := buf.Bytes()
		x += data[len(data)/3]
	}

	// Final side effect.
	fmt.Println(x)
	fmt.Println(b.N)
}

// Check how much time it takes to compress a serialized world.
func BenchmarkSerializedPlaythrough_Compression(b *testing.B) {
	// Initialize.
	w := GetLargeWorld()

	// Serialize the world to buf.
	buf := new(bytes.Buffer)
	Serialize(buf, int64(Version))
	Serialize(buf, w.Seed.ToInt64())
	Serialize(buf, w.TargetDifficulty.ToInt64())
	SerializeSlice(buf, w.History)

	// Run benchmark loop.
	b.ResetTimer()
	x := byte(0)
	for n := 0; n < b.N; n++ {
		data := Zip(buf.Bytes())

		// Accumulate result for final side effect.
		x += data[len(data)/2]
	}

	// Final side effect.
	fmt.Println(x)
	fmt.Println(b.N)
}

func BenchmarkWorldClone(b *testing.B) {
	// Initialize.
	w := GetLargeWorld()

	// Run benchmark loop.
	b.ResetTimer()
	x := 0
	for n := 0; n < b.N; n++ {
		w1 := w.Clone()

		// Accumulate result for final side effect.
		p := w1.History[len(w1.History)/2]
		x += p.MousePt.X.ToInt()
	}

	// Final side effect.
	fmt.Println(x)
	fmt.Println(b.N)
}
