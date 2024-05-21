package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image/color"
	. "playful-patterns.com/miln/ints"
	"slices"
)

type Player struct {
	Pos Pt
}

type Enemy struct {
	Pos Pt
}

type World struct {
	Player    Player
	Enemy     Enemy
	Obstacles Matrix
}

type Gui struct {
	defaultFont font.Face
	imgGround   *ebiten.Image
	imgPlayer   *ebiten.Image
	imgEnemy    *ebiten.Image
	world       World
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func (g *Gui) Update() error {
	// Get keyboard input.
	var pressedKeys []ebiten.Key
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)

	// Choose which is the active player based on Alt being pressed.
	moveLeft := slices.Contains(pressedKeys, ebiten.KeyA)
	moveUp := slices.Contains(pressedKeys, ebiten.KeyW)
	moveDown := slices.Contains(pressedKeys, ebiten.KeyS)
	moveRight := slices.Contains(pressedKeys, ebiten.KeyD)

	if moveLeft {
		if g.world.Player.Pos.X.Gt(ZERO) {
			g.world.Player.Pos.X.Dec()
		}
	}

	if moveRight {
		if g.world.Player.Pos.X.Lt(g.world.Obstacles.NCols().Minus(I(1))) {
			g.world.Player.Pos.X.Inc()
		}
	}

	if moveUp {
		if g.world.Player.Pos.Y.Gt(ZERO) {
			g.world.Player.Pos.Y.Dec()
		}
	}

	if moveDown {
		if g.world.Player.Pos.Y.Lt(g.world.Obstacles.NRows().Minus(I(1))) {
			g.world.Player.Pos.Y.Inc()
		}
	}

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

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos Pt) {
	sz := screen.Bounds().Size()
	numX := g.world.Obstacles.NCols().ToInt()
	numY := g.world.Obstacles.NRows().ToInt()
	blockWidth := float64(sz.X) / float64(numX)
	blockHeight := float64(sz.Y) / float64(numY)
	margin := float64(1)
	posX := pos.X.ToFloat64() * blockWidth
	posY := pos.Y.ToFloat64() * blockHeight
	DrawSprite(screen, img, posX+margin, posY+margin, blockWidth-2*margin, blockHeight-2*margin)
}

func (g *Gui) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 0})
	//message := "PAUSED"
	//text.Draw(screen, message, g.defaultFont, 60, 60, colorHex(0xee005a))

	// Draw ground.
	numX := g.world.Obstacles.NCols().ToInt()
	numY := g.world.Obstacles.NRows().ToInt()
	for y := 0; y < numY; y++ {
		for x := 0; x < numX; x++ {
			g.DrawTile(screen, g.imgGround, IPt(x, y))
		}
	}

	// Draw player.
	g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)

	// Draw enemy.
	g.DrawTile(screen, g.imgEnemy, g.world.Enemy.Pos)

	// Output TPS (ticks per second, which is like frames per second).
	ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
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
	ebiten.SetWindowSize(400, 400)
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(700, 100)

	var g Gui
	g.imgGround = ebiten.NewImage(20, 20)
	g.imgGround.Fill(intToCol(0))
	g.imgPlayer = ebiten.NewImage(20, 20)
	g.imgPlayer.Fill(intToCol(1))
	g.imgEnemy = ebiten.NewImage(20, 20)
	g.imgEnemy.Fill(intToCol(2))

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

	//g.world.Obstacles.Init(I(15), I(15))
	g.world.Obstacles.Init(I(15), I(15))
	g.world.Enemy.Pos = IPt(5, 3)

	// Start the game.
	err = ebiten.RunGame(&g)
	Check(err)
}
