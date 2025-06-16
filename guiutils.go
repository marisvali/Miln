package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"slices"
	"strconv"
)

func (g *Gui) uploadCurrentWorld() {
	// Pass a World clone to the channel and not a serialized playthrough.
	// Serialization takes much longer than cloning.
	// Also, it is important for the channel to have some buffer, otherwise this
	// call will block until the previous world instance was uploaded.
	// If the connection to the server drops for a few seconds, either due to
	// the player's connection or the server not being available, it will
	// interrupt the gameplay.
	clone := g.world
	g.uploadDataChannel <- uploadData{g.username, Version, g.world.Id, &clone}
}

func UploadPlaythroughs(ch chan uploadData) {
	for {
		// Receive a playthrough from the channel.
		// Blocks until a playthrough is received.
		data := <-ch

		// Upload the data.
		UploadDataToDbHttp(data.user, data.version, data.id, data.world.SerializedPlaythrough())
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

func (g *Gui) TileToScreen(pos Pt) Pt {
	half := g.BlockSize.DivBy(TWO)
	return pos.Times(g.BlockSize).Plus(Pt{half, half}).Plus(Pt{g.guiMargin, g.guiMargin})
}

func (g *Gui) TilesToScreen(ipt []Pt) (opt []Pt) {
	for _, pt := range ipt {
		opt = append(opt, g.TileToScreen(pt))
	}
	return
}

func (g *Gui) ScreenToTile(pos Pt) Pt {
	return pos.Minus(Pt{g.guiMargin, g.guiMargin}).DivBy(g.BlockSize)
}

func (g *Gui) TileToPlayRegion(pos Pt) Pt {
	half := g.BlockSize.DivBy(TWO)
	return pos.Times(g.BlockSize).Plus(Pt{half, half})
}

func (g *Gui) WorldToGuiPos(pt Pt) Pt {
	return pt.Times(g.BlockSize).DivBy(g.world.BlockSize).Plus(Pt{g.guiMargin, g.guiMargin})
}

func (g *Gui) WorldToPlayRegion(pt Pt) Pt {
	return pt.Times(g.BlockSize).DivBy(g.world.BlockSize)
}

func (g *Gui) MouseCursorIsOverATile() bool {
	return g.world.Obstacles.InBounds(g.ScreenToTile(g.mousePt))
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

func (g *Gui) GetAttackTarget() (valid bool, target Pt) {
	if g.AutoAimAttack {
		attackablePositions := g.world.VulnerableEnemyPositions()
		attackablePositions.IntersectWith(g.world.VisibleTiles)
		pos := attackablePositions.ToArray()
		tilePos, dist := g.ClosestTileToMouse(pos.V[:pos.N])
		closeEnough := dist.Lt(g.BlockSize.Times(g.AutoAimAttackFactor).DivBy(I(100)))
		attackOk := g.world.Player.OnMap && closeEnough
		return attackOk, tilePos
	} else {
		attackablePositions := g.world.VulnerableEnemyPositions()
		attackablePositions.IntersectWith(g.world.VisibleTiles)
		tilePos := g.ScreenToTile(g.mousePt)
		mouseCursorIsOverAVulnerableEnemy :=
			attackablePositions.InBounds(tilePos) &&
				attackablePositions.At(tilePos)
		attackOk := g.world.Player.OnMap && mouseCursorIsOverAVulnerableEnemy
		return attackOk, tilePos
	}
}

func (g *Gui) GetMoveTarget() (valid bool, target Pt) {
	if g.AutoAimMove {
		freePositions := g.world.Player.ComputeFreePositions(&g.world).ToArray()
		tilePos, dist := g.ClosestTileToMouse(freePositions.V[:freePositions.N])
		closeEnough := dist.Lt(g.BlockSize.Times(g.AutoAimMoveFactor).DivBy(I(100)))
		return closeEnough, tilePos
	} else {
		freePositions := g.world.Player.ComputeFreePositions(&g.world)
		tilePos := g.ScreenToTile(g.mousePt)
		mouseCursorIsOverAFreePosition :=
			freePositions.InBounds(tilePos) &&
				freePositions.At(tilePos)
		return mouseCursorIsOverAFreePosition, tilePos
	}
}

func (g *Gui) JustPressed(key ebiten.Key) bool {
	return slices.Contains(g.justPressedKeys, key)
}

func (g *Gui) Pressed(key ebiten.Key) bool {
	return slices.Contains(g.pressedKeys, key)
}

func (g *Gui) JustClicked(button Rectangle) bool {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		return false
	}
	return button.ContainsPt(g.mousePt)
}

func (g *Gui) LeftClickPressedOn(button Rectangle) bool {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		return false
	}
	return button.ContainsPt(g.mousePt)
}
