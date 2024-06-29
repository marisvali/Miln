package world

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	. "github.com/marisvali/miln/gamelib"
	"math"
)

var PlayerMoveCooldown = I(20)
var GremlinMoveCooldown = I(50)
var GremlinFreezeCooldown = I(30)
var GremlinMaxHealth = I(1)
var HoundMoveCooldown = I(70)
var HoundFreezeCooldown = I(300)
var HoundMaxHealth = I(1)
var UltraHoundMoveCooldown = I(100)
var UltraHoundFreezeCooldown = I(10)
var UltraHoundMaxHealth = I(1)
var PillarMoveCooldown = I(100)
var PillarFreezeCooldown = I(300)
var PillarMaxHealth = I(1)
var KingMoveCooldown = I(60)
var KingFreezeCooldown = I(200)
var KingMaxHealth = I(3)
var QuestionMaxHealth = I(1)
var SpawnPortalCooldown = I(300)

type Beam struct {
	Idx Int // if this is greater than 0 it means the beam is active for Idx time steps
	End Pt  // this is the point to where the beam ends
}

type World struct {
	Player           Player
	Enemies          []Enemy
	Beam             Beam
	Obstacles        Matrix[Int]
	AttackableTiles  Matrix[Int]
	TimeStep         Int
	BeamMax          Int
	beamPts          []Pt
	BlockSize        Int
	Ammos            []Ammo
	SpawnPortals     []SpawnPortal
	Keys             []Key
	PillarKeyDropped bool
	HoundKeyDropped  bool
	PortalKeyDropped bool
	KingSpawned      bool
	Id               uuid.UUID
	Playthrough
}

type Playthrough struct {
	Seed    Int
	History []PlayerInput
}

type PlayerInput struct {
	Move    bool
	MovePt  Pt // tile-coordinates
	Shoot   bool
	ShootPt Pt // tile-coordinates
}

func LevelX() string {
	return `
----x----
--x---xx-
--x----x-
--x------
--x--xxx-
---------
---x-----
-x---xxxx
-xx------
---------
`
}

func NewWorld(seed Int) (w World) {
	w.Seed = seed
	w.Id = uuid.New()
	RSeed(seed)

	m, _, _ := LevelFromString(LevelX())
	w.Obstacles = NewMatrix[Int](m.Size())
	for y := ZERO; y.Lt(m.Size().Y); y.Inc() {
		for x := ZERO; x.Lt(m.Size().X); x.Inc() {
			if m.Get(Pt{x, y}) != ZERO {
				w.Enemies = append(w.Enemies, NewQuestion(Pt{x, y}))
			}
		}
	}

	// Obstacles
	//w.Obstacles = RandomLevel2()

	// Place each item at an unoccupied position (and occupy that position).
	//occ := w.Obstacles.Clone() // Keeps track of occupied positions.

	//var limit int
	//limit = RInt(I(2), I(4)).ToInt()
	//for i := 0; i < limit; i++ {
	//	w.Enemies = append(w.Enemies, NewEnemy(ZERO, occ.NewlyOccupiedRandomPos()))
	//}
	//limit = RInt(I(17), I(20)).ToInt()
	//for i := 0; i < limit; i++ {
	//	w.Enemies = append(w.Enemies, NewQuestion(occ.NewlyOccupiedRandomPos()))
	//}
	//limit = RInt(I(1), I(1)).ToInt()
	//for i := 0; i < limit; i++ {
	//	w.Enemies = append(w.Enemies, NewEnemy(TWO, occ.NewlyOccupiedRandomPos()))
	//}

	//w.SpawnPortals = append(w.SpawnPortals, NewSpawnPortal(occ.NewlyOccupiedRandomPos()))

	//w.Enemies = append(w.Enemies, NewEnemy(I(4), occ.NewlyOccupiedRandomPos()))

	// Params
	w.BlockSize = I(1000)
	w.BeamMax = I(15)
	w.Player = NewPlayer()
	w.Player.HitPermissions.CanHitGremlin = true
	w.Player.HitPermissions.CanHitQuestion = true
	w.Player.HitPermissions.CanHitKing = true

	// GUI needs this even without the world ever doing a step.
	// Note: this was true when the player started on the map, so it might not
	// be relevant now that the player doesn't start on the map. But, keep it
	// in case things change again.
	w.computeAttackableTiles()
	return
}

func (w *World) SerializedPlaythrough() []byte {
	buf := new(bytes.Buffer)
	Serialize(buf, w.Seed.ToInt64())
	SerializeSlice(buf, w.History)
	return Zip(buf.Bytes())
}

func DeserializePlaythrough(data []byte) (p Playthrough) {
	buf := bytes.NewBuffer(Unzip(data))
	var token int64
	Deserialize(buf, &token)
	p.Seed = I64(token)
	DeserializeSlice(buf, &p.History)
	return
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
	w.AttackableTiles = NewMatrix[Int](w.Obstacles.Size())

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
		center := w.TileToWorldPos(enemy.Pos())
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

	// Mark all attackable tiles connected to the start with TWO.
	FloodFill(w.AttackableTiles, w.Player.Pos, TWO)

	// Eliminate tiles which were marked as attackable but are disconnected from
	// the attackable region that contains the player's position.
	// This is needed in order to eliminate tiles which are technically
	// reachable if you respect the math, but which just look weird to people.
	// Do the elimination by marking ONEs with ZERO and TWOs with ONEs.
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			if w.AttackableTiles.Get(Pt{x, y}) == TWO {
				w.AttackableTiles.Set(Pt{x, y}, ONE)
			} else if w.AttackableTiles.Get(Pt{x, y}) == ONE {
				w.AttackableTiles.Set(Pt{x, y}, ZERO)
			}
		}
	}
}

func (w *World) Step(input PlayerInput) {
	w.History = append(w.History, input)
	w.computeAttackableTiles()
	w.Player.Step(w, input)

	// Step the enemies.
	for i := range w.Enemies {
		w.Enemies[i].Step(w)
	}

	// Cull dead enemies.
	newEnemies := []Enemy{}
	for i := range w.Enemies {
		if w.Enemies[i].Alive() {
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

func RandomLevel1() (m Matrix[Int], pos1 []Pt, pos2 []Pt) {
	m = NewMatrix[Int](IPt(10, 10))
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

func RandomLevel2() (m Matrix[Int]) {
	// Create matrix with obstacles.
	m = NewMatrix[Int](IPt(10, 10))
	for i := 0; i < 0; i++ {
		m.Set(m.RandomPos(), ONE)
	}
	return
}
