package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Question struct {
	EnemyBase
}

func NewQuestion(pos Pt) *Question {
	var q Question
	q.pos = pos
	q.maxHealth = QuestionMaxHealth
	q.health = q.maxHealth
	q.freezeCooldown = I(1)
	return &q
}

func (q *Question) onDeath(w *World) {
	nQuestions := ZERO
	for i := range w.Enemies {
		if _, ok := w.Enemies[i].(*Question); ok {
			nQuestions.Inc()
		}
	}
	if nQuestions == ONE {
		// w.SpawnPortals = append(w.SpawnPortals, NewSpawnPortal(q.pos))
	} else if nQuestions == TWO {
		w.Keys = append(w.Keys, NewPillarKey(q.pos))
	} else {
		// question mark
		nGremlins := ZERO
		for i := range w.Enemies {
			if _, ok := w.Enemies[i].(*Gremlin); ok {
				nGremlins.Inc()
			}
		}
		if nQuestions.Mod(I(4)) == ZERO && nGremlins.Leq(I(4)) {
			nHounds := ZERO
			for i := range w.Enemies {
				if _, ok := w.Enemies[i].(*UltraHound); ok {
					nHounds.Inc()
				}
			}
			if nHounds.Lt(ONE) && (RInt(I(0), I(100)).Leq(I(40)) || nQuestions.Leq(I(4))) {
				w.Enemies = append(w.Enemies, NewUltraHound(q.pos))
			} else {
				w.Enemies = append(w.Enemies, NewPillar(q.pos))
			}
		} else {
			w.Obstacles.Set(q.pos)
		}
	}
}

func (q *Question) Step(w *World) {
	if q.beamJustHit(w) {
		q.freezeCooldownIdx = q.freezeCooldown
		if w.Player.HitPermissions.CanHitQuestion {
			q.health.Dec()
			if q.health == ZERO {
				q.onDeath(w)
			}
		}
	}
}
