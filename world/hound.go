package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Hound struct {
	Rand
	EnemyBase
	state                        EnemyState
	previousState                EnemyState
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

func NewHound(seed Int, w WorldParams, pos Pt) *Hound {
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
	return &g
}

func (h *Hound) Clone() Enemy {
	ng := *h
	return &ng
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
		h.targetPos = h.getNextPositionTowardsRandomTarget(w)
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

	// Only move or reduce move tick down every w.EnemyMoveCooldown frames.
	if w.EnemyMoveCooldown.Ready() {
		h.moveCooldownIdx.Dec()

		if h.moveCooldownIdx.IsZero() {
			h.moveTo(h.targetPos, w)
			h.targetPos = h.getNextPositionTowardsRandomTarget(w)

			// If player is visible, prepare to attack.
			if w.Player.OnMap && w.VisibleTiles.At(h.pos) {
				h.state = Attacking
				return
			}

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

		// Choose target.
		h.targetPos = h.getNextPositionTowardsPlayer(w, getObstaclesAndEnemies(w))
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

	// Only go to player or reduce attack tick down every w.EnemyMoveCooldown
	// frames.
	if w.EnemyMoveCooldown.Ready() {
		// Tick down counter to when we attack.
		h.attackCooldownIdx.Dec()

		if h.attackCooldownIdx.IsZero() {
			// Go to player.
			h.moveTo(h.targetPos, w)
			h.targetPos = h.getNextPositionTowardsPlayer(w, getObstaclesAndEnemies(w))

			// If player is no longer visible, go back to searching.
			if !w.Player.OnMap || !w.VisibleTiles.At(h.pos) {
				h.state = Searching
				return
			}

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
	if h.hitCooldownIdx.Leq(ZERO) {
		if w.EnemyMoveCooldown.Ready() {
			// If player is visible, prepare to attack.
			if w.Player.OnMap && w.VisibleTiles.At(h.pos) {
				h.state = Attacking
				return
			} else {
				// If not, search.
				h.state = Searching
				return
			}
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

func (h *Hound) getNextPositionTowardsPlayer(w *World, m MatBool) Pt {
	path := FindPath(h.pos, w.Player.Pos(), m.Matrix, false)
	if len(path) > 1 {
		return path[1]
	} else {
		return h.pos
	}
}

func (h *Hound) moveTo(pos Pt, w *World) {
	if h.hitsPlayer {
		// Move to the position either way and hit player if necessary.
		h.pos = pos
		if pos.Eq(w.Player.Pos()) {
			w.Player.Hit()
		}
	} else {
		// Move to the position only if not occupied by the player.
		if !pos.Eq(w.Player.Pos()) {
			h.pos = pos
		}
	}
}

func (h *Hound) getNextPositionTowardsRandomTarget(w *World) Pt {
	m := getObstaclesAndEnemies(w)

	// Try to move a few times before giving up.
	for i := 0; i < 10; i++ {
		path := FindPath(h.pos, h.randomTarget, m.Matrix, false)
		if len(path) > 1 {
			// Can go towards the current random target.
			return path[1]
		} else {
			// For some reason we can't go towards the random target anymore.
			// Maybe we reached it. Maybe it became inaccessible because someone
			// is blocking the way. Either way, get a new random target.
			h.randomTarget = m.RandomUnoccupiedPos(&h.Rand)
		}
	}
	return h.pos
}
