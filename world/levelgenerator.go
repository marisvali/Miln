package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
)

type WaveData struct {
	SecondsAfterLastWave Int `yaml:"SecondsAfterLastWave"`
	NHoundMin            Int `yaml:"NHoundMin"`
	NHoundMax            Int `yaml:"NHoundMax"`
}

type SpawnPortalData struct {
	Waves []WaveData `yaml:"Waves"`
}

type NEntities struct {
	NumRows          Int               `yaml:"NumRows"`
	NumCols          Int               `yaml:"NumCols"`
	NObstaclesMin    Int               `yaml:"NObstaclesMin"`
	NObstaclesMax    Int               `yaml:"NObstaclesMax"`
	SpawnPortalDatas []SpawnPortalData `yaml:"SpawnPortalDatas"`
}

type LevelGeneratorParams struct {
	NEntitiesPath   string `yaml:"NEntitiesPath"`
	WorldParamsPath string `yaml:"WorldParamsPath"`
	NEntities
	WorldParams
}

func IsLevelValid(m MatBool) bool {
	// Get all unoccupied positions connected to the first unoccupied position.
	m2 := m.ConnectedPositions(FirstUnoccupiedPos(m))

	// Check if all unoccupied positions in m are also unoccupied in m2.
	// If yes, all unoccupied positions are connected.
	// Unoccupied positions in m2 are true, while in m they are false. So negate
	// m2 and compare m2 and m.
	m2.Negate()
	return m == m2
}

func RandomLevel(nObstacles Int) (m MatBool) {
	// Create matrix with obstacles.
	for i := ZERO; i.Lt(nObstacles); i.Inc() {
		m.OccupyRandomPos(&DefaultRand)
	}
	return
}

func ValidRandomLevel(nObstacles Int) (m MatBool) {
	nTries := 0
	for {
		nTries++
		if nTries > 1000 {
			panic(fmt.Errorf("failed to generate valid level for nObstacles: %d", nObstacles))
		}
		m = RandomLevel(nObstacles)
		if IsLevelValid(m) {
			return
		}
	}
}

func LoadLevelGeneratorParams(fsys FS) LevelGeneratorParams {
	// Read from the disk over and over until a full read is possible.
	// This repetition is meant to avoid crashes due to reading files
	// while they are still being written.
	// It's a hack but possibly a quick and very useful one.
	// This repeated reading is only useful when we're not reading from the
	// embedded filesystem. When we're reading from the embedded filesystem we
	// want to crash as soon as possible. We might be in the browser, in which
	// case we want to see an error in the developer console instead of a page
	// that keeps trying to load and reports nothing.
	var p LevelGeneratorParams
	if fsys == nil {
		CheckCrashes = false
	}
	for {
		CheckFailed = nil
		LoadYAML(fsys, "data/levelgenerator/level.yaml", &p)
		LoadYAML(fsys, "data/levelgenerator/"+p.NEntitiesPath, &p.NEntities)
		LoadYAML(fsys, "data/levelgenerator/"+p.WorldParamsPath, &p.WorldParams)
		if CheckFailed == nil {
			break
		}
	}
	CheckCrashes = true
	return p
}

func GenerateLevel(fsys FS) (l Level) {
	p := LoadLevelGeneratorParams(fsys)

	l.Boardgame = p.Boardgame
	l.UseAmmo = p.UseAmmo
	l.AmmoLimit = p.AmmoLimit
	l.EnemyMoveCooldownDuration = p.EnemyMoveCooldownDuration
	l.EnemiesAggroWhenVisible = p.EnemiesAggroWhenVisible
	l.WorldParams = p.WorldParams

	l.Obstacles = ValidRandomLevel(RInt(p.NObstaclesMin, p.NObstaclesMax))

	occ := l.Obstacles
	for idx, portal := range p.SpawnPortalDatas {
		// Build Waves from WaveDatas.
		var waves WavesArray
		for i, wd := range portal.Waves {
			var wave Wave
			wave.SecondsAfterLastWave = wd.SecondsAfterLastWave
			wave.NHounds = RInt(wd.NHoundMin, wd.NHoundMax)
			waves.V[i] = wave
		}
		waves.N = int64(len(portal.Waves))

		// Build spawn portal using waves.
		l.SpawnPortalsParams.V[idx] = SpawnPortalParams{occ.OccupyRandomPos(&DefaultRand),
			RInt(p.SpawnPortalCooldownMin, p.SpawnPortalCooldownMax), waves}
	}
	l.SpawnPortalsParams.N = int64(len(p.SpawnPortalDatas))
	return
}

func FirstUnoccupiedPos(m MatBool) (unoccupiedPos Pt) {
	unoccupiedPos = IPt(0, 0)
	for unoccupiedPos.Y = ZERO; unoccupiedPos.Y.Lt(I(NRows)); unoccupiedPos.Y.Inc() {
		for unoccupiedPos.X = ZERO; unoccupiedPos.X.Lt(I(NCols)); unoccupiedPos.X.Inc() {
			if !m.At(unoccupiedPos) {
				return
			}
		}
	}
	panic(fmt.Errorf("no unoccupied position found"))
}
