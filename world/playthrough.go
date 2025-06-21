package world

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	. "github.com/marisvali/miln/gamelib"
	"slices"
)

// Playthrough represents all the input sent to a World during the execution
// of a level. Given this input and a compatible simulation, the same output
// should be generated in the end.
type Playthrough struct {
	Level
	Id      uuid.UUID
	Seed    Int
	History []PlayerInput
}

func (p *Playthrough) Serialize() []byte {
	buf := new(bytes.Buffer)
	Serialize(buf, int64(Version))
	Serialize(buf, p)
	return Zip(buf.Bytes())
}

func (p *Playthrough) Clone() *Playthrough {
	clone := *p
	clone.History = slices.Clone(p.History)
	return &clone
}

func DeserializePlaythrough(data []byte) (p Playthrough) {
	buf := bytes.NewBuffer(Unzip(data))
	var token int64
	Deserialize(buf, &token)
	// Add a temporary hardcoded rule between versions 13 and 16 as I know they
	// are compatible.
	if token != Version {
		Check(fmt.Errorf("this code can't simulate this playthrough "+
			"correctly - we are version %d and playthrough was generated "+
			"with version %d",
			Version, token))
	}
	Deserialize(buf, &p)
	return
}

// Step is just a utility function if you find yourself repeating the same
// step operation many times, and you need to step a Playthrough and a World
// at the same time.
func Step(p *Playthrough, w *World, input PlayerInput) {
	p.History = append(p.History, input)
	w.Step(input)
}
