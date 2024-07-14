package world

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	. "github.com/marisvali/miln/gamelib"
	"math"
)

var nObstaclesMin = I(10)
var nObstaclesMax = I(25)

var nPortalsMin = I(1)
var nPortalsMax = I(5)
var SpawnPortalCooldownMin = I(40)
var SpawnPortalCooldownMax = I(150)

var GremlinMoveCooldown = I(50)
var GremlinFreezeCooldown = I(30)
var GremlinMaxHealth = I(1)
var nGremlinMin = I(3)
var nGremlinMax = I(7)

var HoundMoveCooldown = I(70)
var HoundFreezeCooldown = I(300)
var HoundMaxHealth = I(4)
var nHoundMin = I(0)
var nHoundMax = I(2)

var UltraHoundMoveCooldown = I(100)
var UltraHoundFreezeCooldown = I(10)
var UltraHoundMaxHealth = I(1)
var nUltraHoundMin = I(0)
var nUltraHoundMax = I(1)

var PillarMoveCooldown = I(100)
var PillarFreezeCooldown = I(300)
var PillarMaxHealth = I(1)
var KingMoveCooldown = I(60)
var KingFreezeCooldown = I(200)
var KingMaxHealth = I(3)
var QuestionMaxHealth = I(1)

type Beam struct {
	Idx Int // if this is greater than 0 it means the beam is active for Idx time steps
	End Pt  // this is the point to where the beam ends
}

const Version = 2

type World struct {
	Player           Player
	Enemies          []Enemy
	Beam             Beam
	Obstacles        MatBool
	AttackableTiles  MatBool
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
	Playthrough
	Seeds
}

type Playthrough struct {
	Id      uuid.UUID
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

func RandomLevel(nObstacles Int) (m MatBool) {
	// Create matrix with obstacles.
	m = NewMatBool(IPt(10, 10))
	for i := ZERO; i.Lt(nObstacles); i.Inc() {
		m.Set(m.RandomPos())
	}
	return
}

type PortalSeed struct {
	Cooldown     Int
	NGremlins    Int
	NHounds      Int
	NUltraHounds Int
}

type Seeds struct {
	NObstacles Int
	Portals    []PortalSeed
}

func GenerateSeeds(seed Int) (s Seeds) {
	RSeed(seed)
	s.NObstacles = RInt(nObstaclesMin, nObstaclesMax)
	nPortals := RInt(nPortalsMin, nPortalsMax)

	for i := ZERO; i.Lt(nPortals); i.Inc() {
		var portal PortalSeed
		portal.Cooldown = RInt(SpawnPortalCooldownMin, SpawnPortalCooldownMax)
		portal.NGremlins = RInt(nGremlinMin, nGremlinMax)
		portal.NHounds = RInt(nHoundMin, nHoundMax)
		portal.NUltraHounds = RInt(nUltraHoundMin, nUltraHoundMax)
		s.Portals = append(s.Portals, portal)
	}
	return
}

func GenerateSeedsTargetDifficulty(seed Int, target Int) (s Seeds) {
	RSeed(seed)
	for {
		s = GenerateSeeds(RInt(ZERO, I(1000000)))
		difficulty := ZERO
		difficulty.Add(s.NObstacles)
		difficulty.Add(I(len(s.Portals)))
		cd := ZERO
		for _, p := range s.Portals {
			difficulty.Add(p.NGremlins)
			difficulty.Add(p.NHounds.Times(I(2)))
			difficulty.Add(p.NUltraHounds.Times(I(2)))
			cd.Add(p.Cooldown)
		}
		cd = cd.DivBy(I(len(s.Portals)))
		difficulty.Add(cd.Times(I(2)).DivBy(I(100)))
		dif := target.Minus(difficulty)
		if dif.Abs().Leq(I(5)) {
			s.NObstacles.Add(dif)
			return
		}
	}
}

func NewWorld(seed Int) (w World) {
	w.Seed = seed
	w.Id = uuid.New()
	// w.Seeds = GenerateSeeds(seed)
	w.Seeds = GenerateSeedsTargetDifficulty(seed, I(53))
	w.Obstacles = RandomLevel(w.NObstacles)
	occ := w.Obstacles.Clone()
	for _, portal := range w.Portals {
		w.SpawnPortals = append(w.SpawnPortals, NewSpawnPortal(
			occ.OccupyRandomPos(),
			portal.Cooldown,
			portal.NGremlins,
			portal.NHounds,
			portal.NUltraHounds,
			I(0)))
	}
	w.SpawnPortals[0].nKingsLeftToSpawn = I(1)

	// Params
	w.BlockSize = I(1000)
	w.BeamMax = I(15)
	w.Player = NewPlayer()
	w.Player.HitPermissions.CanHitGremlin = true
	w.Player.HitPermissions.CanHitHound = true
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
	Serialize(buf, int64(Version))
	Serialize(buf, w.Seed.ToInt64())
	SerializeSlice(buf, w.History)
	return Zip(buf.Bytes())
}

func DeserializePlaythrough(data []byte) (p Playthrough) {
	buf := bytes.NewBuffer(Unzip(data))
	var token int64
	Deserialize(buf, &token)
	if token != Version {
		Check(fmt.Errorf("this code can't simulate this playthrough "+
			"correctly - we are version %d and playthrough was generated "+
			"with version %d",
			Version, token))
	}
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
	w.AttackableTiles = NewMatBool(w.Obstacles.Size())

	rows := w.Obstacles.Size().Y
	cols := w.Obstacles.Size().X
	w.beamPts = make([]Pt, rows.Times(cols).ToInt64())

	// Get a list of squares.
	squares := []Square{}
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			pt := Pt{x, y}
			if w.Obstacles.At(pt) {
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
					w.AttackableTiles.Set(Pt{x, y})
				} else {
					w.AttackableTiles.Clear(Pt{x, y})
					idx := w.AttackableTiles.PtToIndex(Pt{x, y}).ToInt()
					w.beamPts[idx] = pt
				}
			} else {
				w.AttackableTiles.Set(Pt{x, y})
			}
		}
	}

	// Get all attackable tiles connected to the player's pos.
	connectedTiles := w.AttackableTiles.ConnectedPositions(w.Player.Pos)

	// Eliminate tiles which were marked as attackable but are disconnected from
	// the attackable region that contains the player's position.
	// This is needed in order to eliminate tiles which are technically
	// reachable if you respect the math, but which just look weird to people.
	// Do the elimination by intersecting sets.
	w.AttackableTiles.IntersectWith(connectedTiles)
}

func (w *World) EnemyPositions() (m MatBool) {
	m = NewMatBool(w.Obstacles.Size())
	for i := range w.Enemies {
		m.Set(w.Enemies[i].Pos())
	}
	return
}

func (w *World) SpawnPortalPositions() (m MatBool) {
	m = NewMatBool(w.Obstacles.Size())
	for i := range w.SpawnPortals {
		m.Set(w.SpawnPortals[i].Pos)
	}
	return
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

	// Step Portals.
	for i := range w.SpawnPortals {
		w.SpawnPortals[i].Step(w)
	}

	// Cull dead Portals.
	newPortals := []SpawnPortal{}
	for i := range w.SpawnPortals {
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

// func RandomLevel1() (m Matrix[Int], pos1 []Pt, pos2 []Pt) {
// 	m = NewMatrix[Int](IPt(10, 10))
// 	for i := 0; i < 10; i++ {
// 		var pt Pt
// 		pt.X = RInt(ZERO, m.Size().X.Minus(ONE))
// 		pt.Y = RInt(ZERO, m.Size().Y.Minus(ONE))
// 		m.Set(pt, ONE)
// 	}
// 	pos1 = append(pos1, IPt(0, 0))
// 	pos2 = append(pos2, IPt(2, 2))
// 	return
// }
//
// func RandomLevel2() (m Matrix[Int]) {
// 	// Create matrix with obstacles.
// 	m = NewMatrix[Int](IPt(10, 10))
// 	for i := 0; i < 0; i++ {
// 		m.Set(m.RandomPos(), ONE)
// 	}
// 	return
// }
