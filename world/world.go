package world

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	. "github.com/marisvali/miln/gamelib"
	"math"
	"slices"
)

type Beam struct {
	Idx Int // if this is greater than 0 it means the beam is active for Idx time steps
	End Pt  // this is the point to where the beam ends
}

const Version = 16

type WorldParams struct {
	Boardgame                      bool `yaml:"Boardgame"`
	UseAmmo                        bool `yaml:"UseAmmo"`
	AmmoLimit                      Int  `yaml:"AmmoLimit"`
	EnemyMoveCooldownDuration      Int  `yaml:"EnemyMoveCooldownDuration"`
	EnemiesAggroWhenVisible        bool `yaml:"EnemiesAggroWhenVisible"`
	SpawnPortalCooldownMin         Int  `yaml:"SpawnPortalCooldownMin"`
	SpawnPortalCooldownMax         Int  `yaml:"SpawnPortalCooldownMax"`
	HoundMaxHealth                 Int  `yaml:"HoundMaxHealth"`
	HoundMoveCooldownMultiplier    Int  `yaml:"HoundMoveCooldownMultiplier"`
	HoundPreparingToAttackCooldown Int  `yaml:"HoundPreparingToAttackCooldown"`
	HoundAttackCooldownMultiplier  Int  `yaml:"HoundAttackCooldownMultiplier"`
	HoundHitCooldownDuration       Int  `yaml:"HoundHitCooldownDuration"`
	HoundHitsPlayer                bool `yaml:"HoundHitsPlayer"`
	HoundAggroDistance             Int  `yaml:"HoundAggroDistance"`
}

type WorldObject interface {
	Pos() Pt
	State() string
}

type World struct {
	Playthrough
	Rand
	Player            Player
	Enemies           []Enemy
	Beam              Beam
	VisibleTiles      MatBool
	TimeStep          Int
	BeamMax           Int
	BlockSize         Int
	EnemyMoveCooldown Cooldown
	Ammos             []Ammo
	SpawnPortals      []SpawnPortal
	vision            Vision
}

func (w *World) Clone() World {
	// Copy all non-slice variables.
	// Does shallow copies of slices as well, but those should be overwritten
	// below by deep copies.
	clone := *w

	// Do deep copies for slices.
	clone.History = slices.Clone(w.History)

	clone.Level = w.Level.Clone()

	clone.Enemies = []Enemy{}
	for i := range w.Enemies {
		clone.Enemies = append(clone.Enemies, w.Enemies[i].Clone())
	}

	clone.VisibleTiles = w.VisibleTiles.Clone()

	clone.Ammos = slices.Clone(w.Ammos)

	clone.SpawnPortals = []SpawnPortal{}
	for i := range w.SpawnPortals {
		clone.SpawnPortals = append(clone.SpawnPortals, w.SpawnPortals[i].Clone())
	}

	clone.vision = NewVision(w.Obstacles.Size())
	return clone
}

type Playthrough struct {
	Level
	Id      uuid.UUID
	Seed    Int
	History []PlayerInput
}

type PlayerInput struct {
	MousePt            Pt
	LeftButtonPressed  bool
	RightButtonPressed bool
	Move               bool
	MovePt             Pt // tile-coordinates
	Shoot              bool
	ShootPt            Pt // tile-coordinates
}

type SpawnPortalParams struct {
	Pos                 Pt     `yaml:"Pos"`
	SpawnPortalCooldown Int    `yaml:"SpawnPortalCooldown"`
	Waves               []Wave `yaml:"Waves"`
}

type Entities struct {
	Obstacles          MatBool             `yaml:"Obstacles"`
	SpawnPortalsParams []SpawnPortalParams `yaml:"SpawnPortalsParams"`
}

type Level struct {
	Entities    `yaml:"Entities"`
	WorldParams `yaml:"WorldParams"`
}

func (l *Level) Clone() Level {
	// Copy all non-slice variables.
	// Does shallow copies of slices as well, but those should be overwritten
	// below by deep copies.
	clone := *l

	// Do deep copies for slices.
	clone.Obstacles = l.Obstacles.Clone()
	clone.SpawnPortalsParams = []SpawnPortalParams{}
	for _, spp := range l.SpawnPortalsParams {
		clone.SpawnPortalsParams = append(clone.SpawnPortalsParams,
			SpawnPortalParams{spp.Pos, spp.SpawnPortalCooldown,
				slices.Clone(spp.Waves)})
	}
	return clone
}

type LevelYaml struct {
	Version Int `yaml:"Version"`
	Seed    Int `yaml:"Seed"`
	Level   `yaml:"Level"`
}

type VersionYaml struct {
	Version Int `yaml:"Version"`
}

func (l *Level) SaveToYAML(seed Int, filename string) {
	var lYaml LevelYaml
	lYaml.Version = I(Version)
	lYaml.Seed = seed
	lYaml.Level = *l
	SaveYAML(filename, lYaml)
}

func LoadLevelFromYAML(fsys FS, filename string) (seed Int, l Level) {
	var vYaml VersionYaml
	LoadYAML(fsys, filename, &vYaml)
	if vYaml.Version.ToInt64() != Version {
		Check(fmt.Errorf("this code can't simulate this playthrough "+
			"correctly - we are version %d and playthrough was generated "+
			"with version %d",
			Version, vYaml.Version.ToInt64()))
	}

	var lYaml LevelYaml
	LoadYAML(fsys, filename, &lYaml)
	return lYaml.Seed, lYaml.Level
}

func IsYamlLevel(filename string) bool {
	b := ReadFile(filename)
	versionS := "Version"
	if len(b) <= len(versionS) {
		return false
	}
	isYamlLevel := string(b[0:len(versionS)]) == versionS
	return isYamlLevel
}

func NewWorld(seed Int, l Level) (w World) {
	w.Level = l
	w.Seed = seed
	w.RSeed(w.Seed)
	w.Id = uuid.New()
	for _, spParams := range l.SpawnPortalsParams {
		w.SpawnPortals = append(w.SpawnPortals,
			NewSpawnPortal(w.RInt63(), spParams, w.WorldParams))
	}
	w.vision = NewVision(w.Obstacles.Size())

	// Params
	w.BlockSize = I(1000)
	w.BeamMax = I(15)
	w.Player = NewPlayer()
	w.Player.AmmoLimit = w.AmmoLimit
	w.EnemyMoveCooldown = NewCooldown(w.EnemyMoveCooldownDuration)

	// GUI needs this even without the world ever doing a step.
	// Note: this was true when the player started on the map, so it might not
	// be relevant now that the player doesn't start on the map. But, keep it
	// in case things change again.
	w.computeVisibleTiles()
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
	Serialize(buf, w.Player.Pos())
	Serialize(buf, int64(len(w.Enemies)))
	for _, e := range w.Enemies {
		switch e.(type) {
		case *Hound:
			Serialize(buf, int64(0))
		}
		Serialize(buf, e.Health())
		Serialize(buf, e.Pos())
	}
	Serialize(buf, w.Obstacles.ToSlice())
	return HashBytes(buf.Bytes())
}

func (w *World) SerializedPlaythrough() []byte {
	buf := new(bytes.Buffer)
	Serialize(buf, int64(Version))
	Serialize(buf, w.Playthrough)
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
	Deserialize(buf, &p)
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

func (w *World) computeVisibleTiles() {
	// Compute which tiles are visible.
	obstacles := w.Obstacles.Clone()
	for _, enemy := range w.Enemies {
		obstacles.Set(enemy.Pos())
	}
	w.VisibleTiles = w.vision.Compute(w.Player.Pos(), obstacles)
}

func (w *World) EnemyPositions() (m MatBool) {
	m = NewMatBool(w.Obstacles.Size())
	for i := range w.Enemies {
		m.Set(w.Enemies[i].Pos())
	}
	return
}

func (w *World) VulnerableEnemyPositions() (m MatBool) {
	m = NewMatBool(w.Obstacles.Size())
	for i := range w.Enemies {
		if w.Enemies[i].Vulnerable(w) {
			m.Set(w.Enemies[i].Pos())
		}
	}
	return
}

func (w *World) SpawnPortalPositions() (m MatBool) {
	m = NewMatBool(w.Obstacles.Size())
	for i := range w.SpawnPortals {
		m.Set(w.SpawnPortals[i].pos)
	}
	return
}

func (w *World) Step(input PlayerInput) {
	w.History = append(w.History, input)
	w.Player.Step(w, input)
	w.computeVisibleTiles()

	stepEnemies := true
	if w.Boardgame && !input.Move && !input.Shoot {
		stepEnemies = false
	}

	if stepEnemies {
		// Step the ammos.
		if w.UseAmmo {
			w.SpawnAmmos()
		}

		w.EnemyMoveCooldown.Update()

		// Step the enemies.
		for i := range w.Enemies {
			w.Enemies[i].Step(w)
		}

		// Step SpawnPortalsParams.
		for i := range w.SpawnPortals {
			w.SpawnPortals[i].Step(w)
		}

		if w.EnemyMoveCooldown.Ready() {
			w.EnemyMoveCooldown.Reset()
		}
	}

	// Cull dead enemies.
	newEnemies := []Enemy{}
	for i := range w.Enemies {
		if w.Enemies[i].Alive() {
			newEnemies = append(newEnemies, w.Enemies[i])
		}
	}
	w.Enemies = newEnemies

	// Cull dead SpawnPortalsParams.
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

func (w *World) SpawnAmmos() {
	// Spawn new ammos
	for {
		// Count ammo available in the world now.
		available := ZERO
		for _, a := range w.Ammos {
			available.Add(a.Count)
		}

		// Count total ammo.
		totalAmmo := w.Player.AmmoCount.Plus(available)

		if totalAmmo.Geq(w.Player.AmmoLimit) {
			// There's enough ammo available.
			return
		}

		// Build matrix with positions occupied everywhere we don't want to
		// spawn ammo.
		occ := w.Obstacles.Clone()
		for _, a := range w.Ammos {
			occ.Set(a.Pos)
		}
		occ.Set(w.Player.Pos())
		for _, e := range w.Enemies {
			occ.Set(e.Pos())
		}

		// Spawn ammo.
		ammo := Ammo{
			Pos:   occ.OccupyRandomPos(&w.Rand),
			Count: I(3),
		}
		w.Ammos = append(w.Ammos, ammo)
	}
}
