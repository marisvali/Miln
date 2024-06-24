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
	k.Permissions = HitPermissions{}
	k.Permissions.CanHitPillar = true
	return
}

func NewHoundKey(pos Pt) (k Key) {
	k.Type = I(1)
	k.Pos = pos
	k.Permissions = HitPermissions{}
	k.Permissions.CanHitHound = true
	k.Permissions.CanHitUltraHound = true
	k.Permissions.CanHitPortal = true
	return
}

func NewPortalKey(pos Pt) (k Key) {
	k.Type = I(2)
	k.Pos = pos
	k.Permissions = HitPermissions{}
	k.Permissions.CanHitPortal = true
	return
}

func NewGremlinKey(pos Pt) (k Key) {
	k.Type = I(3)
	k.Pos = pos
	k.Permissions = HitPermissions{}
	k.Permissions.CanHitGremlin = true
	return
}
