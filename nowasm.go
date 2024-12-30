//go:build !(js && wasm)

package main

import (
	"github.com/go-vgo/robotgo"
	. "github.com/marisvali/miln/gamelib"
)

func getUsername() string {
	return "vali-dev"
}

func moveCursor(pt Pt) {
	robotgo.Move(pt.X.ToInt(), pt.Y.ToInt())
}
