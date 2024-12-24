package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	_ "image/png"
	"log"
)

type Game struct {
	count        int
	frameIdx     int
	layoutWidth  int
	layoutHeight int
	img          *ebiten.Image
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.frameIdx++
	screen.Fill(color.RGBA{150, 150, 150, 255})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, 1)
	op.GeoM.Translate(6, 3)
	screen.DrawImage(g.img, op)

	// osX, osY := robotgo.Location()
	// gameX, gameY := ebiten.CursorPosition()
	// ebitenutil.DebugPrint(screen,
	// 	fmt.Sprintf("os x, y: %d %d\n"+
	// 		"ebiten game x, y: %d %d\n",
	// 		osX, osY,
	// 		gameX, gameY,
	// 	))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.layoutWidth = 1000
	g.layoutHeight = 1000
	return g.layoutWidth, g.layoutHeight
}

func (g *Game) Update() error {
	return nil
}

func main() {
	g := Game{}
	g.img = ebiten.NewImage(800, 800)
	g.img.Fill(color.RGBA{0, 0, 255, 255})
	for i := 0; i < g.img.Bounds().Dx(); i++ {
		for j := 0; j < 100; j++ {
			g.img.Set(i+j, i, color.RGBA{255, 0, 0, 255})
		}
	}

	ebiten.SetWindowSize(300, 300)
	ebiten.SetWindowTitle("Layout")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
