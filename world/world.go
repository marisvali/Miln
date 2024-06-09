package world

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "image/png"
	"math"
	. "playful-patterns.com/miln/geometry"
	. "playful-patterns.com/miln/ints"
	. "playful-patterns.com/miln/pathfinding"
	. "playful-patterns.com/miln/point"
	. "playful-patterns.com/miln/utils"
)

type Player struct {
	Pos        Pt
	TimeoutIdx Int
	Health     Int
}

type Enemy struct {
	Pos    Pt
	Health Int
}

type World struct {
	Player          Player
	Enemy           Enemy
	Obstacles       Matrix
	AttackableTiles Matrix
	TimeStep        Int
	pathfinding     Pathfinding
}

type PlayerInput struct {
	Move    bool
	MovePt  Pt // tile-coordinates
	Shoot   bool
	ShootPt Pt // tile-coordinates
}

var PlayerCooldown Int = I(15)
var EnemyCooldown Int = I(40)

func (w *World) computeAttackableTiles() {
	// Compute which tiles are attackable.
	w.AttackableTiles.Init(w.Obstacles.Size())

	var enlargeConstant = I(1000)
	rows := w.Obstacles.Size().Y
	cols := w.Obstacles.Size().X

	// Get a list of squares.
	squares := []Square{}
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			center := Pt{x, y}.Times(enlargeConstant)
			size := enlargeConstant.Times(I(90)).DivBy(I(100))
			squares = append(squares, Square{center, size})
		}
	}

	// Draw a line from the player's pos to each of the tiles and test if that
	// line intersects the squares.
	lineStart := w.Player.Pos.Times(enlargeConstant)
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			// Check if tile can be attacked.
			lineEnd := Pt{x, y}.Times(enlargeConstant)
			l := Line{lineStart, lineEnd}
			if intersects, _ := LineSquaresIntersection(l, squares); !intersects {
				w.AttackableTiles.Set(Pt{x, y}, ONE)
			} else {
				w.AttackableTiles.Set(Pt{x, y}, ZERO)
			}
		}
	}
}

func (w *World) Step(input *PlayerInput) {
	if w.Player.TimeoutIdx.Gt(ZERO) {
		w.Player.TimeoutIdx.Dec()
	}

	w.computeAttackableTiles()

	if input.Move && w.Player.TimeoutIdx.Eq(ZERO) {
		if w.Obstacles.Get(input.MovePt).Eq(ZERO) {
			w.Player.Pos = input.MovePt
			w.Player.TimeoutIdx = PlayerCooldown
		}
	}

	if input.Shoot && w.Player.TimeoutIdx.Eq(ZERO) {
		if w.AttackableTiles.Get(input.ShootPt).Neq(ZERO) {
			w.Player.TimeoutIdx = PlayerCooldown
			w.Enemy.Health.Dec()
		}
	}

	w.TimeStep.Inc()
	if w.TimeStep.Eq(I(math.MaxInt64)) {
		// Damn.
		Check(fmt.Errorf("got to an unusually large time step: %d", w.TimeStep.ToInt64()))
	}

	// Get keyboard input.
	var pressedKeys []ebiten.Key
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)

	// Move the enemy.
	if w.TimeStep.Mod(EnemyCooldown).Eq(ZERO) {
		path := w.pathfinding.FindPath(w.Enemy.Pos, w.Player.Pos)
		if len(path) > 1 {
			w.Enemy.Pos = path[1]
		}
	}
}

func (w *World) Initialize() {
	w.pathfinding.Initialize(w.Obstacles)
}
