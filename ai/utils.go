package ai

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
	"os"
	"slices"
)

func GetFramesWithActions(playthrough Playthrough) (framesWithActions []int64) {
	for i := range playthrough.History {
		input := playthrough.History[i]
		if input.Move || input.Shoot {
			framesWithActions = append(framesWithActions, int64(i))
		}
	}
	return
}

func GetDecisionFrames(framesWithActions []int64) (decisionFrames []int64) {
	// Assume player decides about 5 frames after the last action.
	// decisionFrames = append(decisionFrames, framesWithActions[0]-15)
	// for i := 0; i < len(framesWithActions)-1; i++ {
	// 	decisionFrames = append(decisionFrames, framesWithActions[i]+5)
	// }

	// Since the player can only do valid actions and he is protected by auto
	// aims, it's possible that a previously invalid action became valid by the
	// time the player acted.
	// For example the player might decide to shoot a guy, but by the time he
	// got around to clicking, the guy moved (real case). Now, the previously
	// best course of action became impossible, technically, because the guy
	// moved and attacking the previous position is useless. But, the auto-aim
	// will save you from this and right-clicking on the previous position will
	// result in an attack on the enemy's new position.
	// So, just.. screw it for now. Have the algorithm compute the best decision
	// on exactly the frame when the player acts.
	decisionFrames = slices.Clone(framesWithActions)
	return
}
func WorldState(frameIdx int, w *World) {
	pts := []Pt{}
	pts = append(pts, w.Player.Pos())

	for i := range w.Enemies.N {
		pts = append(pts, w.Enemies.V[i].Pos())
	}

	fmt.Printf("%04d  ", frameIdx)
	for i := range pts {
		fmt.Printf("%02d %02d  ", pts[i].X.ToInt(), pts[i].Y.ToInt())
	}
	fmt.Println()
}

// NeutralInput generates an input that doesn't do anything but has some values
// for the positions of the mouse. A simple PlayerInput{} would also be neutral
// but a list containing mostly PlayerInput{} values would zip and unzip very
// quickly and efficiently, and this is not representative of realistic
// conditions.
func NeutralInput() PlayerInput {
	return PlayerInput{
		MousePt:            Pt{RInt(I(0), I(1919)), RInt(I(0), I(1079))},
		LeftButtonPressed:  false,
		RightButtonPressed: false,
		Move:               false,
		MovePt:             Pt{RInt(I(0), I(7)), RInt(I(0), I(7))},
		Shoot:              false,
		ShootPt:            Pt{RInt(I(0), I(7)), RInt(I(0), I(7))},
	}
}
func GetHistogram(s []int64) map[int64]int64 {
	h := map[int64]int64{}
	for _, v := range s {
		h[v]++
	}
	return h
}

func OutputHistogram(histogram map[int64]int64, outputFile string, header string) {
	// Collect all the keys.
	keys := make([]int64, 0)
	for k := range histogram {
		keys = append(keys, k)
	}

	// Sort the keys.
	slices.Sort(keys)

	// Print out the map, sorted by keys.
	f, err := os.Create(outputFile)
	Check(err)
	_, err = f.WriteString(fmt.Sprintf("%s\n", header))
	Check(err)
	for _, k := range keys {
		_, err = f.WriteString(fmt.Sprintf("%d,%d\n", k, histogram[k]))
		Check(err)
	}
	Check(f.Close())
}
