package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	"github.com/marisvali/miln/world/oldworld"
	"math"
)

const Version = 999

type World struct {
	WorldDebugInfo
	Rand
	WorldParams
	Obstacles         MatBool
	Player            Player
	Enemies           EnemiesArray
	Beam              Beam
	VisibleTiles      MatBool
	TimeStep          Int
	BeamMax           Int
	BlockSize         Int
	EnemyMoveCooldown Cooldown
	Ammos             AmmosArray
	SpawnPortals      SpawnPortalsArray
	vision            Vision
}

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

type Beam struct {
	Idx Int // if this is greater than 0 it means the beam is active for Idx time steps
	End Pt  // this is the point to where the beam ends
}

type WorldObject interface {
	Pos() Pt
	State() string
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

func NewWorld(seed Int, l Level) (w World) {
	// Initialize values.
	w.Obstacles = l.Obstacles
	w.WorldParams = l.WorldParams
	w.RSeed(seed)
	for i := range l.SpawnPortalsParams.N {
		w.SpawnPortals.Data[i] = NewSpawnPortal(w.RInt63(), l.SpawnPortalsParams.Data[i], w.WorldParams)
	}
	w.SpawnPortals.N = l.SpawnPortalsParams.N
	w.vision = NewVision()

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
	obstacles := w.Obstacles
	for i := range w.Enemies.N {
		obstacles.Set(w.Enemies.Data[i].Pos())
	}
	w.VisibleTiles = w.vision.Compute(w.Player.Pos(), obstacles)
}

func (w *World) EnemyPositions() (m MatBool) {
	for i := range w.Enemies.N {
		m.Set(w.Enemies.Data[i].Pos())
	}
	return
}

func (w *World) VulnerableEnemyPositions() (m MatBool) {
	for i := range w.Enemies.N {
		if w.Enemies.Data[i].Vulnerable(w) {
			m.Set(w.Enemies.Data[i].Pos())
		}
	}
	return
}

func (w *World) SpawnPortalPositions() (m MatBool) {
	for i := range w.SpawnPortals.N {
		m.Set(w.SpawnPortals.Data[i].pos)
	}
	return
}

func (w *World) Step(input PlayerInput) {
	w.StepDebug(input)

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
		for i := range w.Enemies.N {
			w.Enemies.Data[i].Step(w)
		}

		// Step SpawnPortalsParams.
		for i := range w.SpawnPortals.N {
			w.SpawnPortals.Data[i].Step(w)
		}

		if w.EnemyMoveCooldown.Ready() {
			w.EnemyMoveCooldown.Reset()
		}
	}

	// Cull dead enemies.
	n := int64(0)
	for i := range w.Enemies.N {
		if w.Enemies.Data[i].Alive() {
			w.Enemies.Data[n] = w.Enemies.Data[i]
			n++
		}
	}
	w.Enemies.N = n

	// Cull dead SpawnPortals.
	n = int64(0)
	for i := range w.SpawnPortals.N {
		if w.SpawnPortals.Data[i].Health.IsPositive() {
			w.SpawnPortals.Data[n] = w.SpawnPortals.Data[i]
			n++
		}
	}
	w.SpawnPortals.N = n

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
		for i := range w.Ammos.N {
			available.Add(w.Ammos.Data[i].Count)
		}

		// Count total ammo.
		totalAmmo := w.Player.AmmoCount.Plus(available)

		if totalAmmo.Geq(w.Player.AmmoLimit) {
			// There's enough ammo available.
			return
		}

		// Build matrix with positions occupied everywhere we don't want to
		// spawn ammo.
		occ := w.Obstacles
		for i := range w.Ammos.N {
			occ.Set(w.Ammos.Data[i].Pos)
		}
		occ.Set(w.Player.Pos())
		for i := range w.Enemies.N {
			occ.Set(w.Enemies.Data[i].Pos())
		}

		// Spawn ammo.
		ammo := Ammo{
			Pos:   occ.OccupyRandomPos(&w.Rand),
			Count: I(3),
		}
		w.Ammos.Data[w.Ammos.N] = ammo
		w.Ammos.N++
	}
}

func (w *World) AllEnemiesDead() bool {
	for i := range w.Enemies.N {
		if w.Enemies.Data[i].Alive() {
			return false
		}
	}
	for i := range w.SpawnPortals.N {
		if w.SpawnPortals.Data[i].Active() {
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

func OldPt(pt oldworld.Pt) Pt {
	return IPt(pt.X.ToInt(), pt.Y.ToInt())
}

func DeserializePlaythroughFromOld(data []byte) (p Playthrough) {
	po := oldworld.DeserializePlaythrough(data)
	// Entities
	for y := 0; y < po.Obstacles.Size().X.ToInt(); y++ {
		for x := 0; x < po.Obstacles.Size().X.ToInt(); x++ {
			if po.Obstacles.Get(oldworld.Pt{oldworld.I(x), oldworld.I(y)}) {
				p.Obstacles.Set(IPt(x, y))
			}

		}
	}
	for i := range po.SpawnPortalsParams {
		p.SpawnPortalsParams.Data[i].Pos = IPt(po.SpawnPortalsParams[i].Pos.X.ToInt(), po.SpawnPortalsParams[i].Pos.Y.ToInt())
		p.SpawnPortalsParams.Data[i].SpawnPortalCooldown = I(po.SpawnPortalsParams[i].SpawnPortalCooldown.ToInt())
		for j := range po.SpawnPortalsParams[i].Waves {
			p.SpawnPortalsParams.Data[i].Waves.Data[j].SecondsAfterLastWave = I(po.SpawnPortalsParams[i].Waves[j].SecondsAfterLastWave.ToInt())
			p.SpawnPortalsParams.Data[i].Waves.Data[j].NHounds = I(po.SpawnPortalsParams[i].Waves[j].NHounds.ToInt())
		}
		p.SpawnPortalsParams.Data[i].Waves.N = int64(len(po.SpawnPortalsParams[i].Waves))
	}
	p.SpawnPortalsParams.N = int64(len(po.SpawnPortalsParams))

	// WorldParams
	p.Boardgame = po.Boardgame
	p.UseAmmo = po.UseAmmo
	p.AmmoLimit = I(po.AmmoLimit.ToInt())
	p.EnemyMoveCooldownDuration = I(po.EnemyMoveCooldownDuration.ToInt())
	p.EnemiesAggroWhenVisible = po.EnemiesAggroWhenVisible
	p.SpawnPortalCooldownMin = I(po.SpawnPortalCooldownMin.ToInt())
	p.SpawnPortalCooldownMax = I(po.SpawnPortalCooldownMax.ToInt())
	p.HoundMaxHealth = I(po.HoundMaxHealth.ToInt())
	p.HoundMoveCooldownMultiplier = I(po.HoundMoveCooldownMultiplier.ToInt())
	p.HoundPreparingToAttackCooldown = I(po.HoundPreparingToAttackCooldown.ToInt())
	p.HoundAttackCooldownMultiplier = I(po.HoundAttackCooldownMultiplier.ToInt())
	p.HoundHitCooldownDuration = I(po.HoundHitCooldownDuration.ToInt())
	p.HoundHitsPlayer = po.HoundHitsPlayer
	p.HoundAggroDistance = I(po.HoundAggroDistance.ToInt())

	// Rest of Playthrough
	p.Id = po.Id
	p.Seed = I(po.Seed.ToInt())
	p.History = make([]PlayerInput, len(po.History))
	for i := range po.History {
		p.History[i].MousePt = OldPt(po.History[i].MousePt)
		p.History[i].LeftButtonPressed = po.History[i].LeftButtonPressed
		p.History[i].RightButtonPressed = po.History[i].RightButtonPressed
		p.History[i].Move = po.History[i].Move
		p.History[i].MovePt = OldPt(po.History[i].MovePt)
		p.History[i].Shoot = po.History[i].Shoot
		p.History[i].ShootPt = OldPt(po.History[i].ShootPt)
	}
	return
}
