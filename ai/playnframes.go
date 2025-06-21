package ai

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

// PlayLevelForAtLeastNFrames will play a level for at least nFrames. It will
// attempt to generate a playthrough as close to nFrames as it can.
// It is useful to generate large playthroughs automatically in order to
// benchmark the performance of the simulation or to use in regression tests.
// However, the playthrough must be "interesting". It's easy to just not do the
// first move for n frames, but that isn't a good representation of the
// simulation, it doesn't trigger enough of the simulation code.
// The player should move around the map, not lose, but not win either.
// The way PlayLevelForAtLeastNFrames does it is that the player:
// - waits a little
// - appears on the map and waits until it gets hit
// - moves and kills enemies until there are only 3 left
// - moves around the map without attacking enemies, until all n frames are done
// - waits until it gets hit
// - moves and kills the rest of the enemies
// This works because the player moves each 20 frames which is quite fast.
// The enemies must move at a reasonable speed and it helps if there are some
// obstacles but not too many.
func PlayLevelForAtLeastNFrames(l Level, seed Int, nFrames int) (p Playthrough) {
	p.Seed = seed
	p.Level = l
	w := NewWorld(seed, l)

	// Wait some period in the beginning.
	frameIdx := 0
	for ; frameIdx < 100; frameIdx++ {
		Step(&p, &w, NeutralInput())
	}

	var rankedActions ActionsArray

	// Move on the map.
	{
		ComputeRankedActions(w, &rankedActions)
		Step(&p, &w, ActionToInput(rankedActions.V[0]))
	}

	// Fight until only 3 enemy is left.
	for {
		if frameIdx%20 == 0 {
			ComputeRankedActions(w, &rankedActions)
			Step(&p, &w, ActionToInput(rankedActions.V[0]))

			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return
			}
			if w.Enemies.N == 3 {
				break
			}
		} else {
			Step(&p, &w, NeutralInput())
			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return
			}
		}
		frameIdx++
	}

	// Wait until getting hit once.
	for {
		Step(&p, &w, NeutralInput())

		// After each world step, check if the game is over.
		if w.Status() != Ongoing {
			return
		}
		if w.Player.JustHit {
			break
		}
	}

	// Move around without attacking.
	for ; frameIdx < nFrames; frameIdx++ {
		if frameIdx%20 == 0 {
			ComputeRankedActions(w, &rankedActions)
			for i := range rankedActions.N {
				if rankedActions.V[i].Move {
					Step(&p, &w, ActionToInput(rankedActions.V[i]))

					// After each world step, check if the game is over.
					if w.Status() != Ongoing {
						return
					}
					break
				}
			}
		} else {
			Step(&p, &w, NeutralInput())

			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return
			}
		}
	}

	// Wait until getting hit once.
	for {
		Step(&p, &w, NeutralInput())

		// After each world step, check if the game is over.
		if w.Status() != Ongoing {
			return
		}
		if w.Player.JustHit {
			break
		}
	}

	// Fight until all enemies are killed.
	for {
		if frameIdx%20 == 0 {
			ComputeRankedActions(w, &rankedActions)
			Step(&p, &w, ActionToInput(rankedActions.V[0]))

			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return
			}
		} else {
			Step(&p, &w, NeutralInput())

			// After each world step, check if the game is over.
			if w.Status() != Ongoing {
				return
			}
		}
		frameIdx++
	}
}
