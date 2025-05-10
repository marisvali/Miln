package world

import (
	"bytes"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_WorldRegression1(t *testing.T) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20250319-170648.mln010"))
	expected := string(ReadFile("playthroughs/20250319-170648.mln010-hash"))

	// Run the playthrough.
	w := NewWorld(playthrough.Seed, LoadWorldData(os.DirFS("playthroughs")))
	for _, input := range playthrough.History {
		w.Step(input)
	}

	println(w.RegressionId())

	assert.Equal(t, expected, w.RegressionId())
}

func RunPlaythrough(p Playthrough) {
	w := NewWorld(p.Seed, LoadWorldData(os.DirFS("playthroughs")))
	for _, input := range p.History {
		w.Step(input)
	}
}

func BenchmarkPlaythroughSpeed(b *testing.B) {
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20250510-104800.mln999"))
	for b.Loop() {
		RunPlaythrough(playthrough)
	}
}

// The benchmarks below are relevant mostly in relation to each other, to
// answer the question: how long does world serialization take and can it be
// performed every frame without impacting the FPS?

// Typical output:
// BenchmarkSerializedPlaythrough_WithoutCompression-12    	 145	  8.246808 ms/op
// BenchmarkSerializedPlaythrough_Compression-12    	     218	  5.373649 ms/op
// BenchmarkWorldClone-12    	   						    5740	  0.201694 ms/op
// Conclusion: serializing is too expensive to perform every frame, better to
// just clone the world and send it to a go routine once every K frames, where
// it can then be serialized and saved to a file or uploaded to a database.

// Get large world (17122 frames, 285 sec).
// Note: I don't really care about the final state of the world, I just need a
// world that ran for a relatively long time.

func GetLargeWorld() World {
	// Load the playthrough.
	playthrough := DeserializePlaythrough(ReadFile("playthroughs/20250510-104800.mln999"))

	// Run the playthrough.
	w := NewWorld(playthrough.Seed, LoadWorldData(os.DirFS("playthroughs")))
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
	for b.Loop() {
		// Serialize.
		buf := new(bytes.Buffer)
		Serialize(buf, int64(Version))
		Serialize(buf, w.Playthrough)
	}
}

// Check how much time it takes to compress a serialized world.
func BenchmarkSerializedPlaythrough_Compression(b *testing.B) {
	// Initialize.
	w := GetLargeWorld()

	// Serialize the world to buf.
	buf := new(bytes.Buffer)
	Serialize(buf, int64(Version))
	Serialize(buf, w.Playthrough)

	// Run benchmark loop.
	for b.Loop() {
		Zip(buf.Bytes())
	}
}

func BenchmarkWorldClone(b *testing.B) {
	// Initialize.
	w := GetLargeWorld()

	// Run benchmark loop.
	for b.Loop() {
		w.Clone()
	}
}
