package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
)

// Level is only meant to be an input for World. It is important for World to
// use arrays, not slices, for performance reasons. Level isn't relevant to
// performance yet, so it could use slices. Slices are treated nicely by the
// YAML serialization library. The YAML is meant to be human-readable and it's
// annoying to have full arrays in there that are mostly empty.
// However, I also want to copy/clone the Playthrough, which implies the Level.
// And if Level has slices I have to write clone logic for that. That's annoying
// for me. It's actually easier to write MarshalYAML and UnmarshalYAML for the
// 2 arrays that need it. I prefer to have special code for serializing to YAML,
// and have a simpler version of a basic operation like copy/clone. This way I
// can also easily include the Level in WorldDebugInfo if I want, and trust
// that it gets copied quickly and correctly.
// Playthrough.History is different, it makes sense to have it a slice because
// during real playthroughs I don't know how many frames the player will use
// so I would have to add extra logic to make sure I don't crash. But Level is
// exactly the kind of structure where it makes sense to decide the maximum
// size in advance.
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
	InputVersion Int `yaml:"InputVersion"`
	Seed         Int `yaml:"Seed"`
	Level        `yaml:"Level"`
}

// VersionYaml is used in order to load only the version part of a yaml which
// contains a whole Level. The version can then be checked and if it's ok, we
// can attempt to load the whole level. The reason to not just try to load the
// whole level at once is because it might fail and it will give some mismatch
// error that is due to the version not being right in the first place. I prefer
// to have a clearer error that just tells me the versions don't match.
type VersionYaml struct {
	InputVersion Int `yaml:"InputVersion"`
}

func (l *Level) SaveToYAML(seed Int, filename string) {
	var lYaml LevelYaml
	lYaml.InputVersion = I(InputVersion)
	lYaml.Seed = seed
	lYaml.Level = *l
	SaveYAML(filename, lYaml)
}

func LoadLevelFromYAML(fsys FS, filename string) (seed Int, l Level) {
	var vYaml VersionYaml
	LoadYAML(fsys, filename, &vYaml)
	if vYaml.InputVersion.ToInt64() != InputVersion {
		Check(fmt.Errorf("can't load this level as it was generated using "+
			"InputVersion %d and we are at InputVersion %d - it is very "+
			"likely that if we continued deserialization the info would be "+
			"loaded wrong, but silently, because that's how .yaml loading "+
			"works; if you feel confident that the level matches the exe "+
			"that's trying to load it, change the InputVersion manually in"+
			" the .yaml file; this can happen if the InputVersion changed but "+
			"the Level stayed the same (currently this means that the "+
			"PlayerInput structure changed but not the Level structure)",
			vYaml.InputVersion.ToInt64(), InputVersion))
	}

	var lYaml LevelYaml
	LoadYAML(fsys, filename, &lYaml)
	return lYaml.Seed, lYaml.Level
}

func IsYamlLevel(filename string) bool {
	b := ReadFile(filename)
	versionS := "InputVersion"
	if len(b) <= len(versionS) {
		return false
	}
	isYamlLevel := string(b[0:len(versionS)]) == versionS
	return isYamlLevel
}
