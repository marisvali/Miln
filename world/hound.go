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

func (g *Hound) Clone() Enemy {
	ng := *g
	return &ng
}

func (g *Hound) Vulnerable(w *World) bool {
	return g.state == Searching || g.state == PreparingToAttack || g.state == Attacking
}

func (g *Hound) Step(w *World) {
	var justEnteredState bool
	if !g.solvedFirstState {
		justEnteredState = true
		g.solvedFirstState = true
	} else {
		justEnteredState = g.state != g.previousState
		g.previousState = g.state
	}

	switch g.state {
	case Searching:
		g.searching(justEnteredState, w)
	case PreparingToAttack:
		g.preparingToAttack(justEnteredState, w)
	case Attacking:
		g.attacking(justEnteredState, w)
	case Hit:
		g.hit(justEnteredState, w)
	case Dead:
		g.dead(justEnteredState, w)
	}
}

func (g *Hound) searching(justEnteredState bool, w *World) {
	// Searching means moving around randomly. We only move once in a while,
	// not at every frame. The interval at which we move is a number of frames.
	// That number is computed like this:
	// nFramesInInterval = g.moveCooldownMultiplier * w.EnemyMoveCooldown
	// The way we wait for that interval is that every time the global counter
	// is zero, we tick down our own counter.

	// On entry, reset the "move" countdown and get a new random target to move
	// towards.
	if justEnteredState {
		g.moveCooldownIdx = g.moveCooldownMultiplier
		g.randomTarget = w.Obstacles.RandomUnoccupiedPos(&g.Rand)
	}

	// React to being hit.
	if g.beamJustHit(w) {
		g.health.Dec()
		if g.health.IsZero() {
			g.state = Dead
			return
		} else {
			g.state = Hit
			return
		}
	}

	// If player is visible, prepare to attack.
	if w.Player.OnMap && w.VisibleTiles.At(g.pos) {
		g.state = PreparingToAttack
		return
	}

	// Only move or reduce move tick down every w.EnemyMoveCooldown frames.
	if w.EnemyMoveCooldown.Ready() {
		g.moveCooldownIdx.Dec()

		if g.moveCooldownIdx.IsZero() {
			g.moveRandomly(w)

			// Reset the counter to when we move.
			g.moveCooldownIdx = g.moveCooldownMultiplier
		}
	}
}

func (g *Hound) preparingToAttack(justEnteredState bool, w *World) {
	// On entry, reset the "prepare to attack" countdown.
	if justEnteredState {
		g.preparingToAttackCooldownIdx = g.preparingToAttackCooldown
	}

	// React to being hit.
	if g.beamJustHit(w) {
		g.health.Dec()
		if g.health.IsZero() {
			g.state = Dead
			return
		} else {
			g.state = Hit
			return
		}
	}

	// If player is no longer visible, go back to searching.
	if !w.Player.OnMap || !w.VisibleTiles.At(g.pos) {
		g.state = Searching
		return
	}

	// Tick down counter to when we attack.
	g.preparingToAttackCooldownIdx.Dec()

	// If we have done waiting before attacking, attack.
	if g.preparingToAttackCooldownIdx.IsZero() {
		g.state = Attacking
		return
	}
}

func (g *Hound) attacking(justEnteredState bool, w *World) {
	// Attacking means moving towards the player. We only move once in a while,
	// not at every frame. The interval at which we move is a number of frames.
	// That number is computed like this:
	// nFramesInInterval = g.attackCooldownMultiplier * w.EnemyMoveCooldown
	// The way we wait for that interval is that every time the global counter
	// is zero, we tick down our own counter.

	// On entry, reset the "attack" countdown.
	if justEnteredState {
		g.attackCooldownIdx = g.attackCooldownMultiplier
	}

	// React to being hit.
	if g.beamJustHit(w) {
		g.health.Dec()
		if g.health.IsZero() {
			g.state = Dead
			return
		} else {
			g.state = Hit
			return
		}
	}

	// If player is no longer visible, go back to searching.
	if !w.Player.OnMap || !w.VisibleTiles.At(g.pos) {
		g.state = Searching
		return
	}

	// Only go to player or reduce attack tick down every w.EnemyMoveCooldown
	// frames.
	if w.EnemyMoveCooldown.Ready() {
		// Tick down counter to when we attack.
		g.attackCooldownIdx.Dec()

		if g.attackCooldownIdx.IsZero() {
			// Go to player.
			g.goToPlayer(w, getObstaclesAndEnemies(w))

			// Reset the counter to when we attack.
			g.attackCooldownIdx = g.attackCooldownMultiplier
		}
	}
}

func (g *Hound) hit(justEnteredState bool, w *World) {
	// On entry, reset the "we're hit" countdown.
	if justEnteredState {
		g.hitCooldownIdx = g.hitCooldown
	}

	// Tick down counter to when we move.
	g.hitCooldownIdx.Dec()
	if g.hitCooldownIdx.IsZero() {
		// If player is visible, prepare to attack.
		if w.Player.OnMap && w.VisibleTiles.At(g.pos) {
			g.state = PreparingToAttack
			return
		} else {
			// If not, search.
			g.state = Searching
			return
		}
	}
}

func (g *Hound) dead(justEnteredState bool, w *World) {
	// Do nothing, this is an end state.
	// We should get destroyed/cleaned-up by the world at some point.
}

func (e *Hound) MoveCooldownMultiplier() Int {
	return e.moveCooldownMultiplier
}

func (e *Hound) MoveCooldownIdx() Int {
	return e.moveCooldownIdx
}

func (e *Hound) State() string { return enemyStateName[e.state] }

func (e *Hound) goToPlayer(w *World, m MatBool) {
	ComputePath(e.pos, w.Player.Pos(), m)
	if Path.N > 1 {
		if e.hitsPlayer {
			// Move to the position either way and hit player if necessary.
			e.pos = Path.V[1]
			if Path.V[1].Eq(w.Player.Pos()) {
				w.Player.Hit()
			}
		} else {
			// Move to the position only if not occupied by the player.
			if !Path.V[1].Eq(w.Player.Pos()) {
				e.pos = Path.V[1]
			}
		}
	}
}

func (e *Hound) moveRandomly(w *World) {
	m := getObstaclesAndEnemies(w)
	// Try to move a few times before giving up.
	for i := 0; i < 10; i++ {
		ComputePath(e.pos, e.randomTarget, m)

		if Path.N > 1 {
			// Can go towards the current random target.
			e.pos = Path.V[1]
			return
		} else {
			// For some reason we can't go towards the random target anymore.
			// Maybe we reached it. Maybe it became inaccessible because someone
			// is blocking the way. Either way, get a new random target.
			e.randomTarget = m.RandomUnoccupiedPos(&e.Rand)
		}
	}
}
