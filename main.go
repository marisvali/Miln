package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image/color"
	_ "image/png"
	. "playful-patterns.com/miln/gamelib"
	. "playful-patterns.com/miln/world"
)

var EnemyCooldown Int = I(40)
var PlayerCooldown Int = I(15)
var BlockSize Int = I(80)

type Gui struct {
	defaultFont   font.Face
	imgGround     *ebiten.Image
	imgTree       *ebiten.Image
	imgPlayer     *ebiten.Image
	imgEnemy      *ebiten.Image
	imgBeam       *ebiten.Image
	imgShadow     *ebiten.Image
	world         World
	frameIdx      Int
	screenSize    Pt
	folderWatcher FolderWatcher
}

func (g *Gui) Update() error {
	x, y := ebiten.CursorPosition()
	mousePt := IPt(x, y).DivBy(BlockSize)

	var input PlayerInput
	input.Move = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
	input.MovePt = mousePt
	input.Shoot = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2)
	input.ShootPt = mousePt
	g.world.Step(&input)

	if g.folderWatcher.FolderContentsChanged() {
		g.loadGuiData()
	}

	return nil
}

func (g *Gui) LineObstaclesIntersection(l Line) (bool, Pt) {
	sz := g.screenSize
	numX := g.world.Obstacles.Size().X.ToInt()
	numY := g.world.Obstacles.Size().Y.ToInt()
	blockWidth := sz.X.ToFloat64() / float64(numX)
	blockHeight := sz.Y.ToFloat64() / float64(numY)

	rows := g.world.Obstacles.Size().Y
	cols := g.world.Obstacles.Size().X
	ipts := []Pt{}
	var pt Pt
	for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
		for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
			if !g.world.Obstacles.Get(pt).IsZero() {
				if !EqualFloats(blockWidth, blockHeight) {
					panic(fmt.Errorf("blocks are not squares"))
				}
				s := Square{g.TileToScreen(pt), I(int(blockWidth * 0.9))}
				if intersects, ipt := LineSquareIntersection(l, s); intersects {
					ipts = append(ipts, ipt)
				}
			}
		}
	}

	return GetClosestPoint(ipts, l.Start)
}

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos Pt) {
	sz := screen.Bounds().Size()
	numX := g.world.Obstacles.Size().X.ToInt()
	numY := g.world.Obstacles.Size().Y.ToInt()
	blockWidth := float64(sz.X) / float64(numX)
	blockHeight := float64(sz.Y) / float64(numY)
	margin := float64(1)
	posX := pos.X.ToFloat64() * blockWidth
	posY := pos.Y.ToFloat64() * blockHeight
	DrawSprite(screen, img, posX+margin, posY+margin, blockWidth-2*margin, blockHeight-2*margin)
}

func (g *Gui) TileToScreen(pos Pt) Pt {
	sz := g.screenSize
	numX := g.world.Obstacles.Size().X
	numY := g.world.Obstacles.Size().Y
	blockWidth := sz.X.DivBy(numX)
	blockHeight := sz.Y.DivBy(numY)
	x := pos.X.Times(blockWidth).Plus(blockWidth.DivBy(TWO))
	y := pos.Y.Times(blockHeight).Plus(blockHeight.DivBy(TWO))
	return Pt{x, y}
}

func (g *Gui) WorldToGuiPos(pt Pt) Pt {
	sz := g.screenSize
	numX := g.world.Obstacles.Size().X
	numY := g.world.Obstacles.Size().Y
	blockWidth := sz.X.DivBy(numX)
	blockHeight := sz.Y.DivBy(numY)
	if blockWidth.Neq(blockHeight) {
		panic(fmt.Errorf("blocks are not squares"))
	}
	return pt.Times(blockWidth).DivBy(g.world.BlockSize)
}

func (g *Gui) Draw(screen *ebiten.Image) {
	// Draw background.
	screen.Fill(color.RGBA{0, 0, 0, 0})

	// Draw ground and trees.
	rows := g.world.Obstacles.Size().Y
	cols := g.world.Obstacles.Size().X
	var pt Pt
	for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
		for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
			g.DrawTile(screen, g.imgGround, pt)
			if g.world.Obstacles.Get(pt).Eq(ONE) {
				g.DrawTile(screen, g.imgTree, pt)
			}
		}
	}

	// Draw player.
	mask := ebiten.NewImageFromImage(g.imgPlayer)
	{
		percent := g.world.Player.TimeoutIdx.Times(I(100)).DivBy(PlayerCooldown)
		var alpha Int
		if percent.Gt(ZERO) {
			alpha = (percent.Plus(I(100))).Times(I(255)).DivBy(I(200))
		} else {
			alpha = ZERO
		}

		sz := mask.Bounds().Size()
		for y := 0; y < sz.Y; y++ {
			for x := 0; x < sz.X; x++ {
				_, _, _, a := mask.At(x, y).RGBA()
				if a > 0 {
					mask.Set(x, y, color.RGBA{0, 0, 0, uint8(alpha.ToInt())})
				}
			}
		}

		totalWidth := I(mask.Bounds().Size().X)
		lineWidth := g.world.Player.TimeoutIdx.Times(totalWidth).DivBy(PlayerCooldown)
		l := Line{IPt(0, 0), Pt{lineWidth, I(0)}}
		DrawLine(mask, l, color.RGBA{0, 0, 0, 255})
	}
	g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)
	g.DrawTile(screen, mask, g.world.Player.Pos)
	//g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)

	// Draw enemy.
	g.DrawTile(screen, g.imgEnemy, g.world.Enemy.Pos)

	// Draw beam.
	beamScreen := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	if g.world.Beam.Idx.Gt(ZERO) {
		var beam Line
		if g.world.Beam.Enemy != nil {
			beam = Line{g.TileToScreen(g.world.Player.Pos), g.TileToScreen(g.world.Beam.Enemy.Pos)}
		} else {
			beam = Line{g.TileToScreen(g.world.Player.Pos), g.WorldToGuiPos(g.world.Beam.End)}
		}

		alpha := uint8(g.world.Beam.Idx.Times(I(255)).DivBy(g.world.BeamMax).ToInt())
		colr, colg, colb, _ := g.imgBeam.At(0, 0).RGBA()
		beamCol := color.RGBA{uint8(colr), uint8(colg), uint8(colb), alpha}
		DrawLine(beamScreen, beam, beamCol)
	}
	DrawSprite(screen, beamScreen, 0, 0, float64(beamScreen.Bounds().Dx()), float64(beamScreen.Bounds().Dy()))

	// Mark attackable tiles.
	for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
		for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
			if g.world.AttackableTiles.Get(pt).Neq(ZERO) {
				g.DrawTile(screen, g.imgShadow, pt)
			}
		}
	}

	// Output TPS (ticks per second, which is like frames per second).
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.screenSize = IPt(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
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

func (g *Gui) loadGuiData() {
	// Read from the disk over and over until a full read is possible.
	// This repetition is meant to avoid crashes due to reading files
	// while they are still being written.
	// It's a hack but possibly a quick and very useful one.
	CheckCrashes = false
	for {
		CheckFailed = nil
		g.imgGround = LoadImage("data/ground.png")
		g.imgTree = LoadImage("data/tree.png")
		g.imgPlayer = LoadImage("data/player.png")
		g.imgEnemy = LoadImage("data/enemy.png")
		g.imgBeam = LoadImage("data/beam.png")
		g.imgShadow = LoadImage("data/shadow.png")
		//g.imgShadow.Fill(color.RGBA{0, 0, 0, 100})
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true
}

func main() {
	var g Gui
	g.world.Initialize()

	// screen size
	g.screenSize.X = BlockSize.Times(g.world.Obstacles.Size().X)
	g.screenSize.Y = BlockSize.Times(g.world.Obstacles.Size().Y)

	ebiten.SetWindowSize(g.screenSize.X.ToInt(), g.screenSize.Y.ToInt())
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(100, 100)

	g.folderWatcher.Folder = "data"
	g.loadGuiData()
	//g.imgEnemy.Fill(intToCol(3))

	// font
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
