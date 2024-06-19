package world

import (
	"bytes"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	_ "image/png"
	"math"
)

var playerCooldown Int = I(1)
var enemyCooldowns []Int = []Int{I(40), I(20), I(100), I(120)}
var enemyHealths []Int = []Int{I(1), I(4), I(2), I(4)}
var enemyFrozenCooldowns []Int = []Int{I(130), I(130), I(130), I(130)}
var spawnPortalCooldown Int = I(100)

type Player struct {
	Pos        Pt
	TimeoutIdx Int
	Health     Int
	MaxHealth  Int
	AmmoCount  Int
}

type Enemy struct {
	Pos       Pt
	Health    Int
	MaxHealth Int
	FrozenIdx Int
	MaxFrozen Int
	Type      Int
}

type SpawnPortal struct {
	Pos        Pt
	Health     Int
	MaxHealth  Int
	MaxTimeout Int
	TimeoutIdx Int
}

type Ammo struct {
	Pos   Pt
	Count Int
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
	beamPts         []Pt
	BlockSize       Int
	Ammos           []Ammo
	SpawnPortals    []SpawnPortal
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

func (w *World) TileToWorldPos(pt Pt) Pt {
	half := w.BlockSize.DivBy(TWO)
	offset := Pt{half, half}
	return pt.Times(w.BlockSize).Plus(offset)
}

func (w *World) WorldPosToTile(pt Pt) Pt {
	return pt.DivBy(w.BlockSize)
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
				size := w.BlockSize.Times(I(98)).DivBy(I(100))
				squares = append(squares, Square{center, size})
			}
		}
	}
	for _, enemy := range w.Enemies {
		center := w.TileToWorldPos(enemy.Pos)
		size := w.BlockSize.Times(I(98)).DivBy(I(100))
		squares = append(squares, Square{center, size})
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
				obstacleTile := w.WorldPosToTile(pt)
				if obstacleTile.Eq(Pt{x, y}) {
					w.AttackableTiles.Set(Pt{x, y}, ONE)
				} else {
					w.AttackableTiles.Set(Pt{x, y}, ZERO)
					idx := w.AttackableTiles.PtToIndex(Pt{x, y}).ToInt()
					w.beamPts[idx] = pt
				}
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
		if w.Obstacles.Get(input.MovePt).Eq(ZERO) &&
			w.AttackableTiles.Get(input.MovePt).Neq(ZERO) {
			w.Player.Pos = input.MovePt
			w.Player.TimeoutIdx = playerCooldown

			// Collect ammos.
			newAmmos := make([]Ammo, 0)
			for i := range w.Ammos {
				if w.Ammos[i].Pos == w.Player.Pos {
					w.Player.AmmoCount.Add(w.Ammos[i].Count)
				} else {
					newAmmos = append(newAmmos, w.Ammos[i])
				}
			}
			w.Ammos = newAmmos
		}
	}

	// Spawn new ammos
	for {
		if len(w.Ammos) == 1 {
			break
		}

		if w.Player.AmmoCount.IsPositive() {
			break
		}

		pt := w.Obstacles.RPos()
		if !w.Obstacles.Get(pt).IsZero() {
			continue
		}
		if w.Player.Pos == pt {
			continue
		}
		invalid := false
		for i := range w.Ammos {
			if w.Ammos[i].Pos == pt {
				invalid = true
				break
			}
		}
		if invalid {
			continue
		}
		ammo := Ammo{
			Pos:   pt,
			Count: I(3),
		}
		w.Ammos = append(w.Ammos, ammo)
	}

	// See about the beam.
	if w.Beam.Idx.Gt(ZERO) {
		w.Beam.Idx.Dec()
	}
	if input.Shoot &&
		w.Player.TimeoutIdx.Eq(ZERO) &&
		!w.AttackableTiles.Get(input.ShootPt).IsZero() {

		shotEnemies := []*Enemy{}
		for i, _ := range w.Enemies {
			if w.Enemies[i].Pos.Eq(input.ShootPt) {
				shotEnemies = append(shotEnemies, &w.Enemies[i])
			}
		}

		shotPortals := []*SpawnPortal{}
		for i, _ := range w.SpawnPortals {
			if w.SpawnPortals[i].Pos.Eq(input.ShootPt) {
				shotPortals = append(shotPortals, &w.SpawnPortals[i])
			}
		}

		if len(shotEnemies) > 0 || len(shotPortals) > 0 {
			w.Beam.Idx = w.BeamMax // show beam
			w.Player.TimeoutIdx = playerCooldown
			w.Beam.End = w.TileToWorldPos(input.ShootPt)
		}
	}

	// Step the enemies.
	for i, _ := range w.Enemies {
		w.Enemies[i].Step(w)
	}

	// Cull dead enemies.
	// This kind of operation makes me think I should have a slice of pointers,
	// not values.
	newEnemies := []Enemy{}
	for i, _ := range w.Enemies {
		if w.Enemies[i].Health.IsPositive() {
			newEnemies = append(newEnemies, w.Enemies[i])
		}
	}
	w.Enemies = newEnemies

	// Step portals.
	for i := range w.SpawnPortals {
		w.SpawnPortals[i].Step(w)
	}

	// Cull dead portals.
	newPortals := []SpawnPortal{}
	for i, _ := range w.SpawnPortals {
		if w.SpawnPortals[i].Health.IsPositive() {
			newPortals = append(newPortals, w.SpawnPortals[i])
		}
	}
	w.SpawnPortals = newPortals

	w.TimeStep.Inc()
	if w.TimeStep.Eq(I(math.MaxInt64)) {
		// Damn.
		Check(fmt.Errorf("got to an unusually large time step: %d", w.TimeStep.ToInt64()))
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

func RandomLevel2() (m Matrix, pos1 []Pt, pos2 []Pt, pos3 []Pt, pos4 []Pt) {
	// Create matrix with obstacles.
	m.Init(IPt(10, 10))
	for i := 0; i < 5; i++ {
		m.Set(m.RPos(), ONE)
	}

	// Search for a non-occupied position to place the player in.
	for {
		pt := m.RPos()
		if m.Get(pt).IsZero() {
			pos1 = append(pos1, pt)
			break
		}
	}

	// Place n enemies at unoccupied positions.
	for {
		pt := m.RPos()
		if m.Get(pt).IsZero() && !pt.Eq(pos1[0]) {
			pos2 = append(pos2, pt)
			if len(pos2) == 10 {
				break
			}
		}
	}

	// Search for a non-occupied position to place the ammo in.
	for {
		pt := m.RPos()
		if m.Get(pt).IsZero() {
			pos3 = append(pos3, pt)
			break
		}
	}

	// Search for a non-occupied position to place the spawn portal in.
	for {
		pt := m.RPos()
		if m.Get(pt).IsZero() {
			pos4 = append(pos4, pt)
			break
		}
	}

	return
}

func NewEnemy(enemyType Int, pos Pt) Enemy {
	e := Enemy{}
	e.Type = enemyType
	e.Pos = pos
	e.MaxHealth = enemyHealths[e.Type.ToInt()]
	e.Health = e.MaxHealth
	e.MaxFrozen = enemyFrozenCooldowns[e.Type.ToInt()]
	e.FrozenIdx = e.MaxFrozen.DivBy(TWO)
	return e
}

func (w *World) Initialize() {
	// Obstacles
	//g.world.Obstacles.Init(I(15), I(15))
	pos1 := []Pt{}
	pos2 := []Pt{}
	pos3 := []Pt{}
	pos4 := []Pt{}
	//g.world.Obstacles, pos1, pos2 = LevelFromString(Level1())
	w.Obstacles, pos1, pos2, pos3, pos4 = RandomLevel2()
	if len(pos1) > 0 {
		w.Player.Pos = pos1[0]
	}
	for _, pos := range pos2 {
		w.Enemies = append(w.Enemies, NewEnemy(RInt(I(0), I(3)), pos))
	}

	for _, pos := range pos3 {
		ammo := Ammo{}
		ammo.Pos = pos
		ammo.Count = I(1)
		w.Ammos = append(w.Ammos, ammo)
	}

	for _, pos := range pos4 {
		portal := SpawnPortal{}
		portal.Pos = pos
		portal.MaxHealth = I(1)
		portal.Health = portal.MaxHealth
		portal.MaxTimeout = spawnPortalCooldown
		w.SpawnPortals = append(w.SpawnPortals, portal)
	}

	// Params
	w.BlockSize = I(1000)
	w.BeamMax = I(15)
	w.Player.MaxHealth = I(3)
	w.Player.Health = w.Player.MaxHealth

	// GUI needs this even without the world ever doing a step.
	w.computeAttackableTiles()
}

func (e *Enemy) Step(w *World) {
	if w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		beamEndTile := w.WorldPosToTile(w.Beam.End)
		if beamEndTile.Eq(e.Pos) {
			// We have been shot.
			e.Health.Dec()
			e.FrozenIdx = e.MaxFrozen
		}
	}

	if e.FrozenIdx.IsPositive() {
		e.FrozenIdx.Dec()
		return // Don't move.
	}
	if w.TimeStep.Mod(enemyCooldowns[e.Type.ToInt()]).Neq(ZERO) {
		return
	}

	// Move.
	// Clone obstacle matrix and put (other) enemies on it.
	allObstacles := w.Obstacles.Clone()
	for _, enemy := range w.Enemies {
		if !enemy.Pos.Eq(e.Pos) {
			allObstacles.Set(enemy.Pos, TWO)
		}
	}

	path := FindPath(e.Pos, w.Player.Pos, allObstacles)
	if len(path) > 1 {
		e.Pos = path[1]
		if e.Pos.Eq(w.Player.Pos) {
			w.Player.Health.Dec()
		}
	}
}

func (p *SpawnPortal) Step(w *World) {
	if w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		beamEndTile := w.WorldPosToTile(w.Beam.End)
		if beamEndTile.Eq(p.Pos) {
			if w.Player.AmmoCount.Gt(I(0)) {
				// We have been shot.
				p.Health.Dec()
				w.Player.AmmoCount.Dec()
			}
		}
	}

	if p.TimeoutIdx.IsPositive() {
		p.TimeoutIdx.Dec()
		return // Don't spawn.
	}

	// Spawn guy.
	// Check if there is already a guy here.
	occupied := false
	for _, enemy := range w.Enemies {
		if enemy.Pos == p.Pos {
			occupied = true
			break
		}
	}
	if occupied {
		return // Don't spawn.
	}

	w.Enemies = append(w.Enemies, NewEnemy(RInt(I(0), I(3)), p.Pos))
	p.TimeoutIdx = p.MaxTimeout
}
