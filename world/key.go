package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Key struct {
	Pos         Pt
	Permissions HitPermissions
	Type        Int
}

func NewPillarKey(pos Pt) (k Key) {
	k.Type = I(0)
	k.Pos = pos
	k.Permissions = NewHitPermissions()
	k.Permissions.CanHitEnemy[1] = true
	return
}

func NewHoundKey(pos Pt) (k Key) {
	k.Type = I(1)
	k.Pos = pos
	k.Permissions = NewHitPermissions()
	k.Permissions.CanHitEnemy[2] = true
	k.Permissions.CanHitPortal = true
	return
}

func NewPortalKey(pos Pt) (k Key) {
	k.Type = I(2)
	k.Pos = pos
	k.Permissions = NewHitPermissions()
	k.Permissions.CanHitPortal = true
	return
}

func NewGremlinKey(pos Pt) (k Key) {
	k.Type = I(3)
	k.Pos = pos
	k.Permissions = NewHitPermissions()
	k.Permissions.CanHitEnemy[0] = true
	return
}
