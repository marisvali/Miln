package world

import (
	"bytes"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Q: Is serialization at least self-consistent? If I serialize, deserialize
// then serialize back, do I get the original thing? What about if I
// deserialize, serialize and deserialize?
func TestSerializationForSelfConsistency(t *testing.T) {
	p1 := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln17-17"))
	data1 := p1.Serialize()
	p2 := DeserializePlaythrough(data1)
	data2 := p2.Serialize()

	// assert.Equal does a deep compare here and checks equality the way I want
	// it checked. See reflect.DeepEqual().
	assert.Equal(t, p1, p2)
	assert.Equal(t, data1, data2)
}

// The benchmarks below are relevant mostly in relation to each other, to
// answer the question: how long does playthrough serialization take and can it
// be performed every frame without impacting the FPS?

// Typical output:
// BenchmarkSerializedPlaythrough_WithoutCompression-12    	 268       4.138422 ms/op
// BenchmarkSerializedPlaythrough_Compression-12    	      50	  22.555374 ms/op
// BenchmarkPlaythroughClone-12    	    					5536	   0.218937 ms/op
// Conclusion: serializing is too expensive to perform every frame, better to
// just clone the playthrough and send it to a go routine once every K frames,
// where it can then be serialized and saved to a file or uploaded to a
// database.

// Check how much time it takes to serialize a playthrough (without compressing
// it).
func BenchmarkSerializedPlaythrough_WithoutCompression(b *testing.B) {
	// Initialize, get large playthrough.
	p := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln17-17"))

	// Run benchmark loop.
	for b.Loop() {
		buf := new(bytes.Buffer)
		Serialize(buf, p.InputVersion)
		Serialize(buf, p.SimulationVersion)
		Serialize(buf, p.ReleaseVersion)
		Serialize(buf, p.Level)
		Serialize(buf, p.Id)
		Serialize(buf, p.Seed)
		SerializeSlice(buf, p.History)
	}
}

// Check how much time it takes to compress a serialized world.
func BenchmarkSerializedPlaythrough_Compression(b *testing.B) {
	// Initialize, get large playthrough.
	p := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln17-17"))

	// Serialize the world to buf.
	buf := new(bytes.Buffer)
	Serialize(buf, p.InputVersion)
	Serialize(buf, p.SimulationVersion)
	Serialize(buf, p.ReleaseVersion)
	Serialize(buf, p.Level)
	Serialize(buf, p.Id)
	Serialize(buf, p.Seed)
	SerializeSlice(buf, p.History)

	// Run benchmark loop.
	for b.Loop() {
		Zip(buf.Bytes())
	}
}

func BenchmarkPlaythroughClone(b *testing.B) {
	// Initialize, get large playthrough.
	p := DeserializePlaythrough(ReadFile("playthroughs/large-playthrough.mln17-17"))

	// Run benchmark loop.
	res := 0
	for b.Loop() {
		p2 := p.Clone()
		res += len(p2.History)
	}
}
