package gamelib

import "slices"

type PathArray struct {
	N int64
	V [NCols * NRows]Pt
}

func ComputePath(startPt, endPt Pt, m MatBool) (path PathArray) {
	const NDirs = 8

	type queueArray struct {
		N int64
		V [NCols * NRows]int64
	}

	var neighbors [NCols * NRows * NDirs]int64
	var visited [NCols * NRows]bool
	var parents [NCols * NRows]int64
	var queue queueArray

	// Turn matrix into an array of ints.
	dirs := Directions8()

	// At neighbors[i] we will find the 8 neighbors of node with index i.
	// Each neighbor is another index. If the index is -1, the neighbor is
	// invalid.
	for y := I(0); y.Lt(I(NRows)); y.Inc() {
		for x := I(0); x.Lt(I(NCols)); x.Inc() {
			pt := Pt{x, y}
			index := m.PtToIndex(pt).ToInt() * NDirs
			ns := neighbors[index : index+NDirs]
			for i := range dirs {
				neighbor := pt.Plus(dirs[i])
				if m.InBounds(neighbor) && !m.At(neighbor) {
					ns[i] = m.PtToIndex(neighbor).ToInt64()
				} else {
					ns[i] = -1
				}
			}
		}
	}

	// Convert Pts to ints.
	start := m.PtToIndex(startPt).ToInt64()
	end := m.PtToIndex(endPt).ToInt64()

	// Initialize our structures.
	queue.N = 0 // Make len(p.queue) == 0 without re-allocating.
	for i := range parents {
		parents[i] = -1
		visited[i] = false
	}

	// Process the start element.
	queue.V[queue.N] = start
	queue.N++
	visited[start] = true

	idx := int64(0)
	for idx < queue.N {
		// peek the first element from the queue
		topEl := queue.V[idx]
		if topEl == end {
			// Compute path.
			node := end
			for node >= 0 {
				path.V[path.N] = m.IndexToPt(I64(node))
				path.N++
				node = parents[node]
			}
			slices.Reverse(path.V[0:path.N])
			return
		}

		nIndex := topEl * NDirs
		ns := neighbors[nIndex : nIndex+NDirs]
		for _, n := range ns {
			if n >= 0 && !visited[n] {
				queue.V[queue.N] = n
				queue.N++

				parents[n] = topEl
				visited[n] = true
			}
		}

		// pop the first element out of the queue
		idx++
	}
	return
}
