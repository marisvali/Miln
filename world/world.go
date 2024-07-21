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

const Version = 6

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
	vision           Vision
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
	w.Seeds = GenerateSeedsTargetDifficulty(seed, I(60))
	w.Obstacles = RandomLevel(w.NObstacles)
	w.vision = NewVision(w.Obstacles.Size())
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

// RegressionId returns a string which uniquely identifies the world in this
// state. It is meant to be used this way:
// - Compute the RegressionId for the world after a playthrough.
// - Refactor the implementation of the World.
// - Compute the RegressionId for the world after a playthrough that uses the
// exact same inputs.
// - If two worlds return the same RegressionId, they are (pretty much*)
// identical. The refactoring did not alter the playthrough.
// - If two worlds return a different RegressionId, they are not the same, which
// should indicate a regression failure. Something in the refactoring is now
// causing the play experience to be different.
//
// * "pretty much" is meant to mean, so similar that it would be nearly
// impossible to get to the same state with the same inputs, but have the
// gameplay be any different. This is an assumption and it relies on how quickly
// a simulation goes insane if any part of it doesn't act right. But it is very
// dependent on how long the playthrough is and how many world elements it
// involves. If it's just the player starting with some obstacles and no enemies
// and winning after 1 frame, that won't catch errors with refactoring enemy
// behavior.
func (w *World) RegressionId() string {
	// The world is "the same" if it has the same:
	// - player pos and health
	// - enemy types, positions and health
	// - obstacles
	//
	// Explanation:
	//
	// Here we need a definition of what it means that the world is "the same"
	// after its implementation changed.
	//
	// Option 1 - check all bits
	// -------------------------
	//
	// The most straightforward test is to check if it contains exactly the same
	// bits at the end. But this would mean that any change in the data
	// structures of the world would have to be a breaking change, which is not
	// exactly what I'm looking for. If I can get the same behavior relevant
	// for the outside as before, I should be free to change the world.
	//
	// Option 2 - check what the GUI shows
	// -----------------------------------
	//
	// Following the previous reasoning, I could say that whatever is shown to
	// the player is the actual state of the world. So I should first of all
	// have a more well-defined interface between the world and the GUI so that
	// I know exactly what the GUI gets from the world. This way, I can use that
	// as the bits I check to be the same.
	// The disadvantage here is that I'm still changing things a lot and I
	// don't want the friction of an interface that I adjust every time I change
	// something. I want everything in the World public because I want to
	// inspect it either from the GUI, or a future AI or some analysis script.
	// Another disadvantage is that I might want to include things that are not
	// shown in the GUI.
	//
	// Option 3 - freestyle but kind of follow the GUI (selected)
	// -----------------------------------------------
	//
	// Following the previous reasoning, I could say that I can follow what is
	// shown in the GUI as a sanity check that I'm including everything that
	// sounds reasonable for a check like this. But, I provide my own definition
	// here for what it means that two worlds at the same.
	// WARNING: this method makes some assumptions.
	// a. I assume that a reasonable playthrough is provided where the player
	// goes through an entire level, mostly winning.
	// b. I assume that the player's moves are highly relevant, as in, if you
	// take out any of the moves or make a significant deviation, the simulation
	// goes in a very different direction quickly. You don't get this if the
	// player doesn't do anything, for example, or makes moves which have no
	// impact on the game (hard to do, but possible).
	// c. I assume that the playthrough contains enough World elements that the
	// regression test makes sense. If the playthrough contains no enemies, the
	// regression test will not catch any changes in enemy behavior, for
	// example.
	//
	// In the end, most feasible regression tests are imperfect. I trust that
	// the assumptions I make here are reasonable and this test provides a good
	// enough check that I didn't break anything, at least good enough for my
	// current needs.

	buf := new(bytes.Buffer)
	Serialize(buf, w.Seed.ToInt64())
	Serialize(buf, w.Player.Health)
	Serialize(buf, w.Player.Pos)
	Serialize(buf, int64(len(w.Enemies)))
	for _, e := range w.Enemies {
		switch e.(type) {
		case *Gremlin:
			Serialize(buf, int64(0))
		case *Hound:
			Serialize(buf, int64(1))
		case *UltraHound:
			Serialize(buf, int64(2))
		case *Pillar:
			Serialize(buf, int64(3))
		case *King:
			Serialize(buf, int64(4))
		case *Question:
			Serialize(buf, int64(5))
		}
		Serialize(buf, e.Health())
		Serialize(buf, e.Pos())
	}
	SerializeSlice(buf, w.Obstacles.ToSlice())
	return HashBytes(buf.Bytes())
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
	obstacles := w.Obstacles.Clone()
	for _, enemy := range w.Enemies {
		obstacles.Set(enemy.Pos())
	}
	w.AttackableTiles = w.vision.Compute(w.Player.Pos, obstacles)
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
