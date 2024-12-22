package main

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	_ "image/png"
	"log"
	"math"
)

type Game struct {
	count        int
	frameIdx     int
	layoutWidth  int
	layoutHeight int
}

// OsToGame converts an (x, y) position from the "OS coordinate system" to the
// "Game coordinate system". See GameToOs for an explanation of what these
// coordinate systems are.
//
// Basically it transforms what is returned by robotgo.Location() to match
// what is returned by ebiten.CursorPosition(). The main use of this function
// is to check that the conversion from OS to Game is correct (matches what
// is returned by ebiten.CursorPosition()), so that we can then implement
// GameToOs by reversing the operations.
func OsToGame(osX, osY, layoutWidth, layoutHeight int) (gameX, gameY int) {
	// Compute the size of the drawn region.
	windowWidth, windowHeight := ebiten.WindowSize()
	scaleX := float64(windowWidth) / float64(layoutWidth)
	scaleY := float64(windowHeight) / float64(layoutHeight)
	scale := math.Min(scaleX, scaleY)
	drawnWidth := int(float64(layoutWidth) * scale)
	drawnHeight := int(float64(layoutHeight) * scale)

	// OS -> Window
	windowOffsetX, windowOffsetY := ebiten.WindowPosition()
	windowX := osX - windowOffsetX
	windowY := osY - windowOffsetY

	// Window -> Drawn region
	drawnOffsetX := (windowWidth - drawnWidth) / 2
	drawnOffsetY := (windowHeight - drawnHeight) / 2
	drawnX := windowX - drawnOffsetX
	drawnY := windowY - drawnOffsetY

	// Drawn -> Game
	gameX = drawnX * layoutWidth / drawnWidth
	gameY = drawnY * layoutHeight / drawnHeight
	return
}

// GameToOs converts an (x, y) position from the "Game coordinate system" to the
// "OS coordinate system". See below for an explanation of what these coordinate
// systems are.
//
// OS coordinate system: (0, 0) is the top-left of the monitor, x, y is the
// number of pixels to the right and down from that corner. If the OS has
// a resolution of 1920x1080 then the most bottom-right pixel is
// (1919, 1079).
//
// Window coordinate system: (0, 0) is the top-left pixel in the window
// spawned when the game is started. The size of this area is set and
// retrieved using ebiten.SetWindowSize() and ebiten.WindowSize(). The
// position of this area within the OS coordinate system is set and
// retrieved using ebiten.SetWindowPosition() and ebiten.WindowPosition().
// This window contains the game's drawn region. A pixel in this coordinate
// system has the same size as in the OS. So if ebiten.WindowSize() returns
// (13, 25) and ebiten.WindowPosition() returns (20, 30), then the
// bottom-right pixel in the window is (12, 24), corresponding to pixel
// (32, 54) in the OS coordinate system.
//
// Drawn region coordinate system: (0, 0) is the top-left pixel inside the
// game's window that is actually drawn. A pixel in this coordinate system
// has the same size as in the Window coordinate system and OS. The drawn
// region has its width or height equal to the window, but the other
// dimension is equal or smaller. This is so that the drawn region always
// fits inside the window. The dimensions of the drawn region depend on
// what the game's Layout() function returns. If Layout() returns a width
// and height proportional to the width and height returned by WindowSize(),
// then the drawn region will fill the window perfectly.
//
// Game coordinate system: (0, 0) is the top-left pixel inside the game's
// window that is actually drawn. Layout() returns the number of pixels in
// this coordinate system. A pixel in this coordinate system is not the same
// as in the OS coordinate system. It is scaled so that layout width matches
// the drawn region's width and the layout height matches the drawn region's
// height.
func GameToOs(gameX, gameY, layoutWidth, layoutHeight int) (osX, osY int) {
	// Compute the size of the drawn region.
	windowWidth, windowHeight := ebiten.WindowSize()
	scaleX := float64(windowWidth) / float64(layoutWidth)
	scaleY := float64(windowHeight) / float64(layoutHeight)
	scale := math.Min(scaleX, scaleY)
	drawnWidth := int(float64(layoutWidth) * scale)
	drawnHeight := int(float64(layoutHeight) * scale)

	// Game -> Drawn region
	drawnX := gameX * drawnWidth / layoutWidth
	drawnY := gameY * drawnHeight / layoutHeight

	// Drawn region -> Window
	drawnOffsetX := (windowWidth - drawnWidth) / 2
	drawnOffsetY := (windowHeight - drawnHeight) / 2
	windowX := drawnX + drawnOffsetX
	windowY := drawnY + drawnOffsetY

	// Window -> OS
	windowOffsetX, windowOffsetY := ebiten.WindowPosition()
	osX = windowX + windowOffsetX
	osY = windowY + windowOffsetY
	return
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.frameIdx++
	osX, osY := robotgo.Location()
	gameX, gameY := ebiten.CursorPosition()
	// The purpose here is to obtain the same result as CursorPosition.
	// If I can do this, then I can be confident in implementing a reverse
	// function, GameToOs, which I can use to replicate a player's recorded
	// movements on my own screen.
	myGameX, myGameY := OsToGame(osX, osY, g.layoutWidth, g.layoutHeight)
	myOsX, myOsY := GameToOs(myGameX, myGameY, g.layoutWidth, g.layoutHeight)
	// i := g.frameIdx % 1000
	// robotgo.Move(i, i)
	screen.Fill(color.RGBA{150, 150, 150, 255})
	ebitenutil.DebugPrint(screen,
		fmt.Sprintf("os x, y: %d %d\n"+
			"ebiten game x, y: %d %d\n"+
			"my game x, y: %d %d\n"+
			"my os x, y: %d %d\n",
			osX, osY,
			gameX, gameY,
			myGameX, myGameY,
			myOsX, myOsY))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.layoutWidth = 200
	g.layoutHeight = 650
	return g.layoutWidth, g.layoutHeight
}

func (g *Game) Update() error {
	return nil
}

func main() {
	g := Game{}
	ebiten.SetWindowSize(800, 500)
	// ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Mouse position")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
