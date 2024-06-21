package world

type HitPermissions struct {
	CanHitEnemy  []bool
	CanHitPortal bool
}

func NewHitPermissions() (h HitPermissions) {
	h.CanHitEnemy = make([]bool, NEnemyTypes)
	for i := range h.CanHitEnemy {
		h.CanHitEnemy[i] = false
	}
	h.CanHitPortal = false
	return
}

func (h *HitPermissions) Add(other HitPermissions) {
	h.CanHitPortal = h.CanHitPortal || other.CanHitPortal
	for i := 0; i < NEnemyTypes; i++ {
		h.CanHitEnemy[i] = h.CanHitEnemy[i] || other.CanHitEnemy[i]
	}
}
