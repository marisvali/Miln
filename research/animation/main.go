package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	_ "image/png"
	"log"
)

type Game struct {
	count int
	a     Animation
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(100, 100)
	op.GeoM.Scale(0.6, 0.6)
	screen.DrawImage(g.a.Img(), op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 800
}

func (g *Game) Update() error {
	return nil
}

func main() {
	g := Game{}
	g.a = NewAnimation("PickleCarrotHealthDeath")

	ebiten.SetWindowSize(800, 800)
	ebiten.SetWindowTitle("Animation")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
