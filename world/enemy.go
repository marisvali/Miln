package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Enemy interface {
	Step(w *World)
	Pos() Pt
	Alive() bool
	FreezeCooldownIdx() Int
	FreezeCooldown() Int
	MoveCooldownMultiplier() Int
	MoveCooldownIdx() Int
	Health() Int
	MaxHealth() Int
	Clone() Enemy
	Vulnerable(w *World) bool
	State() string
}

type EnemyState int

const (
	Searching EnemyState = iota
	PreparingToAttack
	Attacking
	Hit
	Dead
)

var enemyStateName = map[EnemyState]string{
	Searching:         "Searching",
	PreparingToAttack: "PreparingToAttack",
	Attacking:         "Attacking",
	Hit:               "Hit",
	Dead:              "Dead",
}

type EnemyBase struct {
	pos                          Pt
	health                       Int
	maxHealth                    Int
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

func (e *EnemyBase) Pos() Pt {
	return e.pos
}

func (e *EnemyBase) FreezeCooldownIdx() Int {
	return e.hitCooldownIdx
}

func (e *EnemyBase) FreezeCooldown() Int {
	return e.hitCooldown
}

func (e *EnemyBase) MoveCooldownMultiplier() Int {
	return e.moveCooldownMultiplier
}

func (e *EnemyBase) MoveCooldownIdx() Int {
	return e.moveCooldownIdx
}

func (e *EnemyBase) Health() Int {
	return e.health
}

func (e *EnemyBase) MaxHealth() Int {
	return e.maxHealth
}

func (e *EnemyBase) Alive() bool {
	return e.health.IsPositive()
}

func (e *EnemyBase) State() string { return enemyStateName[e.state] }

func (e *EnemyBase) Step(w *World) {
	var justEnteredState bool
	if !e.solvedFirstState {
		justEnteredState = true
		e.solvedFirstState = true
	} else {
		justEnteredState = e.state != e.previousState
		e.previousState = e.state
	}

	switch e.state {
	case Searching:
		e.searching(justEnteredState, w)
	case PreparingToAttack:
		e.preparingToAttack(justEnteredState, w)
	case Attacking:
		e.attacking(justEnteredState, w)
	case Hit:
		e.hit(justEnteredState, w)
	case Dead:
		e.dead(justEnteredState, w)
	}
}

func (e *EnemyBase) searching(justEnteredState bool, w *World) {
	// Searching means moving around randomly. We only move once in a while,
	// not at every frame. The interval at which we move is a number of frames.
	// That number is computed like this:
	// nFramesInInterval = e.moveCooldownMultiplier * w.EnemyMoveCooldown
	// The way we wait for that interval is that every time the global counter
	// is zero, we tick down our own counter.

	// On entry, reset the "move" countdown and get a new random target to move
	// towards.
	if justEnteredState {
		e.moveCooldownIdx = e.moveCooldownMultiplier
		e.randomTarget = w.Obstacles.RandomUnoccupiedPos()
	}

	// React to being hit.
	if e.beamJustHit(w) {
		e.health.Dec()
		if e.health.IsZero() {
			e.state = Dead
			return
		} else {
			e.state = Hit
			return
		}
	}

	// If player is visible, prepare to attack.
	if w.Player.OnMap && w.AttackableTiles.At(e.pos) {
		e.state = PreparingToAttack
		return
	}

	// Only move or reduce move tick down every w.EnemyMoveCooldown frames.
	if w.EnemyMoveCooldownIdx.IsZero() {
		e.moveCooldownIdx.Dec()

		if e.moveCooldownIdx.IsZero() {
			e.moveRandomly(w)

			// Reset the counter to when we move.
			e.moveCooldownIdx = e.moveCooldownMultiplier
		}
	}
}

func (e *EnemyBase) preparingToAttack(justEnteredState bool, w *World) {
	// On entry, reset the "prepare to attack" countdown.
	if justEnteredState {
		e.preparingToAttackCooldownIdx = e.preparingToAttackCooldown
	}

	// React to being hit.
	if e.beamJustHit(w) {
		e.health.Dec()
		if e.health.IsZero() {
			e.state = Dead
			return
		} else {
			e.state = Hit
			return
		}
	}

	// If player is no longer visible, go back to searching.
	if !w.Player.OnMap || !w.AttackableTiles.At(e.pos) {
		e.state = Searching
		return
	}

	// Tick down counter to when we attack.
	e.preparingToAttackCooldownIdx.Dec()

	// If we have done waiting before attacking, attack.
	if e.preparingToAttackCooldownIdx.IsZero() {
		e.state = Attacking
		return
	}
}

func (e *EnemyBase) attacking(justEnteredState bool, w *World) {
	// Attacking means moving towards the player. We only move once in a while,
	// not at every frame. The interval at which we move is a number of frames.
	// That number is computed like this:
	// nFramesInInterval = e.attackCooldownMultiplier * w.EnemyMoveCooldown
	// The way we wait for that interval is that every time the global counter
	// is zero, we tick down our own counter.

	// On entry, reset the "attack" countdown.
	if justEnteredState {
		e.attackCooldownIdx = e.attackCooldownMultiplier
	}

	// React to being hit.
	if e.beamJustHit(w) {
		e.health.Dec()
		if e.health.IsZero() {
			e.state = Dead
			return
		} else {
			e.state = Hit
			return
		}
	}

	// If player is no longer visible, go back to searching.
	if !w.Player.OnMap || !w.AttackableTiles.At(e.pos) {
		e.state = Searching
		return
	}

	// Only go to player or reduce attack tick down every w.EnemyMoveCooldown
	// frames.
	if w.EnemyMoveCooldownIdx.IsZero() {
		// Tick down counter to when we attack.
		e.attackCooldownIdx.Dec()

		if e.attackCooldownIdx.IsZero() {
			// Go to player.
			e.goToPlayer(w, getObstaclesAndEnemies(w))

			// Reset the counter to when we attack.
			e.attackCooldownIdx = e.attackCooldownMultiplier
		}
	}
}

func (e *EnemyBase) hit(justEnteredState bool, w *World) {
	// On entry, reset the "we're hit" countdown.
	if justEnteredState {
		e.hitCooldownIdx = e.hitCooldown
	}

	// Tick down counter to when we move.
	e.hitCooldownIdx.Dec()
	if e.hitCooldownIdx.IsZero() {
		// If player is visible, prepare to attack.
		if w.Player.OnMap && w.AttackableTiles.At(e.pos) {
			e.state = PreparingToAttack
			return
		} else {
			// If not, search.
			e.state = Searching
			return
		}
	}
}

func (e *EnemyBase) dead(justEnteredState bool, w *World) {
	// Do nothing, this is an end state.
	// We should get destroyed/cleaned-up by the world at some point.
}

func (e *EnemyBase) goToPlayer(w *World, m MatBool) {
	path := FindPath(e.pos, w.Player.Pos(), m.Matrix, false)
	if len(path) > 1 {
		if e.hitsPlayer {
			// Move to the position either way and hit player if necessary.
			e.pos = path[1]
			if path[1].Eq(w.Player.Pos()) {
				w.Player.Hit()
			}
		} else {
			// Move to the position only if not occupied by the player.
			if !path[1].Eq(w.Player.Pos()) {
				e.pos = path[1]
			}
		}
	}
}

func (e *EnemyBase) moveRandomly(w *World) {
	m := getObstaclesAndEnemies(w)
	// Try to move a few times before giving up.
	for i := 0; i < 10; i++ {
		path := FindPath(e.pos, e.randomTarget, m.Matrix, false)
		if len(path) > 1 {
			// Can go towards the current random target.
			e.pos = path[1]
			return
		} else {
			// For some reason we can't go towards the random target anymore.
			// Maybe we reached it. Maybe it became inaccessible because someone
			// is blocking the way. Either way, get a new random target.
			e.randomTarget = m.RandomUnoccupiedPos()
		}
	}
}

func (e *EnemyBase) Vulnerable(w *World) bool {
	return e.state == Searching || e.state == PreparingToAttack || e.state == Attacking
}

func getObstaclesAndEnemies(w *World) (m MatBool) {
	m = w.Obstacles.Clone()
	m.Add(w.EnemyPositions())
	return
}

func (e *EnemyBase) beamJustHit(w *World) bool {
	if !w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		return false
	}
	return w.WorldPosToTile(w.Beam.End) == e.pos
}
