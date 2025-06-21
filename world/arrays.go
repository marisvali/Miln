package world

type EnemiesArray struct {
	N int64
	V [30]Hound
}

type AmmosArray struct {
	N int64
	V [30]Ammo
}

type SpawnPortalsArray struct {
	N int64
	V [30]SpawnPortal
}

type PlayerInputArray struct {
	N int64
	V [20000]PlayerInput
}

type SpawnPortalParamsArray struct {
	N int64
	V [30]SpawnPortalParams
}

type WavesArray struct {
	N int64
	V [10]Wave
}
