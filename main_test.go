package main

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"testing"
)

func IsGameWon(w *World) bool {
	for _, enemy := range w.Enemies {
		if enemy.Alive() {
			return false
		}
	}
	for _, portal := range w.SpawnPortals {
		if portal.Active() {
			return false
		}
	}
	return true
}

func IsGameLost(w *World) bool {
	return w.Player.Health.Leq(ZERO)
}

func IsGameOver(w *World) bool {
	return IsGameWon(w) || IsGameLost(w)
}

func Test_Main(t *testing.T) {
	// Run the playthrough.
	for i := 1; i < 100; i++ {
		w := NewWorld(I(i))
		ai := AI{}
		for {
			input := ai.Step(&w)
			w.Step(input)
			if IsGameOver(&w) {
				break
			}
		}
		if IsGameWon(&w) {
			fmt.Printf("%d %d game won\n", i, w.Seed)
		} else {
			fmt.Printf("%d %d game lost\n", i, w.Seed)
		}
	}
	assert.True(t, true)
}
