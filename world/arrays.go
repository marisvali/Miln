package world

import (
	"github.com/goccy/go-yaml"
)

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

// MarshalYAML turns the array into a string.
// Useful because if I just let the YAML library do the default marshalling, it
// will turn the N field into "n" and it will output all the elements in the
// array, not limit to the N.
func (a SpawnPortalParamsArray) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(a.V[0:a.N])
}

func (a *SpawnPortalParamsArray) UnmarshalYAML(b []byte) error {
	var v []SpawnPortalParams
	err := yaml.Unmarshal(b, &v)
	a.N = int64(len(v))
	copy(a.V[:], v)
	return err
}

// MarshalYAML turns the array into a string.
// Useful because if I just let the YAML library do the default marshalling, it
// will turn the N field into "n" and it will output all the elements in the
// array, not limit to the N.
func (a WavesArray) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(a.V[0:a.N])
}

func (a *WavesArray) UnmarshalYAML(b []byte) error {
	var v []Wave
	err := yaml.Unmarshal(b, &v)
	a.N = int64(len(v))
	copy(a.V[:], v)
	return err
}
