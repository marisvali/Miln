package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	. "github.com/marisvali/miln/gamelib"
	_ "image/png"
)

func (g *Gui) loadGuiData() {
	// Read from the disk over and over until a full read is possible.
	// This repetition is meant to avoid crashes due to reading files
	// while they are still being written.
	// It's a hack but possibly a quick and very useful one.
	// This repeated reading is only useful when we're not reading from the
	// embedded filesystem. When we're reading from the embedded filesystem we
	// want to crash as soon as possible. We might be in the browser, in which
	// case we want to see an error in the developer console instead of a page
	// that keeps trying to load and reports nothing.
	if g.EmbeddedFS == nil {
		CheckCrashes = false
	}
	for {
		CheckFailed = nil
		g.LoadJSON("data/gui/gui.json", &g.GuiData)
		g.imgGround = g.LoadImage("data/gui/ground.png")
		g.imgTree = g.LoadImage("data/gui/tree.png")
		g.imgPlayerHealth = g.LoadImage("data/gui/player-health.png")
		g.imgPlayerAmmo = g.LoadImage("data/gui/player-ammo.png")
		g.imgHound = g.LoadImage("data/gui/enemy2.png")
		g.imgEnemyHealth = g.LoadImage("data/gui/enemy-health.png")
		g.imgEnemyCooldown = g.LoadImage("data/gui/enemy-cooldown.png")
		g.imgBeam = g.LoadImage("data/gui/beam.png")
		g.imgShadow = g.LoadImage("data/gui/shadow.png")
		g.imgTextBackground = g.LoadImage("data/gui/text-background.png")
		g.imgTextColor = g.LoadImage("data/gui/text-color.png")
		g.imgAmmo = g.LoadImage("data/gui/ammo.png")
		g.imgSpawnPortal = g.LoadImage("data/gui/spawn-portal.png")
		g.imgPlayerHitEffect = g.LoadImage("data/gui/player-hit-effect.png")
		g.imgHighlightMoveOk = g.LoadImage("data/gui/highlight-move-ok.png")
		g.imgHighlightMoveNotOk = g.LoadImage("data/gui/highlight-move-not-ok.png")
		g.imgHighlightAttack = g.LoadImage("data/gui/highlight-attack.png")
		g.imgBlack = g.LoadImage("data/gui/black.png")
		g.imgCursor = g.LoadImage("data/gui/cursor.png")
		g.animMoveFailed = g.NewAnimation("data/gui/move-failed")
		g.animAttackFailed = g.NewAnimation("data/gui/attack-failed")
		g.animPlayer1 = g.NewAnimation("data/gui/player1")
		g.animPlayer2 = g.NewAnimation("data/gui/player2")
		g.animHoundSearching = g.NewAnimation("data/gui/hound-searching")
		g.animHoundPreparingToAttack = g.NewAnimation("data/gui/hound-preparing-to-attack")
		g.animHoundAttacking = g.NewAnimation("data/gui/hound-attacking")
		g.animHoundHit = g.NewAnimation("data/gui/hound-hit")
		g.animHoundDead = g.NewAnimation("data/gui/hound-dead")
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true

	g.visWorld = NewVisWorld(g.Animations)
	g.updateWindowSize()
}

func (g *Gui) LoadJSON(filename string, v any) {
	if g.EmbeddedFS != nil {
		LoadJSONEmbedded(filename, g.EmbeddedFS, v)
	} else {
		LoadJSON(filename, v)
	}
}

func (g *Gui) LoadImage(filename string) *ebiten.Image {
	if g.EmbeddedFS != nil {
		return LoadImageEmbedded(filename, &embeddedFiles)
	} else {
		return LoadImage(filename)
	}
}

func (g *Gui) NewAnimation(filename string) Animation {
	if g.EmbeddedFS != nil {
		return NewAnimationEmbedded(filename, g.EmbeddedFS)
	} else {
		return NewAnimation(filename)
	}
}
