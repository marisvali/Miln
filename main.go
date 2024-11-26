package main

import (
	"embed"
	"fmt"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	. "github.com/marisvali/miln/ai"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"slices"
	"strconv"
)

//go:embed data/*
var embeddedFiles embed.FS

type GuiData struct {
	BlockSize int
}

type Gui struct {
	GuiData
	defaultFont        font.Face
	imgGround          *ebiten.Image
	imgTree            *ebiten.Image
	imgPlayer          *ebiten.Image
	imgPlayerHealth    *ebiten.Image
	imgGremlin         *ebiten.Image
	imgGremlinMask     *ebiten.Image
	imgHound           *ebiten.Image
	imgHoundMask       *ebiten.Image
	imgUltraHound      *ebiten.Image
	imgUltraHoundMask  *ebiten.Image
	imgPillar          *ebiten.Image
	imgPillarMask      *ebiten.Image
	imgQuestion        *ebiten.Image
	imgQuestionMask    *ebiten.Image
	imgKing            *ebiten.Image
	imgKingMask        *ebiten.Image
	imgEnemyHealth     *ebiten.Image
	imgTileOverlay     *ebiten.Image
	imgBeam            *ebiten.Image
	imgShadow          *ebiten.Image
	imgTextBackground  *ebiten.Image
	imgTextColor       *ebiten.Image
	imgAmmo            *ebiten.Image
	imgSpawnPortal     *ebiten.Image
	imgPlayerHitEffect *ebiten.Image
	imgKey             []*ebiten.Image
	world              World
	worldAtStart       World
	frameIdx           Int
	folderWatcher1     FolderWatcher
	folderWatcher2     FolderWatcher
	recording          bool
	recordingFile      string
	state              GameState
	textHeight         Int
	guiMargin          Int
	useEmbedded        bool
	buttonRegionWidth  Int
	buttonPause        Rectangle
	buttonNewLevel     Rectangle
	buttonRestartLevel Rectangle
	justPressedKeys    []ebiten.Key // keys pressed in this frame
	mousePt            Pt           // mouse position in this frame
	playerHitEffectIdx Int
	playthrough        Playthrough
	uploadDataChannel  chan uploadData
	username           string
	ai                 AI
	// db                 *sql.DB
}

type uploadData struct {
	user        string
	version     int64
	id          uuid.UUID
	playthrough []byte
}

type GameState int64

const (
	GameOngoing GameState = iota
	GamePaused
	GameWon
	GameLost
	Playback
)

func (g *Gui) JustPressed(key ebiten.Key) bool {
	return slices.Contains(g.justPressedKeys, key)
}

func (g *Gui) JustClicked(button Rectangle) bool {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		return false
	}
	return button.ContainsPt(g.mousePt)
}

func (g *Gui) UserRequestedPause() bool {
	return g.JustPressed(ebiten.KeyEscape) || g.JustClicked(g.buttonPause)
}

func (g *Gui) UserRequestedNewLevel() bool {
	return g.JustPressed(ebiten.KeyN) || g.JustClicked(g.buttonNewLevel)
}

func (g *Gui) UserRequestedRestartLevel() bool {
	return g.JustPressed(ebiten.KeyR) || g.JustClicked(g.buttonRestartLevel)
}

func (g *Gui) uploadCurrentWorld() {
	g.uploadDataChannel <- uploadData{g.username, Version, g.world.Id, g.world.SerializedPlaythrough()}
}

func (g *Gui) UpdateGameOngoing() {
	if g.UserRequestedPause() {
		g.state = GamePaused
		g.uploadCurrentWorld()
		return
	}
	if g.UserRequestedRestartLevel() {
		g.state = GamePaused
		g.uploadCurrentWorld()
		g.UpdateGamePaused()
		return
	}
	if g.UserRequestedNewLevel() {
		g.state = GamePaused
		g.uploadCurrentWorld()
		g.UpdateGamePaused()
		return
	}

	allEnemiesDead := true
	for _, enemy := range g.world.Enemies {
		if enemy.Alive() {
			allEnemiesDead = false
		}
	}
	for _, portal := range g.world.SpawnPortals {
		if portal.Active() {
			allEnemiesDead = false
		}
	}

	if allEnemiesDead {
		g.state = GameWon
		g.uploadCurrentWorld()
		return
	}
	if g.world.Player.Health.Leq(ZERO) {
		g.state = GameLost
		g.uploadCurrentWorld()
		return
	}

	var input PlayerInput
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		freePositions := g.world.Player.ComputeFreePositions(&g.world).ToSlice()
		tilePos, dist := g.ClosestTileToMouse(freePositions)
		closeEnough := dist.Lt(I(g.BlockSize).Times(I(300)).DivBy(I(100)))
		if closeEnough {
			input.Move = true
			input.MovePt = tilePos
		}
	}

	if g.world.Player.OnMap && inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) {
		attackablePositions := g.world.VulnerableEnemyPositions().ToSlice()
		tilePos, dist := g.ClosestTileToMouse(attackablePositions)
		closeEnough := dist.Lt(I(g.BlockSize).Times(I(300)).DivBy(I(100)))
		if closeEnough {
			input.Shoot = true
			input.ShootPt = tilePos
		}
	}

	if !g.recording {
		inputs := g.playthrough.History
		if idx := g.frameIdx.ToInt(); idx < len(inputs) {
			input = inputs[idx]
		}
	}

	// input = g.ai.Step(&g.world)
	g.world.Step(input)

	if g.recording {
		if g.recordingFile != "" {
			WriteFile(g.recordingFile, g.world.SerializedPlaythrough())
		}
		if g.frameIdx.Mod(I(60)) == ZERO {
			g.uploadCurrentWorld()
		}
	}

	g.frameIdx.Inc()
}

func (g *Gui) UpdateGamePaused() {
	if g.UserRequestedNewLevel() {
		seed, targetDifficulty := GetNextLevel(g.username)
		g.world = NewWorld(seed, targetDifficulty)
		// g.world = NewWorld(RInt(I(0), I(10000000)), RInt(I(55), I(70)))
		// InitializeIdInDbSql(g.db, g.world.Id)
		InitializeIdInDbHttp(g.username, Version, g.world.Id)
		return
	}
	if g.UserRequestedRestartLevel() {
		g.world = NewWorld(g.world.Seed, g.world.TargetDifficulty)
		// InitializeIdInDbSql(g.db, g.world.Id)
		InitializeIdInDbHttp(g.username, Version, g.world.Id)
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) {
		g.state = GameOngoing
		g.UpdateGameOngoing()
		return
	}
}

func (g *Gui) UpdateGameWon() {
	if g.UserRequestedRestartLevel() {
		g.state = GamePaused
		g.UpdateGamePaused()
		return
	}
	if g.UserRequestedNewLevel() {
		g.state = GamePaused
		g.UpdateGamePaused()
		return
	}
}

func (g *Gui) UpdateGameLost() {
	if g.UserRequestedRestartLevel() {
		g.state = GameOngoing
		g.UpdateGamePaused()
		return
	}
	if g.UserRequestedNewLevel() {
		g.state = GameOngoing
		g.UpdateGamePaused()
		return
	}
}

func (g *Gui) UpdatePlayback() {

}

func (g *Gui) Update() error {
	// One-time initialization. This needs to happen here because I need to
	// operate on ebiten images and it won't let me before I do RunGame.
	// TODO: find a better place for this code
	if g.imgGremlinMask == nil {
		g.imgGremlinMask = ComputeSpriteMask(g.imgGremlin)
		g.imgHoundMask = ComputeSpriteMask(g.imgHound)
		g.imgUltraHoundMask = ComputeSpriteMask(g.imgUltraHound)
		g.imgPillarMask = ComputeSpriteMask(g.imgPillar)
		g.imgQuestionMask = ComputeSpriteMask(g.imgQuestion)
		g.imgKingMask = ComputeSpriteMask(g.imgKing)
	}

	// Updates common to all states.
	if g.folderWatcher1.FolderContentsChanged() {
		g.loadGuiData()
	}
	if g.folderWatcher2.FolderContentsChanged() {
		// Reload world, and rely on the fact that this is makes the new world
		// load the new parameters from world.json.
		g.world = NewWorld(g.world.Seed, g.world.TargetDifficulty)
		g.updateWindowSize()
	}

	// Get input once, so we don't need to get it every time we need it in
	// other functions.
	g.justPressedKeys = g.justPressedKeys[:0]
	g.justPressedKeys = inpututil.AppendJustPressedKeys(g.justPressedKeys)
	x, y := ebiten.CursorPosition()
	g.mousePt = IPt(x, y)

	if g.JustPressed(ebiten.KeyX) {
		return ebiten.Termination
	}

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
	obstaclePositions := g.world.Obstacles.ToSlice()
	ipts := []Pt{}
	for _, pt := range obstaclePositions {
		s := Square{g.TileToScreen(pt), I(g.BlockSize).Times(I(90)).DivBy(I(100))}
		if intersects, ipt := LineSquareIntersection(l, s); intersects {
			ipts = append(ipts, ipt)
		}
	}

	return GetClosestPoint(ipts, l.Start)
}

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos Pt) {
	margin := float64(1)
	pos = pos.Times(I(g.BlockSize))
	x := pos.X.ToFloat64()
	y := pos.Y.ToFloat64()
	tileSize := float64(g.BlockSize) - 2*margin
	DrawSprite(screen, img, x+margin, y+margin, tileSize, tileSize)
}

func (g *Gui) DrawTileAlpha(screen *ebiten.Image, img *ebiten.Image, pos Pt, alpha uint8) {
	margin := float64(1)
	pos = pos.Times(I(g.BlockSize))
	x := pos.X.ToFloat64()
	y := pos.Y.ToFloat64()
	tileSize := float64(g.BlockSize) - 2*margin
	DrawSpriteAlpha(screen, img, x+margin, y+margin, tileSize, tileSize, alpha)
}

func (g *Gui) TileToScreen(pos Pt) Pt {
	half := I(g.BlockSize).DivBy(TWO)
	return pos.Times(I(g.BlockSize)).Plus(Pt{half, half}).Plus(Pt{g.guiMargin, g.guiMargin})
}

func (g *Gui) TilesToScreen(ipt []Pt) (opt []Pt) {
	for _, pt := range ipt {
		opt = append(opt, g.TileToScreen(pt))
	}
	return
}

func (g *Gui) ClosestTileToMouse(tiles []Pt) (tile Pt, dist Int) {
	opt := []Pt{}
	for _, pt := range tiles {
		opt = append(opt, g.TileToScreen(pt))
	}
	_, closestPt := GetClosestPoint(opt, g.mousePt)
	tile = g.ScreenToTile(closestPt)
	dist = closestPt.To(g.mousePt).Len()
	return
}

func (g *Gui) TileToPlayRegion(pos Pt) Pt {
	half := I(g.BlockSize).DivBy(TWO)
	return pos.Times(I(g.BlockSize)).Plus(Pt{half, half})
}

func (g *Gui) ScreenToTile(pos Pt) Pt {
	return pos.Minus(Pt{g.guiMargin, g.guiMargin}).DivBy(I(g.BlockSize))
}

func (g *Gui) WorldToGuiPos(pt Pt) Pt {
	return pt.Times(I(g.BlockSize)).DivBy(g.world.BlockSize).Plus(Pt{g.guiMargin, g.guiMargin})
}

func (g *Gui) WorldToPlayRegion(pt Pt) Pt {
	return pt.Times(I(g.BlockSize)).DivBy(g.world.BlockSize)
}

func (g *Gui) DrawPlayRegion(screen *ebiten.Image) {
	// Draw ground and trees.
	rows := g.world.Obstacles.Size().Y
	cols := g.world.Obstacles.Size().X
	var pt Pt
	for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
		for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
			g.DrawTile(screen, g.imgGround, pt)
			if g.world.Obstacles.At(pt) {
				g.DrawTile(screen, g.imgTree, pt)
			}
		}
	}

	// Draw portals.
	for i := range g.world.SpawnPortals {
		p := &g.world.SpawnPortals[i]
		g.DrawTile(screen, g.imgSpawnPortal, p.Pos)
	}

	// Draw ammo.
	for _, ammo := range g.world.Ammos {
		g.DrawTile(screen, g.imgAmmo, ammo.Pos)
	}

	// Draw keys.
	for _, key := range g.world.Keys {
		g.DrawTile(screen, g.imgKey[key.Type.ToInt()], key.Pos)
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
	if g.world.Player.OnMap {
		for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
			for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
				if g.world.AttackableTiles.At(pt) {
					g.DrawTile(screen, g.imgShadow, pt)
				}
			}
		}
	}

	// Draw hit effect.
	p := &g.world.Player
	if p.CooldownAfterGettingHitIdx.IsPositive() {
		i := p.CooldownAfterGettingHitIdx
		t := p.CooldownAfterGettingHit
		alpha := uint8(i.Times(I(100)).DivBy(t).ToInt()) + 30
		DrawSpriteAlpha(screen, g.imgPlayerHitEffect, 0, 0, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()), alpha)
	}
}

func (g *Gui) Draw(screen *ebiten.Image) {
	// Draw background.
	// percent starts from 100 and goes down to 0
	// percent := g.world.Player.TimeoutIdx.Times(I(100)).DivBy(PlayerCooldown)

	// gray needs to be at 80 when percent is at 0 and 0 when percent is at 100.
	// var gray Int
	// if percent.Gt(ZERO) {
	//	gray = (I(100).Minus(percent)).Times(I(30)).DivBy(I(100))
	// } else {
	//	gray = I(50)
	// }
	// v := uint8(gray.ToInt())
	// screen.Fill(color.RGBA{v, v, v, 255})
	screen.Fill(color.RGBA{0, 0, 0, 255})

	{
		upperLeft := Pt{g.guiMargin, g.guiMargin}
		playSize := g.world.Obstacles.Size().Times(I(g.BlockSize))
		lowerRight := upperLeft.Plus(playSize)
		playRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawPlayRegion(playRegion)
	}

	// buttonRegionX := I(screen.Bounds().Dx()).Minus(g.buttonRegionWidth)
	screenSize := IPt(screen.Bounds().Dx(), screen.Bounds().Dy())
	{
		upperLeft := Pt{ZERO, screenSize.Y.Minus(g.textHeight)}
		// lowerRight := upperLeft.Plus(Pt{buttonRegionX, g.textHeight.DivBy(TWO)})
		lowerRight := Pt{screenSize.X, screenSize.Y.Minus(g.textHeight.DivBy(TWO))}
		textRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		textRegion.Fill(color.RGBA{215, 215, 15, 255})
		g.DrawInstructionalText(textRegion)
	}

	{
		// upperLeft := Pt{buttonRegionX, I(screen.Bounds().Dy()).Minus(g.textHeight)}
		// lowerRight := upperLeft.Plus(Pt{I(screen.Bounds().Dx()), g.textHeight})
		upperLeft := Pt{ZERO, screenSize.Y.Minus(g.textHeight.DivBy(TWO))}
		lowerRight := Pt{screenSize.X, screenSize.Y}
		buttonRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		buttonRegion.Fill(color.RGBA{5, 215, 215, 255})
		g.DrawButtons(buttonRegion)
	}

	// Output TPS (ticks per second, which is like frames per second).
	ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) DrawButtons(screen *ebiten.Image) {
	height := I(screen.Bounds().Dy())
	buttonWidth := I(200)

	buttonCols := []color.RGBA{{35, 115, 115, 255}, {65, 215, 115, 255}, {225, 115, 215, 255}}

	buttons := []*ebiten.Image{}
	for i := I(0); i.Lt(I(3)); i.Inc() {
		upperLeft := Pt{buttonWidth.Times(i), ZERO}
		lowerRight := Pt{buttonWidth.Times(i.Plus(ONE)), height}
		button := SubImage(screen, Rectangle{upperLeft, lowerRight})
		button.Fill(buttonCols[i.ToInt()])
		buttons = append(buttons, button)
	}

	textCol := color.RGBA{0, 0, 0, 255}
	g.DrawText(buttons[0], "[ESC] Pause", true, textCol)
	g.DrawText(buttons[1], "[R] Restart level", true, textCol)
	g.DrawText(buttons[2], "[N] New level", true, textCol)

	// Remember the regions so that Update() can react when they're clicked.
	g.buttonPause = FromImageRectangle(buttons[0].Bounds())
	g.buttonRestartLevel = FromImageRectangle(buttons[1].Bounds())
	g.buttonNewLevel = FromImageRectangle(buttons[2].Bounds())
}

func (g *Gui) DrawInstructionalText(screen *ebiten.Image) {
	var message string
	if g.state == GameOngoing {
		message = "Kill everyone! left click - move, right click - shoot"
	} else if g.state == GamePaused {
		message = "Paused. Kill everyone! left click - move, right click - shoot"
	} else if g.state == GameWon {
		message = "You won, congratulations!"
	} else if g.state == GameLost {
		message = "You lost."
	} else if g.state == Playback {
		// message = fmt.Sprintf("Playing back frame %d / %d", g.frameIdx, len(g.playerInputs))
	} else {
		Check(fmt.Errorf("unhandled game state: %d", g.state))
	}

	DrawSprite(screen, g.imgTextBackground, 0, 0,
		float64(screen.Bounds().Dx()),
		float64(screen.Bounds().Dy()))

	var r image.Rectangle
	r.Min = screen.Bounds().Min
	r.Max = image.Point{screen.Bounds().Max.X, r.Min.Y + screen.Bounds().Dy()}
	textBox := screen.SubImage(r).(*ebiten.Image)
	g.DrawText(textBox, message, true, g.imgTextColor.At(0, 0))
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
	var img *ebiten.Image
	var imgMask *ebiten.Image
	drawHealth := true
	switch e.(type) {
	case *Gremlin:
		img = g.imgGremlin
		imgMask = g.imgGremlinMask
	case *Hound:
		img = g.imgHound
		imgMask = g.imgHoundMask
	case *UltraHound:
		img = g.imgUltraHound
		drawHealth = g.world.Player.HitPermissions.CanHitUltraHound
		imgMask = g.imgUltraHoundMask
	case *Pillar:
		img = g.imgPillar
		imgMask = g.imgPillarMask
	case *King:
		img = g.imgKing
		imgMask = g.imgKingMask
	case *Question:
		img = g.imgQuestion
		imgMask = g.imgQuestionMask
	}

	g.DrawTile(screen, img, e.Pos())

	alpha := ZERO
	if e.FreezeCooldown().IsPositive() {
		percent := e.FreezeCooldownIdx().Times(I(100)).DivBy(e.FreezeCooldown())
		if percent.Gt(ZERO) {
			alpha = (percent.Plus(I(100))).Times(I(255)).DivBy(I(200))
		}
	}

	g.DrawTileAlpha(screen, imgMask, e.Pos(), uint8(alpha.ToInt()))
	if drawHealth {
		g.DrawHealth(screen, g.imgEnemyHealth, e.Health(), e.Pos())
	}
}

func (g *Gui) DrawPlayer(screen *ebiten.Image, p Player) {
	if !p.OnMap {
		return
	}
	mask := ebiten.NewImageFromImage(g.imgPlayer)
	// Draw mask of move cooldown.
	// {
	//	percent := p.MoveCooldownIdx.Times(I(100)).DivBy(p.MoveCooldown)
	//	var alpha Int
	//	if percent.Gt(ZERO) {
	//		alpha = (percent.Plus(I(100))).Times(I(255)).DivBy(I(200))
	//	} else {
	//		alpha = ZERO
	//	}
	//
	//	sz := mask.Bounds().Size()
	//	for y := 0; y < sz.Y; y++ {
	//		for x := 0; x < sz.X; x++ {
	//			_, _, _, a := mask.At(x, y).RGBA()
	//			if a > 0 {
	//				mask.Set(x, y, color.RGBA{0, 0, 0, uint8(alpha.ToInt())})
	//			}
	//		}
	//	}
	//
	//	totalWidth := I(mask.Bounds().Size().X)
	//	//lineWidth := p.AmmoCount.Times(totalWidth).DivBy(I(3))
	//	lineWidth := percent.Times(totalWidth).DivBy(I(100))
	//	l := Line{IPt(0, mask.Bounds().Dy()), Pt{lineWidth, I(mask.Bounds().Dy())}}
	//	DrawLine(mask, l, color.RGBA{0, 0, 0, 255})
	// }
	g.DrawTile(screen, g.imgPlayer, p.Pos)
	g.DrawTile(screen, mask, p.Pos)
	g.DrawHealth(screen, g.imgPlayerHealth, p.Health, p.Pos)
}

func (g *Gui) DrawHealth(screen *ebiten.Image, imgHealth *ebiten.Image, currentHealth Int, tilePos Pt) {
	g.imgTileOverlay.Clear()
	blockSize := float64(g.imgEnemyHealth.Bounds().Dy())
	for i := I(0); i.Lt(currentHealth); i.Inc() {
		DrawSprite(g.imgTileOverlay, imgHealth, blockSize*i.ToFloat64()*1.3, 0, blockSize, blockSize)
	}
	g.DrawTile(screen, g.imgTileOverlay, tilePos)
}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Gui) LoadImage(filename string) *ebiten.Image {
	if g.useEmbedded {
		return LoadImageEmbedded(filename, &embeddedFiles)
	} else {
		return LoadImage(filename)
	}
}

func (g *Gui) loadGuiData() {
	// Read from the disk over and over until a full read is possible.
	// This repetition is meant to avoid crashes due to reading files
	// while they are still being written.
	// It's a hack but possibly a quick and very useful one.
	CheckCrashes = false
	for {
		CheckFailed = nil
		LoadJSON("data/gui/gui.json", &g.GuiData)
		g.imgGround = g.LoadImage("data/gui/ground.png")
		g.imgTree = g.LoadImage("data/gui/tree.png")
		g.imgPlayer = g.LoadImage("data/gui/player.png")
		g.imgPlayerHealth = g.LoadImage("data/gui/player-health.png")
		// g.imgEnemy = append(g.imgEnemy, g.LoadImage("data/enemy.png"))
		g.imgGremlin = g.LoadImage("data/gui/enemy2.png")
		g.imgPillar = g.LoadImage("data/gui/enemy3.png")
		g.imgHound = g.LoadImage("data/gui/enemy4.png")
		g.imgUltraHound = g.LoadImage("data/gui/ultra-hound.png")
		g.imgQuestion = g.LoadImage("data/gui/enemy5.png")
		g.imgKing = g.LoadImage("data/gui/enemy6.png")
		g.imgEnemyHealth = g.LoadImage("data/gui/enemy-health.png")
		g.imgBeam = g.LoadImage("data/gui/beam.png")
		g.imgShadow = g.LoadImage("data/gui/shadow.png")
		g.imgTextBackground = g.LoadImage("data/gui/text-background.png")
		g.imgTextColor = g.LoadImage("data/gui/text-color.png")
		g.imgAmmo = g.LoadImage("data/gui/ammo.png")
		g.imgSpawnPortal = g.LoadImage("data/gui/spawn-portal.png")
		g.imgPlayerHitEffect = g.LoadImage("data/gui/player-hit-effect.png")
		g.imgKey = append(g.imgKey, g.LoadImage("data/gui/key1.png"))
		g.imgKey = append(g.imgKey, g.LoadImage("data/gui/key2.png"))
		g.imgKey = append(g.imgKey, g.LoadImage("data/gui/key3.png"))
		g.imgKey = append(g.imgKey, g.LoadImage("data/gui/key4.png"))
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true

	g.updateWindowSize()
}

func (g *Gui) updateWindowSize() {
	playSize := g.world.Obstacles.Size().Times(I(g.BlockSize))
	windowSize := playSize
	windowSize.Add(Pt{g.guiMargin.Times(TWO), g.guiMargin})
	windowSize.Y.Add(g.textHeight)
	ebiten.SetWindowSize(windowSize.X.ToInt(), windowSize.Y.ToInt())
	ebiten.SetWindowTitle("Miln")
}

func UploadPlaythroughs(ch chan uploadData) {
	for {
		// Receive a playthrough from the channel.
		// Blocks until a playthrough is received.
		data := <-ch

		// Upload the data.
		UploadDataToDbHttp(data.user, data.version, data.id, data.playthrough)
	}
}

func GetNextLevel(user string) (seed Int, targetDifficulty Int) {
	// Get index and increment it.
	filenameIndex := user + "-index.txt"
	index, _ := strconv.Atoi(string(ReadFile(filenameIndex)))
	WriteFile(filenameIndex, []byte(strconv.Itoa(index+1)))

	// Get the seed and target difficulty for the current index.
	filenameLevels := user + "-levels.txt"
	lines := SplitInLines(ReadFile(filenameLevels))
	line := lines[index%len(lines)]
	var token1, token2 int
	fmt.Sscanf(line, "%d %d", &token1, &token2)
	seed = I(token1)
	targetDifficulty = I(token2)
	return
}

func main() {
	ebiten.SetWindowPosition(100, 100)

	var g Gui
	g.username = getUsername()
	g.uploadDataChannel = make(chan uploadData)
	go UploadPlaythroughs(g.uploadDataChannel)
	// g.db = ConnectToDbSql()
	// g.world = NewWorld(RInt(I(0), I(10000000)))

	g.textHeight = I(75)
	g.guiMargin = I(50)
	g.buttonRegionWidth = I(200)
	g.recording = true
	if len(os.Args) == 2 {
		g.recording = false
		g.playthrough = DeserializePlaythrough(ReadFile(os.Args[1]))
		g.world = NewWorld(g.playthrough.Seed, g.playthrough.TargetDifficulty)
		g.state = GameOngoing
	} else if g.recording {
		// g.recordingFile = GetNewRecordingFile()
		seed, targetDifficulty := GetNextLevel(g.username)
		g.world = NewWorld(seed, targetDifficulty)
		// g.world = NewWorld(RInt(I(0), I(1000000)))
		// InitializeIdInDbSql(g.db, g.world.Id)
		// UploadDataToDbSql(g.db, g.world.Id, g.world.SerializedPlaythrough())
		// InitializeIdInDbHttp(g.username, Version, g.world.Id)
		g.state = GamePaused
	} else {
		// g.recordingFile = GetLatestRecordingFile()
		// if g.recordingFile != "" {
		//	g.playthrough = DeserializePlaythrough(ReadFile(g.recordingFile))
		// }

		// id, err := uuid.Parse("dec49e01-bb13-4c63-b3e9-b5b9261dad67")
		// id, err := uuid.Parse("b02433de-bef5-476b-bbf1-7cf23fe8fcef")
		// Check(err)
		// db := ConnectToDbSql()
		// zippedPlaythrough := DownloadDataFromDbSql(db, id)
		// g.playthrough = DeserializePlaythrough(zippedPlaythrough)
		// g.playthrough = DeserializePlaythrough(ReadFile("world/playthroughs/20240714-120933.mln006"))
		g.world = NewWorldFromString(Level1())
		g.state = GameOngoing
	}

	g.useEmbedded = !FileExists("data")
	if !g.useEmbedded {
		g.folderWatcher1.Folder = "data/gui"
		g.folderWatcher2.Folder = "data/world"
	}
	g.loadGuiData()

	g.imgTileOverlay = ebiten.NewImage(g.BlockSize, g.BlockSize)

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
