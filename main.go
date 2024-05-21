package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image/color"
	"math/rand"
	"strconv"
)

type Player struct {
	Pos Pt
}

type World struct {
	Player    Player
	Obstacles Matrix
}

type Gui struct {
	defaultFont font.Face
	img         *ebiten.Image
	world       World
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func (g *Gui) Update() error {
	return nil
}

func colorHex(hexVal int) color.Color {
	if hexVal < 0x000000 || hexVal > 0xFFFFFF {
		panic(fmt.Sprintf("Invalid HEX value for color: %d", hexVal))
	}
	r := uint8(hexVal & 0xFF0000 >> 16)
	g := uint8(hexVal & 0x00FF00 >> 8)
	b := uint8(hexVal & 0x0000FF)
	return color.RGBA{
		R: r,
		G: g,
		B: b,
		A: 255,
	}
}

func (g *Gui) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 0})
	//message := "PAUSED"
	//text.Draw(screen, message, g.defaultFont, 60, 60, colorHex(0xee005a))
	g.img.Fill(intToCol(3))

	sz := screen.Bounds().Size()
	numX := 15
	numY := 15
	blockWidth := float64(sz.X) / float64(numX)
	blockHeight := float64(sz.Y) / float64(numY)
	margin := float64(1)
	for iy := 0; iy < numX; iy++ {
		for ix := 0; ix < numX; ix++ {
			posX := float64(ix) * blockWidth
			posY := float64(iy) * blockHeight
			DrawSprite(screen, g.img, posX+margin, posY+margin, blockWidth-2*margin, blockHeight-2*margin)
		}
	}

}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func intToCol(ival int64) color.Color {
	switch ival {
	case 0:
		return color.RGBA{25, 25, 25, 0}
	case 1:
		return color.RGBA{150, 0, 0, 0}
	case 2:
		return color.RGBA{0, 150, 0, 0}
	case 3:
		return color.RGBA{0, 0, 150, 0}
	case 4:
		return color.RGBA{150, 150, 0, 0}
	case 5:
		return color.RGBA{0, 150, 150, 0}
	case 6:
		return color.RGBA{150, 0, 150, 0}
	case 7:
		return color.RGBA{100, 150, 100, 0}
	}
	return color.Black
}

//func (g *Game) DrawFilledSquare(screen *ebiten.Image, s Square, col color.Color) {
//	size := WorldToScreen(s.Size)
//	x := WorldToScreen(s.Center.X) - size/2
//	y := WorldToScreen(s.Center.Y) - size/2
//
//	if g.img == nil {
//		g.img = ebiten.NewImage(int(size), int(size))
//	}
//	g.img.Fill(col)
//	op := &ebiten.DrawImageOptions{}
//	op.GeoM.Translate(x, y)
//	screen.DrawImage(g.img, op)
//}
//
//func DrawSquare(screen *ebiten.Image, s Square, color color.Color) {
//	halfSize := s.Size.DivBy(I(2)).Plus(s.Size.Mod(I(2)))
//
//	// square corners
//	upperLeftCorner := Pt{s.Center.X.Minus(halfSize), s.Center.Y.Minus(halfSize)}
//	lowerLeftCorner := Pt{s.Center.X.Minus(halfSize), s.Center.Y.Plus(halfSize)}
//	upperRightCorner := Pt{s.Center.X.Plus(halfSize), s.Center.Y.Minus(halfSize)}
//	lowerRightCorner := Pt{s.Center.X.Plus(halfSize), s.Center.Y.Plus(halfSize)}
//
//	DrawLine(screen, Line{upperLeftCorner, upperRightCorner}, color)
//	DrawLine(screen, Line{upperLeftCorner, lowerLeftCorner}, color)
//	DrawLine(screen, Line{lowerLeftCorner, lowerRightCorner}, color)
//	DrawLine(screen, Line{lowerRightCorner, upperRightCorner}, color)
//}
//
//func (g *Game) DrawMatrix(screen *ebiten.Image, m Matrix, squareSize Int) {
//	for y := I(0); y.Lt(m.NRows()); y.Inc() {
//		for x := I(0); x.Lt(m.NCols()); x.Inc() {
//			var s Square
//			s.Center.X = x.Times(squareSize).Plus(squareSize.DivBy(I(2)))
//			s.Center.Y = y.Times(squareSize).Plus(squareSize.DivBy(I(2)))
//			s.Size = squareSize
//
//			var col color.Color
//			col = color.RGBA{160, 160, 160, 0}
//			DrawSquare(screen, s, col)
//
//			mVal := m.Get(y, x).ToInt64()
//			g.DrawFilledSquare(screen, s, intToCol(mVal))
//		}
//	}
//}

func DrawSprite(screen *ebiten.Image, img *ebiten.Image,
	x float64, y float64, targetWidth float64, targetHeight float64) {
	op := &ebiten.DrawImageOptions{}

	// Resize image to fit the target size we want to draw.
	// This kind of scaling is very useful during development when the final
	// sizes are not decided, and thus it's impossible to have final sprites.
	// For an actual release, scaling should be avoided.
	imgSize := img.Bounds().Size()
	newDx := targetWidth / float64(imgSize.X)
	newDy := targetHeight / float64(imgSize.Y)
	op.GeoM.Scale(newDx, newDy)

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorOne
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorOne
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorZero
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorZero
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func main() {
	for j := 1; j < 10; j++ {
		for i := 1; i < 10; i++ {
			num := 10 + rand.Int()%90
			if (i+j+num)%2 == 0 {
				print("\x1B[0;35m")
			} else {
				print("\x1B[0;33m")
			}
			print(strconv.Itoa(num), " ")
		}
		println()
	}

	ebiten.SetWindowSize(400, 400)
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(700, 100)

	var g Gui
	g.img = ebiten.NewImage(20, 20)

	var err error
	// Load the Arial font
	fontData, err := opentype.Parse(goregular.TTF)
	Check(err)

	g.defaultFont, err = opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	Check(err)

	// Start the game.
	err = ebiten.RunGame(&g)
	Check(err)
}
