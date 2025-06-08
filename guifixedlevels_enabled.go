//go:build fixedlevels

package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

// Functions for working with fixed levels. Apparently these are the abstract
// functionalities I need:
// - initialize fixed levels
// - check if we are currently using fixed levels
// - check if we have any more fixed levels
// - get current fixed level
// - advance current fixed level

func (g *Gui) InitializeFixedLevels() {
	g.fixedLevels = GetFiles(g.FSys, "data/levels", "*")
	g.LoadUserData()
}

func (g *Gui) GetCurrentFixedLevel() (seed Int, l Level) {
	return LoadLevelFromYAML(g.FSys, g.fixedLevels[g.CurrentFixedLevelIdx.ToInt()])
}

func (g *Gui) AdvanceCurrentFixedLevel() {
	g.CurrentFixedLevelIdx.Inc()
	g.SaveUserData()
}

func (g *Gui) UsingFixedLevels() bool {
	return len(g.fixedLevels) > 0
}

func (g *Gui) HasMoreFixedLevels() bool {
	return g.CurrentFixedLevelIdx.Lt(I(len(g.fixedLevels)))
}
