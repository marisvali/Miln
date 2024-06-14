package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image"
	"image/color"
	_ "image/png"
	. "playful-patterns.com/miln/gamelib"
	. "playful-patterns.com/miln/world"
	"slices"
)

var PlayerCooldown Int = I(15)
var BlockSize Int = I(80)

type Gui struct {
	defaultFont     font.Face
	imgGround       *ebiten.Image
	imgTree         *ebiten.Image
	imgPlayer       *ebiten.Image
	imgPlayerHealth *ebiten.Image
	imgEnemy        *ebiten.Image
	imgEnemyHealth  *ebiten.Image
	imgTileOverlay  *ebiten.Image
	imgBeam         *ebiten.Image
	imgShadow       *ebiten.Image
	world           World
	frameIdx        Int
	folderWatcher   FolderWatcher
	recording       bool
	recordingFile   string
	allInputs       []PlayerInput
	state           GameState
	textHeight      Int
}

type GameState int64

const (
	GameOngoing GameState = iota
	GamePaused
	GameWon
	GameLost
	Playback
)

func (g *Gui) UpdateGameOngoing() {
	var justPressedKeys []ebiten.Key
	justPressedKeys = inpututil.AppendJustPressedKeys(justPressedKeys)
	if slices.Contains(justPressedKeys, ebiten.KeyEscape) {
		g.state = GamePaused
		return
	}
	if slices.Contains(justPressedKeys, ebiten.KeyR) {
		g.world = World{}
		g.world.Initialize()
	}
	if g.world.Enemy.Health.Leq(ZERO) {
		g.state = GameWon
		return
	}
	if g.world.Player.Health.Leq(ZERO) {
		g.state = GameLost
		return
	}

	x, y := ebiten.CursorPosition()
	mousePt := IPt(x, y).DivBy(BlockSize)
	if mousePt.X.Geq(g.world.Obstacles.Size().X) {
		mousePt.X = g.world.Obstacles.Size().X.Minus(ONE)
	}
	if mousePt.Y.Geq(g.world.Obstacles.Size().Y) {
		mousePt.Y = g.world.Obstacles.Size().Y.Minus(ONE)
	}

	var input PlayerInput
	input.Move = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
	input.MovePt = mousePt
	input.Shoot = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2)
	input.ShootPt = mousePt

	if g.recording {
		g.allInputs = append(g.allInputs, input)
		SerializeInputs(g.allInputs, g.recordingFile)
	} else {
		if idx := g.frameIdx.ToInt(); idx < len(g.allInputs) {
			input = g.allInputs[idx]
		}
	}

	g.world.Step(&input)

	if g.folderWatcher.FolderContentsChanged() {
		g.loadGuiData()
	}

	g.frameIdx.Inc()
}

func (g *Gui) UpdateGamePaused() {
	g.world.Player.TimeoutIdx = ZERO
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		g.state = GameOngoing
		g.UpdateGameOngoing()
	}
}

func (g *Gui) UpdateGameWon() {
	var justPressedKeys []ebiten.Key
	justPressedKeys = inpututil.AppendJustPressedKeys(justPressedKeys)
	if slices.Contains(justPressedKeys, ebiten.KeyR) {
		g.state = GameOngoing
		g.UpdateGameOngoing()
	}
}

func (g *Gui) UpdateGameLost() {
	var justPressedKeys []ebiten.Key
	justPressedKeys = inpututil.AppendJustPressedKeys(justPressedKeys)
	if slices.Contains(justPressedKeys, ebiten.KeyR) {
		g.state = GameOngoing
		g.UpdateGameOngoing()
	}
}

func (g *Gui) UpdatePlayback() {

}

func (g *Gui) Update() error {
	if g.state == GameOngoing {
		g.UpdateGameOngoing()
	} else if g.state == GamePaused {
		g.UpdateGamePaused()
	} else if g.state == GameWon {
		g.UpdateGameWon()
	} else if g.state == GameLost {
		g.UpdateGameLost()
	} else if g.state == Playback {
		g.UpdatePlayback()
	}
	return nil
}

func (g *Gui) LineObstaclesIntersection(l Line) (bool, Pt) {
	rows := g.world.Obstacles.Size().Y
	cols := g.world.Obstacles.Size().X
	ipts := []Pt{}
	var pt Pt
	for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
		for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
			if !g.world.Obstacles.Get(pt).IsZero() {
				s := Square{g.TileToScreen(pt), BlockSize.Times(I(90)).DivBy(I(100))}
				if intersects, ipt := LineSquareIntersection(l, s); intersects {
					ipts = append(ipts, ipt)
				}
			}
		}
	}

	return GetClosestPoint(ipts, l.Start)
}

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos Pt) {
	margin := float64(1)
	pos = pos.Times(BlockSize)
	x := pos.X.ToFloat64()
	y := pos.Y.ToFloat64()
	tileSize := BlockSize.ToFloat64() - 2*margin
	DrawSprite(screen, img, x+margin, y+margin, tileSize, tileSize)
}

func (g *Gui) TileToScreen(pos Pt) Pt {
	half := BlockSize.DivBy(TWO)
	return pos.Times(BlockSize).Plus(Pt{half, half})
}

func (g *Gui) WorldToGuiPos(pt Pt) Pt {
	return pt.Times(BlockSize).DivBy(g.world.BlockSize)
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
	g.DrawPlayer(screen, g.world.Player)
	//g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)

	// Draw enemy.
	g.DrawEnemy(screen, g.world.Enemy)

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

	// Draw instructional text.
	//g.DrawSprite2(g.textBackground, 0,
	//	float64(screen.Bounds().Dy())-textHeight-float64(g.data.PlaybackBarHeight),
	//	float64(screen.Bounds().Dx()),
	//	textHeight)
	var message1 string
	var message2 string
	if g.state == GameOngoing {
		message1 = "Kill everyone!"
		message2 = "left click - move, right click - shoot, R - restart, ESC - pause"
	} else if g.state == GamePaused {
		message1 = "Paused. Kill everyone!"
		message2 = "left click - move, right click - shoot, R - restart"
	} else if g.state == GameWon {
		message1 = "You won, congratulations!"
		message2 = "R - restart"
	} else if g.state == GameLost {
		message1 = "You lost."
		message2 = "R - restart"
	} else if g.state == Playback {
		//message = fmt.Sprintf("Playing back frame %d / %d", g.frameIdx, len(g.playerInputs))
	} else {
		Check(fmt.Errorf("unhandled game state: %d", g.state))
	}

	//textSize1 := text.BoundString(g.defaultFont, message1)
	//textSize2 := text.BoundString(g.defaultFont, message2)
	//
	//g.textHeight.
	//	textSize1.Dy() + textSize2.Dy()
	//actualTextHeight := I(textSize.Min.Y).Abs()

	//var r Rectangle
	//dx := I(screen.Bounds().Dx())
	//dy := I(screen.Bounds().Dy())
	//r.Corner1 = Pt{ZERO, dy.Minus(g.textHeight)}
	//r.Corner2 = Pt{dx, dy}
	//g.DrawFilledRect(screen, r, HexToColor(0x00FF00))
	//
	//var r2 Rectangle
	//
	//r2.Corner1 = Pt{ZERO, dy.Minus(actualTextHeight)}
	//r2.Corner2 = Pt{dx, dy}
	//g.DrawFilledRect(screen, r2, HexToColor(0x0000FF))

	//offsetX := (screen.Bounds().Dx() - textSize.Dx()) / 2
	//offsetY := g.textHeight.Minus(I(textSize.Dy())).DivBy(TWO).ToInt()
	//
	//textX := screen.Bounds().Min.X + offsetX
	//textY := screen.Bounds().Max.Y - offsetY - textSize.Max.Y
	//text.Draw(screen, message, g.defaultFont, textX, textY, HexToColor(0xFF0000))

	var r image.Rectangle
	r.Min = image.Point{screen.Bounds().Min.X, screen.Bounds().Max.Y - g.textHeight.ToInt()}
	r.Max = image.Point{screen.Bounds().Max.X, screen.Bounds().Max.Y - g.textHeight.ToInt()/2}
	textBox1 := screen.SubImage(r).(*ebiten.Image)
	g.DrawText(textBox1, message1, true, HexToColor(0xFF0000))

	r.Min = image.Point{screen.Bounds().Min.X, screen.Bounds().Max.Y - g.textHeight.ToInt()/2}
	r.Max = image.Point{screen.Bounds().Max.X, screen.Bounds().Max.Y}
	textBox2 := screen.SubImage(r).(*ebiten.Image)
	g.DrawText(textBox2, message2, true, HexToColor(0xFF0000))

	// Output TPS (ticks per second, which is like frames per second).
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) DrawText(screen *ebiten.Image, message string, centerX bool, color color.Color) {
	textSize := text.BoundString(g.defaultFont, message)
	var offsetX int
	if centerX {
		offsetX = (screen.Bounds().Dx() - textSize.Dx()) / 2
	} else {
		offsetX = 0
	}
	offsetY := (screen.Bounds().Dy() - textSize.Dy()) / 2
	textX := screen.Bounds().Min.X + offsetX
	textY := screen.Bounds().Max.Y - offsetY - textSize.Max.Y
	text.Draw(screen, message, g.defaultFont, textX, textY, color)
}

func (g *Gui) DrawEnemy(screen *ebiten.Image, e Enemy) {
	g.DrawTile(screen, g.imgEnemy, e.Pos)
	g.DrawHealth(screen, g.imgEnemyHealth, e.MaxHealth, e.Health, e.Pos)
}

func (g *Gui) DrawPlayer(screen *ebiten.Image, p Player) {
	mask := ebiten.NewImageFromImage(g.imgPlayer)
	{
		percent := p.TimeoutIdx.Times(I(100)).DivBy(PlayerCooldown)
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
		lineWidth := p.TimeoutIdx.Times(totalWidth).DivBy(PlayerCooldown)
		l := Line{IPt(0, 0), Pt{lineWidth, I(0)}}
		DrawLine(mask, l, color.RGBA{0, 0, 0, 255})
	}
	g.DrawTile(screen, g.imgPlayer, p.Pos)
	g.DrawTile(screen, mask, p.Pos)
	g.DrawHealth(screen, g.imgPlayerHealth, p.MaxHealth, p.Health, p.Pos)
}

func (g *Gui) DrawHealth(screen *ebiten.Image, imgHealth *ebiten.Image, maxHealth Int, currentHealth Int, tilePos Pt) {
	g.imgTileOverlay.Clear()
	totalWidth := g.imgTileOverlay.Bounds().Dx()
	width := I(totalWidth).Times(currentHealth).DivBy(maxHealth)
	DrawSprite(g.imgTileOverlay, imgHealth, 0, 0, width.ToFloat64(), float64(g.imgEnemyHealth.Bounds().Dy()))
	g.DrawTile(screen, g.imgTileOverlay, tilePos)
}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Gui) DrawFilledSquare(screen *ebiten.Image, s Square, col color.Color) {
	img := ebiten.NewImage(s.Size.ToInt(), s.Size.ToInt())
	img.Fill(col)

	op := &ebiten.DrawImageOptions{}

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

	x := s.Center.X.Minus(s.Size.DivBy(TWO)).ToFloat64()
	y := s.Center.Y.Minus(s.Size.DivBy(TWO)).ToFloat64()
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func (g *Gui) DrawFilledRect(screen *ebiten.Image, r Rectangle, col color.Color) {
	img := ebiten.NewImage(r.Width().ToInt(), r.Height().ToInt())
	img.Fill(col)

	op := &ebiten.DrawImageOptions{}

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

	op.GeoM.Translate(r.Min().X.ToFloat64(), r.Min().Y.ToFloat64())
	screen.DrawImage(img, op)
}

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
		g.imgPlayerHealth = LoadImage("data/player-health.png")
		g.imgEnemy = LoadImage("data/enemy.png")
		g.imgEnemyHealth = LoadImage("data/enemy-health.png")
		g.imgBeam = LoadImage("data/beam.png")
		g.imgShadow = LoadImage("data/shadow.png")
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true
}

func main() {
	var g Gui
	g.world.Initialize()
	g.recording = true
	g.textHeight = I(75)
	if g.recording {
		g.recordingFile = GetNewRecordingFile()
	} else {
		g.recordingFile = GetLatestRecordingFile()
		g.allInputs = DeserializeInputs(g.recordingFile)
	}

	windowSize := g.world.Obstacles.Size().Times(BlockSize)
	windowSize.Y.Add(g.textHeight)
	ebiten.SetWindowSize(windowSize.X.ToInt(), windowSize.Y.ToInt())
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(100, 100)

	g.folderWatcher.Folder = "data"
	g.loadGuiData()
	g.imgTileOverlay = ebiten.NewImage(BlockSize.ToInt(), BlockSize.ToInt())
	g.state = GameOngoing

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
