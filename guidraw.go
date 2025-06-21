package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"slices"
)

func (g *Gui) Draw(screen *ebiten.Image) {
	// Draw background.
	screen.Fill(Col(0, 0, 0, 255))

	playSize := IPt(NCols, NRows).Times(g.BlockSize)
	yPlayRegion := g.guiMargin
	var yInstructionalText, yButtons, yPlayback Int
	if !g.playbackExecution {
		yInstructionalText = yPlayRegion.Plus(playSize.Y)
		yButtons = yInstructionalText.Plus(g.textHeight)
		yPlayback = yButtons.Plus(g.textHeight)
	} else {
		yInstructionalText = yPlayRegion.Plus(playSize.Y)
		yButtons = yInstructionalText.Plus(g.textHeight)
		yPlayback = yButtons
	}

	{
		upperLeft := Pt{g.guiMargin, I(0)}
		lowerRight := Pt{g.guiMargin.Plus(playSize.X), yPlayRegion}
		playerHealthRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawPlayerHealth(playerHealthRegion)
	}

	{
		upperLeft := Pt{g.guiMargin, I(0)}
		lowerRight := Pt{g.guiMargin.Plus(playSize.X), yPlayRegion}
		playerHealthRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawCurrentLevel(playerHealthRegion)
	}

	{
		upperLeft := Pt{g.guiMargin, yPlayRegion}
		lowerRight := Pt{g.guiMargin.Plus(playSize.X), yInstructionalText}
		playRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawPlayRegion(playRegion)
	}

	screenSize := IPt(screen.Bounds().Dx(), screen.Bounds().Dy())
	{
		upperLeft := Pt{ZERO, yInstructionalText}
		lowerRight := Pt{screenSize.X, yButtons}
		textRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		textRegion.Fill(Col(215, 215, 15, 255))
		g.DrawInstructionalText(textRegion)
	}

	if !g.playbackExecution {
		// upperLeft := Pt{buttonRegionX, I(screen.Bounds().Dy()).Minus(g.textHeight)}
		// lowerRight := upperLeft.Plus(Pt{I(screen.Bounds().Dx()), g.textHeight})
		upperLeft := Pt{ZERO, yButtons}
		lowerRight := Pt{screenSize.X, yPlayback}
		buttonRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		buttonRegion.Fill(Col(5, 215, 215, 255))
		g.DrawButtons(buttonRegion)
	}

	if g.playbackExecution {
		// Draw playback bar.
		upperLeft := Pt{ZERO, yPlayback}
		lowerRight := Pt{screenSize.X, screenSize.Y}
		playbackBarRegion := SubImage(screen, Rectangle{upperLeft, lowerRight})
		g.DrawPlaybackBar(playbackBarRegion)

		// Draw virtual cursor.
		if g.DrawVirtualCursorDuringReplay {
			DrawSprite(screen, g.imgCursor,
				g.mousePt.X.ToFloat64(),
				g.mousePt.Y.ToFloat64(),
				20.0, 20.0)
		}
	}

	// Output TPS (ticks per second, which is like frames per second).
	// ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) DrawPlayRegion(screen *ebiten.Image) {
	// Draw ground and trees.
	rows := I(NRows)
	cols := I(NCols)
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
	if g.DrawSpawnPortal {
		for i := range g.world.SpawnPortals.N {
			p := &g.world.SpawnPortals.V[i]
			g.DrawTile(screen, g.imgSpawnPortal, p.Pos())
		}
	}

	// Draw ammo.
	for i := range g.world.Ammos.N {
		g.DrawTile(screen, g.imgAmmo, g.world.Ammos.V[i].Pos)
	}

	// Draw enemy.
	for i := range g.world.Enemies.N {
		g.DrawEnemy(screen, &g.world.Enemies.V[i])
	}

	// Draw all animations for world objects.
	for _, o := range g.visWorld.Objects {
		if o.Animation.Valid() {
			g.DrawTile(screen, o.Animation.Img(), o.Object.Pos())
		}
	}

	// Mark tiles that the player can't see.
	if g.world.Player.OnMap {
		for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
			for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
				if !g.world.VisibleTiles.At(pt) {
					g.DrawTile(screen, g.imgShadow, pt)
				}
			}
		}
	} else {
		for pt.Y = ZERO; pt.Y.Lt(rows); pt.Y.Inc() {
			for pt.X = ZERO; pt.X.Lt(cols); pt.X.Inc() {
				g.DrawTile(screen, g.imgShadow, pt)
			}
		}
	}

	// Draw beam.
	beamScreen := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	if g.world.Beam.Idx.Gt(ZERO) {
		beam := Line{g.TileToPlayRegion(g.world.Player.Pos()), g.WorldToPlayRegion(g.world.Beam.End)}
		alpha := uint8(g.world.Beam.Idx.Times(I(255)).DivBy(g.world.BeamMax).ToInt())
		colr, colg, colb, _ := g.imgBeam.At(0, 0).RGBA()
		beamCol := Col(uint8(colr), uint8(colg), uint8(colb), alpha)
		DrawLine(beamScreen, beam, beamCol)
	}
	DrawSpriteXY(screen, beamScreen, 0, 0)

	// Highlight attack
	attackOk, attackPos := g.GetAttackTarget()
	highlightedPositions := []Pt{}
	positionAlreadyHighlighted := slices.Contains(highlightedPositions, attackPos)
	if attackOk && g.HighlightAttack && !positionAlreadyHighlighted {
		g.DrawTile(screen, g.imgHighlightAttack, attackPos)
		highlightedPositions = append(highlightedPositions, attackPos)
	}

	// Highlight move ok
	moveOk, movePos := g.GetMoveTarget()
	positionAlreadyHighlighted = slices.Contains(highlightedPositions, movePos)
	if moveOk && g.HighlightMoveOk && !positionAlreadyHighlighted {
		g.DrawTile(screen, g.imgHighlightMoveOk, movePos)
		highlightedPositions = append(highlightedPositions, movePos)
	}

	// Highlight move not ok
	tilePos := g.ScreenToTile(g.mousePt)
	positionAlreadyHighlighted = slices.Contains(highlightedPositions, tilePos)
	if !moveOk && g.HighlightMoveNotOk && g.MouseCursorIsOverATile() &&
		!positionAlreadyHighlighted {
		g.DrawTile(screen, g.imgHighlightMoveNotOk, tilePos)
		highlightedPositions = append(highlightedPositions, tilePos)
	}

	// Draw all temporary animations.
	for _, o := range g.visWorld.Temporary {
		if o.Animation.Valid() {
			g.DrawTile(screen, o.Animation.Img(), g.ScreenToTile(o.ScreenPos))
		}
	}

	// Draw player.
	g.DrawPlayer(screen, g.world.Player)

	// Draw hit effect.
	p := &g.world.Player
	if p.CooldownAfterGettingHitIdx.IsPositive() {
		i := p.CooldownAfterGettingHitIdx
		t := p.CooldownAfterGettingHit
		alpha := uint8(i.Times(I(100)).DivBy(t).ToInt()) + 30
		DrawSpriteAlpha(screen, g.imgPlayerHitEffect, 0, 0, float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy()), alpha)
	}
}

func (g *Gui) DrawEnemy(screen *ebiten.Image, e *Hound) {
	if g.DrawEnemyHealth {
		g.DrawHealth(screen, g.imgEnemyHealth, e.Health(), e.Pos())
	}
}

func (g *Gui) DrawPlayer(screen *ebiten.Image, p Player) {
	if !p.OnMap {
		return
	}

	// Draw ammo as a circle around the player.
	g.imgTileOverlay.Clear()
	blockSize := float64(g.imgTileOverlay.Bounds().Dx()) / 4

	// Define the total number of positions on the circle
	var totalPositions = g.world.AmmoLimit.ToFloat64()

	// Define the center of the circle
	bounds := g.imgTileOverlay.Bounds()
	size := bounds.Max.Sub(bounds.Min)
	center := bounds.Min.Add(size.Div(2))
	var centerX = float64(center.X) - blockSize/2
	var centerY = float64(center.Y) - blockSize/2

	// Define the radius of the circle
	var radius = float64(size.X) / 2 * 85 / 100

	// Iterate through positions on the circle
	ammoCount := p.AmmoCount.ToInt()
	for i := 0; i < ammoCount; i++ {
		// Calculate the angle for the current position (in radians)
		angle := (2*math.Pi/totalPositions)*float64(i) - (math.Pi / 2)

		// Calculate the x and y coordinates
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)

		// Display sprite at the x, y coordinates
		DrawSprite(g.imgTileOverlay, g.imgPlayerAmmo, x, y, blockSize, blockSize)
	}
	g.DrawTile(screen, g.imgTileOverlay, p.Pos())
}

func (g *Gui) DrawHealth(screen *ebiten.Image, imgHealth *ebiten.Image, currentHealth Int, tilePos Pt) {
	g.imgTileOverlay.Clear()
	blockSize := float64(g.imgEnemyHealth.Bounds().Dy())
	for i := I(0); i.Lt(currentHealth); i.Inc() {
		DrawSprite(g.imgTileOverlay, imgHealth, blockSize*i.ToFloat64()*1.3, 0, blockSize, blockSize)
	}
	g.DrawTile(screen, g.imgTileOverlay, tilePos)
}

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos Pt) {
	margin := float64(1)
	pos = pos.Times(g.BlockSize)
	x := pos.X.ToFloat64()
	y := pos.Y.ToFloat64()
	tileSize := g.BlockSize.ToFloat64() - 2*margin
	DrawSprite(screen, img, x+margin, y+margin, tileSize, tileSize)
}

func (g *Gui) DrawTileAlpha(screen *ebiten.Image, img *ebiten.Image, pos Pt, alpha uint8) {
	margin := float64(1)
	pos = pos.Times(g.BlockSize)
	x := pos.X.ToFloat64()
	y := pos.Y.ToFloat64()
	tileSize := g.BlockSize.ToFloat64() - 2*margin
	DrawSpriteAlpha(screen, img, x+margin, y+margin, tileSize, tileSize, alpha)
}

func (g *Gui) DrawPlayerHealth(screen *ebiten.Image) {
	for idx := ZERO; idx.Lt(g.world.Player.Health); idx.Inc() {
		tileSize := float64(screen.Bounds().Dy())
		x := tileSize * idx.ToFloat64()
		y := float64(0)
		DrawSprite(screen, g.imgPlayerHealth, x, y, tileSize, tileSize)
	}
}

func (g *Gui) DrawCurrentLevel(screen *ebiten.Image) {
	if g.state == Playback {
		// Don't show current level during playback as the current level index
		// is not currently saved in the playthrough.
		return
	}

	if g.UsingFixedLevels() {
		currentLevel := g.CurrentFixedLevelIdx.ToInt() + 1
		if currentLevel <= len(g.fixedLevels) {
			if g.state == GameWon || g.state == GameLost {
				// We just won/lost a level and g.CurrentFixedLevelIdx was
				// increased by one, in order to save that to the database as
				// soon as possible. But the UI is still showing the message
				// for the previous level. So just fix this here with a quick
				// hack.
				currentLevel--
			}

			str := fmt.Sprintf("Level %d of %d.", currentLevel, len(g.fixedLevels))
			g.DrawText(screen, str, true, g.imgTextColor.At(0, 0))
		}
	}
}

func (g *Gui) DrawInstructionalText(screen *ebiten.Image) {
	DrawSprite(screen, g.imgTextBackground, 0, 0,
		float64(screen.Bounds().Dx()),
		float64(screen.Bounds().Dy()))

	var r image.Rectangle
	r.Min = screen.Bounds().Min
	r.Max = image.Point{screen.Bounds().Max.X, r.Min.Y + screen.Bounds().Dy()}
	textBox := screen.SubImage(r).(*ebiten.Image)
	g.DrawText(textBox, g.instructionalText, true, g.imgTextColor.At(0, 0))
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

func (g *Gui) DrawButtons(screen *ebiten.Image) {
	height := I(screen.Bounds().Dy())
	buttonWidth := I(200)

	buttonCols := []color.Color{Col(35, 115, 115, 255), Col(65, 215, 115, 255),
		Col(225, 115, 215, 255)}
	textCol := Col(0, 0, 0, 255)

	// Pause
	i := I(0)
	upperLeft := Pt{buttonWidth.Times(i), ZERO}
	lowerRight := Pt{buttonWidth.Times(i.Plus(ONE)), height}
	button := SubImage(screen, Rectangle{upperLeft, lowerRight})
	button.Fill(buttonCols[i.ToInt()])
	g.DrawText(button, "[ESC] Pause", true, textCol)
	g.buttonPause = FromImageRectangle(button.Bounds())

	if g.UsingFixedLevels() {
		if g.NextLevelRequestable() {
			// Next level
			i.Inc()
			upperLeft = Pt{buttonWidth.Times(i), ZERO}
			lowerRight = Pt{buttonWidth.Times(i.Plus(ONE)), height}
			button = SubImage(screen, Rectangle{upperLeft, lowerRight})
			button.Fill(buttonCols[i.ToInt()])
			g.DrawText(button, "[N] Next level", true, textCol)
			g.buttonNextLevel = FromImageRectangle(button.Bounds())
		}
	} else {
		// Restart level
		i.Inc()
		upperLeft = Pt{buttonWidth.Times(i), ZERO}
		lowerRight = Pt{buttonWidth.Times(i.Plus(ONE)), height}
		button = SubImage(screen, Rectangle{upperLeft, lowerRight})
		button.Fill(buttonCols[i.ToInt()])
		g.DrawText(button, "[R] Restart level", true, textCol)
		g.buttonRestartLevel = FromImageRectangle(button.Bounds())

		// New level
		i.Inc()
		upperLeft = Pt{buttonWidth.Times(i), ZERO}
		lowerRight = Pt{buttonWidth.Times(i.Plus(ONE)), height}
		button = SubImage(screen, Rectangle{upperLeft, lowerRight})
		button.Fill(buttonCols[i.ToInt()])
		g.DrawText(button, "[N] New level", true, textCol)
		g.buttonNewLevel = FromImageRectangle(button.Bounds())
	}
}

func (g *Gui) DrawPlaybackBar(screen *ebiten.Image) {
	// Background of playback bar.
	DrawSpriteStretched(screen, g.imgTextBackground)

	// Play/pause button.
	playbarHeight := screen.Bounds().Dy()
	playButtonWidth := playbarHeight
	playButtonHeight := playbarHeight
	playButton := SubImage(screen,
		Rectangle{IPt(0, 0), IPt(playButtonWidth, playButtonHeight)})
	if g.playbackPaused {
		DrawSpriteStretched(playButton, g.imgPlaybackPlay)
	} else {
		DrawSpriteStretched(playButton, g.imgPlaybackPause)
	}
	// Remember the region so that Update() can react when it's clicked.
	g.buttonPlaybackPlay = FromImageRectangle(playButton.Bounds())

	// Play bar.
	barXMargin := 10
	barX := playButtonWidth + barXMargin
	barWidth := screen.Bounds().Dx() - barX - barXMargin
	bar := SubImage(screen,
		Rectangle{IPt(barX, 0), IPt(barX+barWidth, playbarHeight)})
	DrawSpriteStretched(bar, g.imgPlayBar)
	// Remember the region so that Update() can react when it's clicked.
	g.buttonPlaybackBar = FromImageRectangle(bar.Bounds())

	// Playback bar cursor.
	cursorWidth := float64(playbarHeight)
	cursorHeight := float64(playbarHeight)
	factor := g.frameIdx.ToFloat64() / float64(len(g.playthrough.History))
	cursorX := factor*g.buttonPlaybackBar.Width().ToFloat64() - cursorWidth/2
	DrawSprite(bar, g.imgPlaybackCursor, cursorX, 0, cursorWidth, cursorHeight)
}
