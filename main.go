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
	"math"
	. "playful-patterns.com/miln/gamelib"
	. "playful-patterns.com/miln/world"
	"slices"
)

var PlayerCooldown Int = I(40)
var BlockSize Int = I(80)

type Gui struct {
	defaultFont       font.Face
	imgGround         *ebiten.Image
	imgTree           *ebiten.Image
	imgPlayer         *ebiten.Image
	imgPlayerHealth   *ebiten.Image
	imgEnemy          *ebiten.Image
	imgEnemyHealth    *ebiten.Image
	imgTileOverlay    *ebiten.Image
	imgBeam           *ebiten.Image
	imgShadow         *ebiten.Image
	imgTextBackground *ebiten.Image
	imgTextColor      *ebiten.Image
	world             World
	frameIdx          Int
	folderWatcher     FolderWatcher
	recording         bool
	recordingFile     string
	allInputs         []PlayerInput
	state             GameState
	textHeight        Int
	guiMargin         Int
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
		g.state = GamePaused
		g.UpdateGamePaused()
		return
	}

	allEnemiesDead := true
	for _, enemy := range g.world.Enemies {
		if enemy.Health.IsPositive() {
			allEnemiesDead = false
		}
	}

	if allEnemiesDead {
		g.state = GameWon
		return
	}
	if g.world.Player.Health.Leq(ZERO) {
		g.state = GameLost
		return
	}

	x, y := ebiten.CursorPosition()
	var input PlayerInput
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		// Decide if this is a move attempt or a shoot attempt.
		// See what is the closest distance to an enemy, from the click point.
		iClosestEnemy := -1
		minDist := I(math.MaxInt64)
		for i, _ := range g.world.Enemies {
			enemyPos := g.TileToScreen(g.world.Enemies[i].Pos)
			dist := enemyPos.To(IPt(x, y)).Len()
			if dist.Lt(minDist) {
				minDist = dist
				iClosestEnemy = i
			}
		}

		// It's a shoot attempt if the click point was "close" to an enemy,
		// where I define "close" relative to the block size.
		shootAttempt := minDist.Lt(BlockSize.Times(I(120)).DivBy(I(100)))
		if shootAttempt {
			input.Shoot = true
			input.ShootPt = g.world.Enemies[iClosestEnemy].Pos
		} else {
			input.Move = true
			tilePos := g.ScreenToTile(IPt(x, y))
			if tilePos.X.IsNegative() {
				tilePos.X = ZERO
			}
			if tilePos.X.Geq(g.world.Obstacles.Size().X) {
				tilePos.X = g.world.Obstacles.Size().X.Minus(ONE)
			}
			if tilePos.Y.IsNegative() {
				tilePos.Y = ZERO
			}
			if tilePos.Y.Geq(g.world.Obstacles.Size().Y) {
				tilePos.Y = g.world.Obstacles.Size().Y.Minus(ONE)
			}
			input.MovePt = tilePos
		}
	}

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
	var justPressedKeys []ebiten.Key
	justPressedKeys = inpututil.AppendJustPressedKeys(justPressedKeys)
	if slices.Contains(justPressedKeys, ebiten.KeyR) {
		g.world = World{}
		g.world.Initialize()
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
	return pos.Times(BlockSize).Plus(Pt{half, half}).Plus(Pt{g.guiMargin, g.guiMargin})
}

func (g *Gui) TileToPlayRegion(pos Pt) Pt {
	half := BlockSize.DivBy(TWO)
	return pos.Times(BlockSize).Plus(Pt{half, half})
}

func (g *Gui) ScreenToTile(pos Pt) Pt {
	return pos.Minus(Pt{g.guiMargin, g.guiMargin}).DivBy(BlockSize)
}

func (g *Gui) WorldToGuiPos(pt Pt) Pt {
	return pt.Times(BlockSize).DivBy(g.world.BlockSize).Plus(Pt{g.guiMargin, g.guiMargin})
}

func (g *Gui) WorldToPlayRegion(pt Pt) Pt {
	return pt.Times(BlockSize).DivBy(g.world.BlockSize)
}

func (g *Gui) DrawPlayRegion(screen *ebiten.Image) {
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

	// Draw enemy.
	for _, enemy := range g.world.Enemies {
		g.DrawEnemy(screen, enemy)
	}

	// Draw beam.
	beamScreen := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	if g.world.Beam.Idx.Gt(ZERO) {
		beam := Line{g.TileToPlayRegion(g.world.Player.Pos), g.WorldToPlayRegion(g.world.Beam.End)}
		alpha := uint8(g.world.Beam.Idx.Times(I(255)).DivBy(g.world.BeamMax).ToInt())
		colr, colg, colb, _ := g.imgBeam.At(0, 0).RGBA()
		beamCol := color.RGBA{uint8(colr), uint8(colg), uint8(colb), alpha}
		DrawLine(beamScreen, beam, beamCol)
	}
	DrawSpriteXY(screen, beamScreen, 0, 0)

	// Mark attackable tiles.
	for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
		for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
			if g.world.AttackableTiles.Get(pt).Neq(ZERO) {
				g.DrawTile(screen, g.imgShadow, pt)
			}
		}
	}
}

func (g *Gui) Draw(screen *ebiten.Image) {
	// Draw background.
	// percent starts from 100 and goes down to 0
	//percent := g.world.Player.TimeoutIdx.Times(I(100)).DivBy(PlayerCooldown)

	// gray needs to be at 80 when percent is at 0 and 0 when percent is at 100.
	//var gray Int
	//if percent.Gt(ZERO) {
	//	gray = (I(100).Minus(percent)).Times(I(30)).DivBy(I(100))
	//} else {
	//	gray = I(50)
	//}
	//v := uint8(gray.ToInt())
	//screen.Fill(color.RGBA{v, v, v, 255})
	screen.Fill(color.RGBA{0, 0, 0, 255})

	{
		upperLeft := Pt{g.guiMargin, g.guiMargin}
		playSize := g.world.Obstacles.Size().Times(BlockSize)
		lowerRight := upperLeft.Plus(playSize)
		playRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawPlayRegion(playRegion)
	}

	{
		upperLeft := Pt{ZERO, I(screen.Bounds().Dy()).Minus(g.textHeight)}
		lowerRight := upperLeft.Plus(Pt{I(screen.Bounds().Dx()), g.textHeight})
		textRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawInstructionalText(textRegion)
	}

	// Output TPS (ticks per second, which is like frames per second).
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) DrawInstructionalText(screen *ebiten.Image) {
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

	DrawSprite(screen, g.imgTextBackground, 0, 0,
		float64(screen.Bounds().Dx()),
		float64(screen.Bounds().Dy()))

	var r image.Rectangle
	r.Min = screen.Bounds().Min
	r.Max = image.Point{screen.Bounds().Max.X, r.Min.Y + screen.Bounds().Dy()/2}
	textBox1 := screen.SubImage(r).(*ebiten.Image)
	g.DrawText(textBox1, message1, true, g.imgTextColor.At(0, 0))

	r.Min = image.Point{screen.Bounds().Min.X, screen.Bounds().Min.Y + screen.Bounds().Dy()/2}
	r.Max = screen.Bounds().Max
	textBox2 := screen.SubImage(r).(*ebiten.Image)
	g.DrawText(textBox2, message2, true, g.imgTextColor.At(0, 0))
}

func (g *Gui) DrawText(screen *ebiten.Image, message string, centerX bool, color color.Color) {
	// Remember that text there is an origin point for the text.
	// That origin point is kind of the lower-left corner of the bounds of the
	// text. Kind of. Read the BoundString docs to understand, particularly this
	// image:
	// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyphterms_2x.png
	// This means that if you do text.Draw at (x, y), most of the text will
	// appear above y, and a little bit under y. If you want all the pixels in
	// your text to be above y, you should do text.Draw at
	// (x, y - text.BoundString().Max.Y).
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
		l := Line{IPt(0, mask.Bounds().Dy()), Pt{lineWidth, I(mask.Bounds().Dy())}}
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
		g.imgTextBackground = LoadImage("data/text-background.png")
		g.imgTextColor = LoadImage("data/text-color.png")
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true
}

func main() {
	var g Gui
	g.world.Initialize()

	g.textHeight = I(75)
	g.guiMargin = I(50)
	g.recording = true
	if g.recording {
		g.recordingFile = GetNewRecordingFile()
	} else {
		g.recordingFile = GetLatestRecordingFile()
		g.allInputs = DeserializeInputs(g.recordingFile)
	}

	playSize := g.world.Obstacles.Size().Times(BlockSize)
	windowSize := playSize
	windowSize.Add(Pt{g.guiMargin.Times(TWO), g.guiMargin})
	windowSize.Y.Add(g.textHeight)
	ebiten.SetWindowSize(windowSize.X.ToInt(), windowSize.Y.ToInt())
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(100, 100)

	g.folderWatcher.Folder = "data"
	g.loadGuiData()
	g.imgTileOverlay = ebiten.NewImage(BlockSize.ToInt(), BlockSize.ToInt())
	g.state = GamePaused

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
