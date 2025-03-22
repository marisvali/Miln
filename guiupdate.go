package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

func (g *Gui) Update() error {
	// Updates common to all states.
	if g.folderWatcher1.FolderContentsChanged() {
		g.loadGuiData()
	}
	if g.folderWatcher2.FolderContentsChanged() {
		// Reload world, and rely on the fact that this is makes the new world
		// load the new parameters from world.json.
		g.world = NewWorld(g.world.Seed, g.world.TargetDifficulty,
			g.EmbeddedFS)
		g.updateWindowSize()
	}

	// Get input once, so we don't need to get it every time we need it in
	// other functions.
	g.pressedKeys = g.pressedKeys[:0]
	g.pressedKeys = inpututil.AppendPressedKeys(g.pressedKeys)
	g.justPressedKeys = g.justPressedKeys[:0]
	g.justPressedKeys = inpututil.AppendJustPressedKeys(g.justPressedKeys)
	x, y := ebiten.CursorPosition()
	g.mousePt = IPt(x, y)
	g.leftButtonJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
	g.rightButtonJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2)

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
	if g.world.AllEnemiesDead() {
		g.state = GameWon
		g.uploadCurrentWorld()
		return
	}
	if g.world.Player.Health.Leq(ZERO) {
		g.state = GameLost
		g.uploadCurrentWorld()
		return
	}

	g.instructionalText = "Kill everyone! left click - move, right click - shoot"

	var input PlayerInput
	// Get input from player.
	input.MousePt = g.mousePt
	input.LeftButtonPressed = g.leftButtonJustPressed
	input.RightButtonPressed = g.rightButtonJustPressed

	if g.leftButtonJustPressed {
		input.Move, input.MovePt = g.GetMoveTarget()
	}
	if g.rightButtonJustPressed {
		input.Shoot, input.ShootPt = g.GetAttackTarget()
	}

	// input = g.ai.Step(&g.world)
	g.world.Step(input)
	g.visWorld.Step(&g.world, input, g.GuiData)

	if g.recordingFile != "" {
		WriteFile(g.recordingFile, g.world.SerializedPlaythrough())
	}
	if g.frameIdx.Mod(I(60)) == ZERO {
		g.uploadCurrentWorld()
	}

	g.frameIdx.Inc()
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

func (g *Gui) UserRequestedPlaybackPause() bool {
	return g.JustPressed(ebiten.KeySpace) || g.JustClicked(g.buttonPlaybackPlay)
}

func (g *Gui) UpdateGamePaused() {
	if g.UserRequestedNewLevel() {
		// seed, targetDifficulty := GetNextLevel(g.username)
		seed, targetDifficulty := RInt(I(0), I(1000000)), RInt(I(60), I(70))
		g.world = NewWorld(seed, targetDifficulty, g.EmbeddedFS)
		// g.world = NewWorld(RInt(I(0), I(10000000)), RInt(I(55), I(70)))
		// InitializeIdInDbSql(g.db, g.world.Id)
		InitializeIdInDbHttp(g.username, Version, g.world.Id)
		g.state = GameOngoing
		return
	}
	if g.UserRequestedRestartLevel() {
		g.world = NewWorld(g.world.Seed, g.world.TargetDifficulty,
			g.EmbeddedFS)
		// InitializeIdInDbSql(g.db, g.world.Id)
		InitializeIdInDbHttp(g.username, Version, g.world.Id)
		g.state = GameOngoing
		return
	}
	if g.leftButtonJustPressed || g.rightButtonJustPressed {
		g.state = GameOngoing
		g.UpdateGameOngoing()
		return
	}
	g.instructionalText = "Paused. Kill everyone! left click - move, right click - shoot"
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
	g.instructionalText = "You won, congratulations!"
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
	g.instructionalText = "You lost."
}

func (g *Gui) UpdatePlayback() {
	nFrames := I(len(g.playthrough.History))

	if g.UserRequestedPlaybackPause() {
		g.playbackPaused = !g.playbackPaused
	}

	// Choose target frame.
	targetFrameIdx := g.frameIdx

	// Compute the target frame index based on where on the play bar the user
	// clicked.
	if g.LeftClickPressedOn(g.buttonPlaybackBar) {
		// Get the distance between the start and the cursor on the play bar.
		dx := g.mousePt.X.Minus(g.buttonPlaybackBar.Corner1.X)
		targetFrameIdx = dx.Times(nFrames).DivBy(g.buttonPlaybackBar.Width())
	}

	if g.JustPressed(ebiten.KeyLeft) && g.Pressed(ebiten.KeyAlt) {
		targetFrameIdx.Subtract(g.FrameSkipAltArrow)
	}

	if g.JustPressed(ebiten.KeyRight) && g.Pressed(ebiten.KeyAlt) {
		targetFrameIdx.Add(g.FrameSkipAltArrow)
	}

	if g.Pressed(ebiten.KeyLeft) && g.Pressed(ebiten.KeyShift) {
		targetFrameIdx.Subtract(g.FrameSkipShiftArrow)
	}

	if g.Pressed(ebiten.KeyRight) && g.Pressed(ebiten.KeyShift) {
		targetFrameIdx.Add(g.FrameSkipShiftArrow)
	}

	if g.Pressed(ebiten.KeyLeft) && !g.Pressed(ebiten.KeyShift) && !g.Pressed(ebiten.KeyAlt) {
		targetFrameIdx.Subtract(g.FrameSkipArrow)
	}

	if g.Pressed(ebiten.KeyRight) && !g.Pressed(ebiten.KeyShift) && !g.Pressed(ebiten.KeyAlt) {
		targetFrameIdx.Add(g.FrameSkipArrow)
	}

	if targetFrameIdx.Lt(ZERO) {
		targetFrameIdx = ZERO
	}

	if targetFrameIdx.Geq(nFrames) {
		targetFrameIdx = nFrames.Minus(ONE)
	}

	if targetFrameIdx != g.frameIdx {
		// Rewind.
		g.world = NewWorld(g.playthrough.Seed, g.playthrough.TargetDifficulty, g.EmbeddedFS)
		g.visWorld = NewVisWorld(g.Animations)

		// Replay the world.
		for i := I(0); i.Lt(targetFrameIdx); i.Inc() {
			input := g.playthrough.History[i.ToInt()]
			g.world.Step(input)
			g.visWorld.Step(&g.world, input, g.GuiData)
		}

		// Set the current frame idx.
		g.frameIdx = targetFrameIdx
	}

	// Get input from recording.
	input := g.playthrough.History[g.frameIdx.ToInt()]
	// Remember cursor position in order to draw the virtual cursor during
	// Draw().
	g.mousePt = input.MousePt
	// Move the actual OS cursor on the screen.
	if g.MoveActualOSCursorDuringReplay {
		osPt := GameToOs(g.mousePt, g.layout)
		moveCursor(osPt)
	}

	// input = g.ai.Step(&g.world)
	if !g.playbackPaused {
		g.world.Step(input)
		g.visWorld.Step(&g.world, input, g.GuiData)

		if g.frameIdx.Lt(nFrames.Minus(ONE)) {
			g.frameIdx.Inc()
		}
	}

	g.instructionalText = fmt.Sprintf("Playing back frame %d / %d.",
		g.frameIdx.Plus(ONE).ToInt64(), nFrames.ToInt64())
	if g.world.AllEnemiesDead() {
		g.instructionalText += " Won."
	}
	if g.world.Player.Health.Leq(ZERO) {
		g.instructionalText += " Lost."
	}
}
