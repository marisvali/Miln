package analysis

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract(t *testing.T) {
	dir := "d:\\gms\\Miln\\research\\analysis\\chunk"
	entries, err := os.ReadDir(dir)
	Check(err)

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		data := ReadFile(fullPath)
		if len(data) > 0 { // Games that never start are null.
			playthrough := DeserializePlaythrough(data)

			w := NewWorld(playthrough.Seed)
			for _, input := range playthrough.History {
				w.Step(input)
			}

			fmt.Printf("%s  %010d  ", entry.Name(), playthrough.Seed)
			if w.Player.Health.IsPositive() {
				fmt.Printf("won, life leftover: %d", w.Player.Health.ToInt())
			} else {
				fmt.Printf("DIED")
			}
			fmt.Println()
		}
	}
	assert.True(t, true)
}
