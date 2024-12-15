package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

type WorldObjectAnimation struct {
	Object    WorldObject
	Animation Animation
}

type TemporaryAnimation struct {
	ScreenPos   Pt
	Animation   Animation
	NFramesLeft Int
}

type VisWorld struct {
	Animations Animations
	Objects    []*WorldObjectAnimation
	Temporary  []*TemporaryAnimation
}

func NewVisWorld(anims Animations) (v VisWorld) {
	v.Animations = anims
	return v
}

func NewWorldObjectAnimation(wo WorldObject, anims Animations) *WorldObjectAnimation {
	var woa WorldObjectAnimation
	woa.Object = wo
	switch wo.(type) {
	// case *Gremlin:
	// 	woa.Animation = anims.animMoveFailed
	// case *Hound:
	// 	woa.Animation = anims.animMoveFailed
	// case *UltraHound:
	// 	woa.Animation = anims.animMoveFailed
	// case *Pillar:
	// 	woa.Animation = anims.animMoveFailed
	// case *King:
	// 	woa.Animation = anims.animMoveFailed
	// case *Question:
	// 	woa.Animation = anims.animMoveFailed
	case *Player:
		woa.Animation = anims.animPlayer
	// case *SpawnPortal:
	// 	woa.Animation = anims.animMoveFailed
	default:
		// For now, allow world objects without animations.
		// Check(fmt.Errorf("no animation defined for object type %T", woType))
	}
	return &woa
}

func (v *VisWorld) UpdateWhichObjectsExist(w *World) {
	// Map objects to animations.
	objToAnim := map[WorldObject]*WorldObjectAnimation{}
	for _, o := range v.Objects {
		objToAnim[o.Object] = o
	}

	// Get array of all world objects.
	objs := []WorldObject{}
	if w.Player.OnMap {
		objs = append(objs, &w.Player)
	}
	for _, e := range w.Enemies {
		objs = append(objs, e)
	}
	for i := range w.SpawnPortals {
		objs = append(objs, &w.SpawnPortals[i])
	}

	// Create animations for objects that don't have them.
	for _, o := range objs {
		if objToAnim[o] == nil {
			woa := NewWorldObjectAnimation(o, v.Animations)
			v.Objects = append(v.Objects, woa)
			objToAnim[o] = woa
		}
	}

	// Keep animations only for the objects that still exist.
	aliveObjects := []*WorldObjectAnimation{}
	for _, o := range objs {
		aliveObjects = append(aliveObjects, objToAnim[o])
	}
	v.Objects = aliveObjects
}

func (v *VisWorld) Step(w *World) {
	v.UpdateWhichObjectsExist(w)

	// Remove obsolete animations.
	validAnimations := []*TemporaryAnimation{}
	for _, anim := range v.Temporary {
		if anim.NFramesLeft.IsPositive() {
			validAnimations = append(validAnimations, anim)
		}
	}
	v.Temporary = validAnimations

	// Get the next frame for each animation.
	for _, o := range v.Objects {
		o.Animation.Step()
	}
	for _, a := range v.Temporary {
		a.NFramesLeft.Dec()
		a.Animation.Step()
	}
}
