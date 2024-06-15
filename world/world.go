package world

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "image/png"
	"math"
	. "playful-patterns.com/miln/gamelib"
)

type Player struct {
	Pos        Pt
	TimeoutIdx Int
	Health     Int
	MaxHealth  Int
}

type Enemy struct {
	Pos       Pt
	Health    Int
	MaxHealth Int
}

type Beam struct {
	Idx Int // if this is greater than 0 it means the beam is active for Idx time steps
	End Pt  // this is the point to where the beam ends
}

type World struct {
	Player          Player
	Enemies         []Enemy
	Beam            Beam
	Obstacles       Matrix
	AttackableTiles Matrix
	TimeStep        Int
	BeamMax         Int
	pathfinding     Pathfinding
	beamPts         []Pt
	BlockSize       Int
}

type PlayerInput struct {
	Move    bool
	MovePt  Pt // tile-coordinates
	Shoot   bool
	ShootPt Pt // tile-coordinates
}

func SerializeInputs(inputs []PlayerInput, filename string) {
	buf := new(bytes.Buffer)
	SerializeSlice(buf, inputs)
	Zip(filename, buf.Bytes())
}

func DeserializeInputs(filename string) []PlayerInput {
	var inputs []PlayerInput
	buf := bytes.NewBuffer(Unzip(filename))
	DeserializeSlice(buf, &inputs)
	return inputs
}

var playerCooldown Int = I(15)
var enemyCooldown Int = I(40)

func (w *World) TileToWorldPos(pt Pt) Pt {
	half := w.BlockSize.DivBy(TWO)
	offset := Pt{half, half}
	return pt.Times(w.BlockSize).Plus(offset)
}

func (w *World) computeAttackableTiles() {
	// Compute which tiles are attackable.
	w.AttackableTiles.Init(w.Obstacles.Size())

	rows := w.Obstacles.Size().Y
	cols := w.Obstacles.Size().X
	w.beamPts = make([]Pt, rows.Times(cols).ToInt64())

	// Get a list of squares.
	squares := []Square{}
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			pt := Pt{x, y}
			if !w.Obstacles.Get(pt).IsZero() {
				center := w.TileToWorldPos(pt)
				size := w.BlockSize.Times(I(90)).DivBy(I(100))
				squares = append(squares, Square{center, size})
			}
		}
	}

	// Draw a line from the player's pos to each of the tiles and test if that
	// line intersects the squares.
	lineStart := w.TileToWorldPos(w.Player.Pos)
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			// Check if tile can be attacked.
			lineEnd := w.TileToWorldPos(Pt{x, y})
			l := Line{lineStart, lineEnd}
			if intersects, pt := LineSquaresIntersection(l, squares); intersects {
				w.AttackableTiles.Set(Pt{x, y}, ZERO)
				idx := w.AttackableTiles.PtToIndex(Pt{x, y}).ToInt()
				w.beamPts[idx] = pt
			} else {
				w.AttackableTiles.Set(Pt{x, y}, ONE)
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
			w.Player.TimeoutIdx = playerCooldown
		}
	}

	// See about the beam.
	if w.Beam.Idx.Gt(ZERO) {
		w.Beam.Idx.Dec()
	}
	if input.Shoot && w.Player.TimeoutIdx.Eq(ZERO) {
		shotEnemies := []*Enemy{}
		for i, _ := range w.Enemies {
			if w.Enemies[i].Pos.Eq(input.ShootPt) {
				shotEnemies = append(shotEnemies, &w.Enemies[i])
			}
		}

		if len(shotEnemies) > 0 {
			w.Beam.Idx = w.BeamMax // show beam
			if w.AttackableTiles.Get(input.ShootPt).Neq(ZERO) {
				w.Player.TimeoutIdx = playerCooldown
				for _, enemy := range shotEnemies {
					enemy.Health.Dec()
				}
				w.Beam.End = w.TileToWorldPos(input.ShootPt)
			} else {
				idx := w.AttackableTiles.PtToIndex(input.ShootPt).ToInt()
				w.Beam.Idx = w.BeamMax
				w.Beam.End = w.beamPts[idx]
			}
		}
	}

	// Cull dead enemies.
	// This kind of operation makes me thing I should have a slice of pointers,
	// not values.
	newEnemies := []Enemy{}
	for i, _ := range w.Enemies {
		if w.Enemies[i].Health.IsPositive() {
			newEnemies = append(newEnemies, w.Enemies[i])
		}
	}
	w.Enemies = newEnemies

	w.TimeStep.Inc()
	if w.TimeStep.Eq(I(math.MaxInt64)) {
		// Damn.
		Check(fmt.Errorf("got to an unusually large time step: %d", w.TimeStep.ToInt64()))
	}

	// Get keyboard input.
	var pressedKeys []ebiten.Key
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)

	// Move the enemies.
	if w.TimeStep.Mod(enemyCooldown).Eq(ZERO) {
		for i, _ := range w.Enemies {
			path := w.pathfinding.FindPath(w.Enemies[i].Pos, w.Player.Pos)
			if len(path) > 1 {
				w.Enemies[i].Pos = path[1]
				if w.Enemies[i].Pos.Eq(w.Player.Pos) {
					w.Player.Health.Dec()
				}
			}
		}
	}
}

func RandomLevel1() (m Matrix, pos1 []Pt, pos2 []Pt) {
	m.Init(IPt(10, 10))
	for i := 0; i < 10; i++ {
		var pt Pt
		pt.X = RInt(ZERO, m.Size().X.Minus(ONE))
		pt.Y = RInt(ZERO, m.Size().Y.Minus(ONE))
		m.Set(pt, ONE)
	}
	pos1 = append(pos1, IPt(0, 0))
	pos2 = append(pos2, IPt(2, 2))
	return
}

func RandomLevel2() (m Matrix, pos1 []Pt, pos2 []Pt) {
	m.Init(IPt(10, 10))
	for i := 0; i < 10; i++ {
		m.Set(m.RPos(), ONE)
	}
	pos1 = append(pos1, IPt(0, 0))
	for i := 0; i < 10; i++ {
		pt := m.RPos()
		if m.Get(pt).IsZero() {
			pos2 = append(pos2, pt)
		}
	}
	return
}

func (w *World) Initialize() {
	// Obstacles
	//g.world.Obstacles.Init(I(15), I(15))
	pos1 := []Pt{}
	pos2 := []Pt{}
	//g.world.Obstacles, pos1, pos2 = LevelFromString(Level1())
	w.Obstacles, pos1, pos2 = RandomLevel2()
	if len(pos1) > 0 {
		w.Player.Pos = pos1[0]
	}
	for _, enemyPos := range pos2 {
		enemy := Enemy{}
		enemy.Pos = enemyPos
		enemy.MaxHealth = I(5)
		enemy.Health = enemy.MaxHealth
		w.Enemies = append(w.Enemies, enemy)
	}
	w.pathfinding.Initialize(w.Obstacles)

	// Params
	w.BlockSize = I(1000)
	w.BeamMax = I(15)
	w.Player.MaxHealth = I(3)
	w.Player.Health = w.Player.MaxHealth
}
