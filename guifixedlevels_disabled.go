//go:build nofixedlevels

package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

func (g *Gui) InitializeFixedLevels() {
}

func (g *Gui) GetCurrentFixedLevel() (seed Int, l Level) {
	return
}

func (g *Gui) AdvanceCurrentFixedLevel() {
}

func (g *Gui) UsingFixedLevels() bool {
	return false
}

func (g *Gui) HasMoreFixedLevels() bool {
	return false
}
