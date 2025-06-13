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
	if token != Version && !(token == 13 && Version == 16) {
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
	// Reset the player's state at the beginning.
	// I don't want to put this inside the Player.Step because if I ever want
	// to step the enemies first and then the player, that will cause an issue.
	// The player doesn't know when others will call Hit() on it. The decision
	// of when to reset this state is a decision that only the coordinating
	// parent (e.g. the World) can take.
	w.Player.JustHit = false
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
	n := 0
	for _, x := range w.Enemies {
		if x.Alive() {
			w.Enemies[n] = x
			n++
		}
	}
	w.Enemies = w.Enemies[:n]

	// Cull dead SpawnPortals.
	n = 0
	for _, x := range w.SpawnPortals {
		if x.Health.IsPositive() {
			w.SpawnPortals[n] = x
			n++
		}
	}

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

func (w *World) AllEnemiesDead() bool {
	for _, enemy := range w.Enemies {
		if enemy.Alive() {
			return false
		}
	}
	for _, portal := range w.SpawnPortals {
		if portal.Active() {
			return false
		}
	}
	return true
}

type WorldStatus int

const (
	Ongoing WorldStatus = iota
	Won
	Lost
)

func (w *World) Status() WorldStatus {
	if w.AllEnemiesDead() {
		return Won
	} else if w.Player.Health.Leq(ZERO) {
		return Lost
	} else {
		return Ongoing
	}
}
