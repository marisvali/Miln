package world

type HitPermissions struct {
	CanHitGremlin    bool
	CanHitHound      bool
	CanHitUltraHound bool
	CanHitPillar     bool
	CanHitKing       bool
	CanHitPortal     bool
	CanHitQuestion   bool
}

func (h *HitPermissions) all() (v []*bool) {
	v = append(v, &h.CanHitGremlin)
	v = append(v, &h.CanHitHound)
	v = append(v, &h.CanHitUltraHound)
	v = append(v, &h.CanHitPillar)
	v = append(v, &h.CanHitKing)
	v = append(v, &h.CanHitPortal)
	v = append(v, &h.CanHitQuestion)
	return
}

func (h *HitPermissions) Add(other HitPermissions) {
	l1 := h.all()
	l2 := other.all()
	for i := range l1 {
		*l1[i] = *l1[i] || *l2[i]
	}
}
