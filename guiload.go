package main

import (
	. "github.com/marisvali/miln/gamelib"
	_ "image/png"
	"os"
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
	if g.FSys == os.DirFS(".") {
		CheckCrashes = false
	}
	for {
		CheckFailed = nil
		LoadYAML(g.FSys, "data/gui/gui.yaml", &g.GuiData)
		g.imgGround = LoadImage(g.FSys, "data/gui/ground.png")
		g.imgTree = LoadImage(g.FSys, "data/gui/tree.png")
		g.imgPlayerHealth = LoadImage(g.FSys, "data/gui/player-health.png")
		g.imgPlayerAmmo = LoadImage(g.FSys, "data/gui/player-ammo.png")
		g.imgHound = LoadImage(g.FSys, "data/gui/enemy2.png")
		g.imgEnemyHealth = LoadImage(g.FSys, "data/gui/enemy-health.png")
		g.imgEnemyCooldown = LoadImage(g.FSys, "data/gui/enemy-cooldown.png")
		g.imgBeam = LoadImage(g.FSys, "data/gui/beam.png")
		g.imgShadow = LoadImage(g.FSys, "data/gui/shadow.png")
		g.imgTextBackground = LoadImage(g.FSys, "data/gui/text-background.png")
		g.imgTextColor = LoadImage(g.FSys, "data/gui/text-color.png")
		g.imgAmmo = LoadImage(g.FSys, "data/gui/ammo.png")
		g.imgSpawnPortal = LoadImage(g.FSys, "data/gui/spawn-portal.png")
		g.imgPlayerHitEffect = LoadImage(g.FSys, "data/gui/player-hit-effect.png")
		g.imgHighlightMoveOk = LoadImage(g.FSys, "data/gui/highlight-move-ok.png")
		g.imgHighlightMoveNotOk = LoadImage(g.FSys, "data/gui/highlight-move-not-ok.png")
		g.imgHighlightAttack = LoadImage(g.FSys, "data/gui/highlight-attack.png")
		g.imgBlack = LoadImage(g.FSys, "data/gui/black.png")
		g.imgCursor = LoadImage(g.FSys, "data/gui/cursor.png")
		g.imgPlaybackPlay = LoadImage(g.FSys, "data/gui/playback-play.png")
		g.imgPlaybackPause = LoadImage(g.FSys, "data/gui/playback-pause.png")
		g.imgPlayBar = LoadImage(g.FSys, "data/gui/playbar.png")
		g.imgPlaybackCursor = LoadImage(g.FSys, "data/gui/playback-cursor.png")
		g.imgEnemyTargetPos = LoadImage(g.FSys, "data/gui/enemy-target-pos.png")

		g.animMoveFailed = NewAnimation(g.FSys, "data/gui/move-failed")
		g.animAttackFailed = NewAnimation(g.FSys, "data/gui/attack-failed")
		g.animPlayer1 = NewAnimation(g.FSys, "data/gui/player1")
		g.animPlayer2 = NewAnimation(g.FSys, "data/gui/player2")
		g.animHoundSearching = NewAnimation(g.FSys, "data/gui/hound-searching")
		g.animHoundPreparingToAttack = NewAnimation(g.FSys, "data/gui/hound-preparing-to-attack")
		g.animHoundAttacking = NewAnimation(g.FSys, "data/gui/hound-attacking")
		g.animHoundHit = NewAnimation(g.FSys, "data/gui/hound-hit")
		g.animHoundDead = NewAnimation(g.FSys, "data/gui/hound-dead")
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true

	g.visWorld = NewVisWorld(g.Animations)
	g.updateWindowSize()
}
