package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type HoundState int

const (
	Searching HoundState = iota
	PreparingToAttack
	Attacking
	Hit
	Dead
)

var enemyStateName = map[HoundState]string{
	Searching:         "Searching",
	PreparingToAttack: "PreparingToAttack",
	Attacking:         "Attacking",
	Hit:               "Hit",
	Dead:              "Dead",
}

type Hound struct {
	Rand
	pos                          Pt
	health                       Int
	maxHealth                    Int
	state                        HoundState
	previousState                HoundState
	solvedFirstState             bool
	moveCooldownMultiplier       Int
	moveCooldownIdx              Int
	preparingToAttackCooldown    Int
	preparingToAttackCooldownIdx Int
	attackCooldownMultiplier     Int
	attackCooldownIdx            Int
	hitCooldown                  Int
	hitCooldownIdx               Int
	hitsPlayer                   bool
	aggroDistance                Int
	randomTarget                 Pt
}

func NewHound(seed Int, w WorldParams, pos Pt) Hound {
	var g Hound
	g.RSeed(seed)
	g.pos = pos
	g.maxHealth = w.HoundMaxHealth
	g.health = g.maxHealth
	g.moveCooldownMultiplier = w.HoundMoveCooldownMultiplier
	g.preparingToAttackCooldown = w.HoundPreparingToAttackCooldown
	g.attackCooldownMultiplier = w.HoundAttackCooldownMultiplier
	g.hitCooldown = w.HoundHitCooldownDuration
	g.hitsPlayer = w.HoundHitsPlayer
	g.aggroDistance = w.HoundAggroDistance
	return g
}

func (h *Hound) Vulnerable(w *World) bool {
	return h.state == Searching || h.state == PreparingToAttack || h.state == Attacking
}

func (h *Hound) Step(w *World) {
	var justEnteredState bool
	if !h.solvedFirstState {
		justEnteredState = true
		h.solvedFirstState = true
	} else {
		justEnteredState = h.state != h.previousState
		h.previousState = h.state
	}

	switch h.state {
	case Searching:
		h.searching(justEnteredState, w)
	case PreparingToAttack:
		h.preparingToAttack(justEnteredState, w)
	case Attacking:
		h.attacking(justEnteredState, w)
	case Hit:
		h.hit(justEnteredState, w)
	case Dead:
		h.dead(justEnteredState, w)
	}
}

func (h *Hound) searching(justEnteredState bool, w *World) {
	// Searching means moving around randomly. We only move once in a while,
	// not at every frame. The interval at which we move is a number of frames.
	// That number is computed like this:
	// nFramesInInterval = h.moveCooldownMultiplier * w.EnemyMoveCooldown
	// The way we wait for that interval is that every time the global counter
	// is zero, we tick down our own counter.

	// On entry, reset the "move" countdown and get a new random target to move
	// towards.
	if justEnteredState {
		h.moveCooldownIdx = h.moveCooldownMultiplier
		h.randomTarget = w.Obstacles.RandomUnoccupiedPos(&h.Rand)
	}

	// React to being hit.
	if h.beamJustHit(w) {
		h.health.Dec()
		if h.health.IsZero() {
			h.state = Dead
			return
		} else {
			h.state = Hit
			return
		}
	}

	// If player is visible, prepare to attack.
	if w.Player.OnMap && w.VisibleTiles.At(h.pos) {
		h.state = PreparingToAttack
		return
	}

	// Only move or reduce move tick down every w.EnemyMoveCooldown frames.
	if w.EnemyMoveCooldown.Ready() {
		h.moveCooldownIdx.Dec()

		if h.moveCooldownIdx.IsZero() {
			h.moveRandomly(w)

			// Reset the counter to when we move.
			h.moveCooldownIdx = h.moveCooldownMultiplier
		}
	}
}

func (h *Hound) preparingToAttack(justEnteredState bool, w *World) {
	// On entry, reset the "prepare to attack" countdown.
	if justEnteredState {
		h.preparingToAttackCooldownIdx = h.preparingToAttackCooldown
	}

	// React to being hit.
	if h.beamJustHit(w) {
		h.health.Dec()
		if h.health.IsZero() {
			h.state = Dead
			return
		} else {
			h.state = Hit
			return
		}
	}

	// If player is no longer visible, go back to searching.
	if !w.Player.OnMap || !w.VisibleTiles.At(h.pos) {
		h.state = Searching
		return
	}

	// Tick down counter to when we attack.
	h.preparingToAttackCooldownIdx.Dec()

	// If we have done waiting before attacking, attack.
	if h.preparingToAttackCooldownIdx.IsZero() {
		h.state = Attacking
		return
	}
}

func (h *Hound) attacking(justEnteredState bool, w *World) {
	// Attacking means moving towards the player. We only move once in a while,
	// not at every frame. The interval at which we move is a number of frames.
	// That number is computed like this:
	// nFramesInInterval = h.attackCooldownMultiplier * w.EnemyMoveCooldown
	// The way we wait for that interval is that every time the global counter
	// is zero, we tick down our own counter.

	// On entry, reset the "attack" countdown.
	if justEnteredState {
		h.attackCooldownIdx = h.attackCooldownMultiplier
	}

	// React to being hit.
	if h.beamJustHit(w) {
		h.health.Dec()
		if h.health.IsZero() {
			h.state = Dead
			return
		} else {
			h.state = Hit
			return
		}
	}

	// If player is no longer visible, go back to searching.
	if !w.Player.OnMap || !w.VisibleTiles.At(h.pos) {
		h.state = Searching
		return
	}

	// Only go to player or reduce attack tick down every w.EnemyMoveCooldown
	// frames.
	if w.EnemyMoveCooldown.Ready() {
		// Tick down counter to when we attack.
		h.attackCooldownIdx.Dec()

		if h.attackCooldownIdx.IsZero() {
			// Go to player.
			h.goToPlayer(w, getObstaclesAndEnemies(w))

			// Reset the counter to when we attack.
			h.attackCooldownIdx = h.attackCooldownMultiplier
		}
	}
}

func (h *Hound) hit(justEnteredState bool, w *World) {
	// On entry, reset the "we're hit" countdown.
	if justEnteredState {
		h.hitCooldownIdx = h.hitCooldown
	}

	// Tick down counter to when we move.
	h.hitCooldownIdx.Dec()
	if h.hitCooldownIdx.IsZero() {
		// If player is visible, prepare to attack.
		if w.Player.OnMap && w.VisibleTiles.At(h.pos) {
			h.state = PreparingToAttack
			return
		} else {
			// If not, search.
			h.state = Searching
			return
		}
	}
}

func (h *Hound) dead(justEnteredState bool, w *World) {
	// Do nothing, this is an end state.
	// We should get destroyed/cleaned-up by the world at some point.
}

func (h *Hound) MoveCooldownMultiplier() Int {
	return h.moveCooldownMultiplier
}

func (h *Hound) MoveCooldownIdx() Int {
	return h.moveCooldownIdx
}

func (h *Hound) State() string { return enemyStateName[h.state] }

func (h *Hound) goToPlayer(w *World, m MatBool) {
	path := ComputePath(h.pos, w.Player.Pos(), m)
	if path.N > 1 {
		if h.hitsPlayer {
			// Move to the position either way and hit player if necessary.
			h.pos = path.V[1]
			if path.V[1].Eq(w.Player.Pos()) {
				w.Player.Hit()
			}
		} else {
			// Move to the position only if not occupied by the player.
			if !path.V[1].Eq(w.Player.Pos()) {
				h.pos = path.V[1]
			}
		}
	}
}

func (h *Hound) moveRandomly(w *World) {
	m := getObstaclesAndEnemies(w)
	// Try to move a few times before giving up.
	for i := 0; i < 10; i++ {
		path := ComputePath(h.pos, h.randomTarget, m)

		if path.N > 1 {
			// Can go towards the current random target.
			h.pos = path.V[1]
			return
		} else {
			// For some reason we can't go towards the random target anymore.
			// Maybe we reached it. Maybe it became inaccessible because someone
			// is blocking the way. Either way, get a new random target.
			h.randomTarget = m.RandomUnoccupiedPos(&h.Rand)
		}
	}
}

func (h *Hound) Pos() Pt {
	return h.pos
}

func (h *Hound) Health() Int {
	return h.health
}

func (h *Hound) MaxHealth() Int {
	return h.maxHealth
}

func (h *Hound) Alive() bool {
	return h.health.IsPositive()
}

func getObstaclesAndEnemies(w *World) (m MatBool) {
	m = w.Obstacles
	m.Add(w.EnemyPositions())
	return
}

func (h *Hound) beamJustHit(w *World) bool {
	if !w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		return false
	}
	return w.WorldPosToTile(w.Beam.End) == h.pos
}
