//go:build js && wasm

package main

import (
	. "github.com/marisvali/miln/gamelib"
	"syscall/js"
)

func getUsername() string {
	// Retrieve parameter from JavaScript global scope.
	return js.Global().Get("username").String()
}

func moveCursor(pt Pt) {
	// Do nothing in WASM builds, because the browser doesn't allow a WASM to
	// control the user's mouse (and it shouldn't).
	// This is here because I want to move the cursor when running Miln on the
	// desktop and reviewing playthroughs. But I want the code that calls
	// moveCursor to compile for WASM as well so I need to provide a no-op
	// variant for this function.
}
