package ai

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

// FitnessOfMoveAction returns the fitness of a specific move action at a
// certain moment. The action is "the player moves to pos".
func FitnessOfMoveAction(world *World, pos Pt) int64 {
	if !ValidMove(world, pos) {
		// If the action isn't even valid, the fitness of the action is zero.
		return 0
	}
	// The action is valid so other factors come into play.

	// Compute how safe the position is.
	safetyFitness := int64(0)
	framesUntilAttacked := NumFramesUntilAttacked(*world, pos)
	// Use time instead of frames because I have an easier time understanding
	// how dangerous a position feels based on how much time it takes for it
	// to be attacked. For example, I know the average for a playthrough was
	// 1 click every 0.6 seconds.
	timeUntilAttacked := float32(framesUntilAttacked) / 60.0
	if timeUntilAttacked <= 0.05 {
		safetyFitness = 0
	}
	if timeUntilAttacked <= 0.3 {
		safetyFitness = 0
	}
	if timeUntilAttacked > 0.3 && timeUntilAttacked <= 0.6 {
		safetyFitness = 5
	}
	if timeUntilAttacked > 0.6 && timeUntilAttacked <= 1 {
		safetyFitness = 10
	}
	if timeUntilAttacked > 1 && timeUntilAttacked <= 2 {
		safetyFitness = 15
	}
	if timeUntilAttacked > 2 {
		safetyFitness = 20
	}

	if safetyFitness == 0 {
		// If the position is too dangerous, the fitness is zero.
		return 0
	}

	ammoFitness := int64(0)
	currentAmmo := world.Player.AmmoCount.ToInt()
	numMovesToAmmo := NumMovesToAmmo(world, pos)
	if currentAmmo == 0 {
		if numMovesToAmmo == 0 {
			ammoFitness = 10
		}
		if numMovesToAmmo == 1 {
			ammoFitness = 5
		}
		if numMovesToAmmo == 2 {
			ammoFitness = 2
		}
	}

	if currentAmmo > 1 && currentAmmo <= 3 {
		if numMovesToAmmo == 0 {
			ammoFitness = 6
		}
		if numMovesToAmmo == 1 {
			ammoFitness = 2
		}
		if numMovesToAmmo == 2 {
			ammoFitness = 1
		}
	}

	if currentAmmo > 3 {
		if numMovesToAmmo == 0 {
			ammoFitness = 2
		}
		if numMovesToAmmo == 1 {
			ammoFitness = 1
		}
	}

	enemyFitness := int64(0)
	movesToVisibleEnemy := NumMovesToEnemy(world, pos)
	if currentAmmo > 0 {
		if movesToVisibleEnemy == 0 {
			enemyFitness = 10
		}
		if movesToVisibleEnemy == 1 {
			enemyFitness = 5
		}
		if movesToVisibleEnemy == 2 {
			enemyFitness = 2
		}
	}

	// if there is some safety, then the other factors come into play
	return safetyFitness + ammoFitness + enemyFitness
}

// FitnessOfAttackAction returns the fitness of a specific attack action at a
// certain moment. The action is "the player attacks pos".
func FitnessOfAttackAction(w World, pos Pt) int64 {
	if !ValidAttack(&w, pos) {
		// If the action isn't even valid, the fitness of the action is zero.
		return 0
	}
	// The action is valid so other factors come into play.

	// there's some extra usefulness if the enemy will finally die with this shot
	// also there's some usefulness in paralyzing the enemy for a while
	// weeurjiuefiuergeugriuegr etc etc cleanup comments etc
	input := PlayerInput{}
	input.Shoot = true
	input.ShootPt = pos
	w.Step(input)

	// The player might have just been attacked and hit.
	if w.Player.JustHit {
		return 0
	}

	if w.Enemies.N == 0 {
		// Just won the game.
		return 1000
	}

	// Compute how safe the current position is.
	safetyFitness := int64(0)
	framesUntilAttacked := NumFramesUntilAttacked(w, w.Player.Pos())
	// Use time instead of frames because I have an easier time understanding
	// how dangerous a position feels based on how much time it takes for it
	// to be attacked. For example, I know the average for a playthrough was
	// 1 click every 0.6 seconds.
	timeUntilAttacked := float32(framesUntilAttacked) / 60.0
	if timeUntilAttacked <= 0.3 {
		safetyFitness = 0
	}
	if timeUntilAttacked > 0.3 && timeUntilAttacked <= 0.6 {
		safetyFitness = 5
	}
	if timeUntilAttacked > 0.6 && timeUntilAttacked <= 1 {
		safetyFitness = 10
	}
	if timeUntilAttacked > 1 && timeUntilAttacked <= 2 {
		safetyFitness = 15
	}
	if timeUntilAttacked > 2 {
		safetyFitness = 20 + int64(timeUntilAttacked*3)
	}

	if safetyFitness < 10 {
		// If the current position is too dangerous, the fitness is zero.
		return 0
	}

	// Prefer attacking to moving, if it's safe and possible.
	biasForAttacking := int64(20)

	return safetyFitness + biasForAttacking
}
