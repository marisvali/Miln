package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
)

type Level struct {
	WorldParams        `yaml:"WorldParams"`
	Obstacles          MatBool                `yaml:"Obstacles"`
	SpawnPortalsParams SpawnPortalParamsArray `yaml:"SpawnPortalsParams"`
}

type SpawnPortalParams struct {
	Pos                 Pt         `yaml:"Pos"`
	SpawnPortalCooldown Int        `yaml:"SpawnPortalCooldown"`
	Waves               WavesArray `yaml:"Waves"`
}

type LevelYaml struct {
	Version Int `yaml:"Version"`
	Seed    Int `yaml:"Seed"`
	Level   `yaml:"Level"`
}

// VersionYaml is used in order to load only the version part of a yaml which
// contains a whole Level. The version can then be checked and if it's ok, we
// can attempt to load the whole level. The reason to not just try to load the
// whole level at once is because it might fail and it will give some mismatch
// error that is due to the version not being right in the first place. I prefer
// to have a clearer error that just tells me the versions don't match.
type VersionYaml struct {
	Version Int `yaml:"Version"`
}

func (l *Level) SaveToYAML(seed Int, filename string) {
	var lYaml LevelYaml
	lYaml.Version = I(Version)
	lYaml.Seed = seed
	lYaml.Level = *l
	SaveYAML(filename, lYaml)
}

func LoadLevelFromYAML(fsys FS, filename string) (seed Int, l Level) {
	var vYaml VersionYaml
	LoadYAML(fsys, filename, &vYaml)
	if vYaml.Version.ToInt64() != Version {
		Check(fmt.Errorf("this code can't simulate this playthrough "+
			"correctly - we are version %d and playthrough was generated "+
			"with version %d",
			Version, vYaml.Version.ToInt64()))
	}

	var lYaml LevelYaml
	LoadYAML(fsys, filename, &lYaml)
	return lYaml.Seed, lYaml.Level
}

func IsYamlLevel(filename string) bool {
	b := ReadFile(filename)
	versionS := "Version"
	if len(b) <= len(versionS) {
		return false
	}
	isYamlLevel := string(b[0:len(versionS)]) == versionS
	return isYamlLevel
}
