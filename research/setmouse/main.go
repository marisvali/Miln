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
	x            int
	y            int
	screen_x     int
	screen_y     int
	my_x         int
	my_y         int
	layoutWidth  int
	layoutHeight int
}

func OsToGameFirstDraft(osX, osY, layoutWidth, layoutHeight int) (int, int) {
	// Compute the scale factor used to scale between layout size and window
	// size.
	windowWidth, windowHeight := ebiten.WindowSize()
	scaleX := float64(windowWidth) / float64(layoutWidth)
	scaleY := float64(windowHeight) / float64(layoutHeight)
	scale := math.Min(scaleX, scaleY)

	// Compute the width and height of the game region, in terms of OS pixels.
	gameWidth := float64(layoutWidth) * scale
	gameHeight := float64(layoutHeight) * scale
	offsetX := (float64(windowWidth) - gameWidth) / 2
	offsetY := (float64(windowHeight) - gameHeight) / 2

	wpx, wpy := ebiten.WindowPosition()
	wmx := int((float64(osX-wpx) - offsetX) / scale)
	wmy := int((float64(osY-wpy) - offsetY) / scale)
	return wmx, wmy
}

func OsToGame(osX, osY, layoutWidth, layoutHeight int) (gameX, gameY int) {
	// Coordinate systems:
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
	//
	// This function converts a position expressed in the OS coordinate system
	// to a position expressed in the game's coordinate system.
	// Basically it transforms what is returned by robotgo.Location() to match
	// what is returned by ebiten.CursorPosition().

	// OS -> Window
	windowOffsetX, windowOffsetY := ebiten.WindowPosition()
	windowX := osX - windowOffsetX
	windowY := osY - windowOffsetY

	// Window -> Drawn region
	// Compute the size of the drawn region.
	windowWidth, windowHeight := ebiten.WindowSize()
	scaleX := float64(windowWidth) / float64(layoutWidth)
	scaleY := float64(windowHeight) / float64(layoutHeight)
	scale := math.Min(scaleX, scaleY)
	drawnWidth := int(float64(layoutWidth) * scale)
	drawnHeight := int(float64(layoutHeight) * scale)
	drawnOffsetX := (windowWidth - drawnWidth) / 2
	drawnOffsetY := (windowHeight - drawnHeight) / 2
	drawnX := windowX - drawnOffsetX
	drawnY := windowY - drawnOffsetY

	// Drawn -> Game
	gameX = drawnX * layoutWidth / drawnWidth
	gameY = drawnY * layoutHeight / drawnHeight
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
	// i := g.frameIdx % 1000
	// robotgo.Move(i, i)
	screen.Fill(color.RGBA{150, 150, 150, 255})
	ebitenutil.DebugPrint(screen,
		fmt.Sprintf("os x, y: %d %d\n"+
			"ebiten game x, y: %d %d\n"+
			"my game x, y: %d %d\n",
			osX, osY,
			gameX, gameY,
			myGameX, myGameY))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.layoutWidth = 1000
	g.layoutHeight = 200
	return g.layoutWidth, g.layoutHeight
}

func (g *Game) Update() error {
	return nil
}

func main() {
	g := Game{}
	ebiten.SetWindowSize(1390, 800)
	// ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Mouse position")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
