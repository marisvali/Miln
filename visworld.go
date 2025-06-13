package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	_ "image/png"
)

type WorldObjectAnimation struct {
	LastState string
	Object    WorldObject
	Animation Animation
}

type TemporaryAnimation struct {
	ScreenPos   Pt
	Animation   Animation
	NFramesLeft Int
}

type VisWorld struct {
	Animations         Animations
	Objects            []*WorldObjectAnimation
	Temporary          []*TemporaryAnimation
	lastPlayerShooting bool
}

func NewVisWorld(anims Animations) (v VisWorld) {
	v.Animations = anims
	return v
}

func NewWorldObjectAnimation(wo WorldObject, anims Animations) *WorldObjectAnimation {
	var woa WorldObjectAnimation
	woa.Object = wo
	woa.LastState = wo.State()
	switch wo.(type) {
	case *Hound:
		switch wo.State() {
		case "Searching":
			woa.Animation = anims.animHoundSearching
		case "PreparingToAttack":
			woa.Animation = anims.animHoundPreparingToAttack
		case "Attacking":
			woa.Animation = anims.animHoundAttacking
		case "Hit":
			woa.Animation = anims.animHoundHit
		case "Dead":
			woa.Animation = anims.animHoundDead
		}
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
		switch wo.(*Player).State() {
		case "Resting":
			woa.Animation = anims.animPlayer1
		case "Shooting":
			woa.Animation = anims.animPlayer2
		}
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
	for i := range w.EnemiesLen {
		objs = append(objs, &w.Enemies[i])
	}
	for i := range w.SpawnPortals {
		objs = append(objs, &w.SpawnPortals[i])
	}

	// Create animations for objects that don't have them.
	// Either an animation wasn't created for this object or the object's
	// state has changed since the animation was created.
	for _, o := range objs {
		if objToAnim[o] == nil || o.State() != objToAnim[o].LastState {
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

func (v *VisWorld) Step(w *World, input PlayerInput, guiData GuiData) {
	if input.LeftButtonPressed && !input.Move && guiData.HighlightMoveNotOk {
		moveFailed := TemporaryAnimation{}
		moveFailed.Animation = v.Animations.animMoveFailed
		moveFailed.NFramesLeft = I(20)
		moveFailed.ScreenPos = input.MousePt
		v.Temporary = append(v.Temporary, &moveFailed)
	}
	if input.RightButtonPressed && !input.Shoot && guiData.HighlightAttack {
		attackFailed := TemporaryAnimation{}
		attackFailed.Animation = v.Animations.animAttackFailed
		attackFailed.NFramesLeft = I(20)
		attackFailed.ScreenPos = input.MousePt
		v.Temporary = append(v.Temporary, &attackFailed)
	}

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
