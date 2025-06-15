package world

type EnemiesArray struct {
	N    int64
	Data [30]Hound
}

type AmmosArray struct {
	N    int64
	Data [30]Ammo
}

type SpawnPortalsArray struct {
	N    int64
	Data [30]SpawnPortal
}

type PlayerInputArray struct {
	N    int64
	Data [20000]PlayerInput
}

type SpawnPortalParamsArray struct {
	N    int64
	Data [30]SpawnPortalParams
}

type WavesArray struct {
	N    int64
	Data [10]Wave
}
