package oldworld

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
)

const Version = 16

type Wave struct {
	SecondsAfterLastWave Int `yaml:"SecondsAfterLastWave"`
	NHounds              Int `yaml:"NHounds"`
}

type WorldParams struct {
	Boardgame                      bool `yaml:"Boardgame"`
	UseAmmo                        bool `yaml:"UseAmmo"`
	AmmoLimit                      Int  `yaml:"AmmoLimit"`
	EnemyMoveCooldownDuration      Int  `yaml:"EnemyMoveCooldownDuration"`
	EnemiesAggroWhenVisible        bool `yaml:"EnemiesAggroWhenVisible"`
	SpawnPortalCooldownMin         Int  `yaml:"SpawnPortalCooldownMin"`
	SpawnPortalCooldownMax         Int  `yaml:"SpawnPortalCooldownMax"`
	HoundMaxHealth                 Int  `yaml:"HoundMaxHealth"`
	HoundMoveCooldownMultiplier    Int  `yaml:"HoundMoveCooldownMultiplier"`
	HoundPreparingToAttackCooldown Int  `yaml:"HoundPreparingToAttackCooldown"`
	HoundAttackCooldownMultiplier  Int  `yaml:"HoundAttackCooldownMultiplier"`
	HoundHitCooldownDuration       Int  `yaml:"HoundHitCooldownDuration"`
	HoundHitsPlayer                bool `yaml:"HoundHitsPlayer"`
	HoundAggroDistance             Int  `yaml:"HoundAggroDistance"`
}

type WorldObject interface {
	Pos() Pt
	State() string
}

type Playthrough struct {
	Level
	Id      uuid.UUID
	Seed    Int
	History []PlayerInput
}

type PlayerInput struct {
	MousePt            Pt
	LeftButtonPressed  bool
	RightButtonPressed bool
	Move               bool
	MovePt             Pt // tile-coordinates
	Shoot              bool
	ShootPt            Pt // tile-coordinates
}

type SpawnPortalParams struct {
	Pos                 Pt     `yaml:"Pos"`
	SpawnPortalCooldown Int    `yaml:"SpawnPortalCooldown"`
	Waves               []Wave `yaml:"Waves"`
}

type Entities struct {
	Obstacles          MatBool             `yaml:"Obstacles"`
	SpawnPortalsParams []SpawnPortalParams `yaml:"SpawnPortalsParams"`
}

type Level struct {
	Entities    `yaml:"Entities"`
	WorldParams `yaml:"WorldParams"`
}

func DeserializePlaythrough(data []byte) (p Playthrough) {
	buf := bytes.NewBuffer(Unzip(data))
	var token int64
	Deserialize(buf, &token)
	// Add a temporary hardcoded rule between versions 13 and 16 as I know they
	// are compatible.
	if token != Version && !(token == 13 && Version == 16) {
		Check(fmt.Errorf("this code can't simulate this playthrough "+
			"correctly - we are version %d and playthrough was generated "+
			"with version %d",
			Version, token))
	}
	Deserialize(buf, &p)
	return
}

func Deserialize(r io.Reader, data any) {
	dec := gob.NewDecoder(r)
	err := dec.Decode(data)
	Check(err)
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Unzip(data []byte) []byte {
	// Get a bytes.Reader, which implements the io.ReaderAt interface required
	// by the zip.NewReader() function.
	bytesReader := bytes.NewReader(data)

	// Open a zip archive for reading.
	r, err := zip.NewReader(bytesReader, int64(len(data)))
	Check(err)

	// We assume there's exactly 1 file in the zip archive.
	if len(r.File) != 1 {
		Check(errors.New(fmt.Sprintf("expected exactly one file in zip archive, got: %d", len(r.File))))
	}

	// Get a reader for that 1 file.
	f := r.File[0]
	rc, err := f.Open()
	Check(err)
	defer func(rc io.ReadCloser) { Check(rc.Close()) }(rc)

	// Keep reading bytes, 1024 bytes at a time.
	buffer := make([]byte, 1024)
	fullContent := make([]byte, 0, 1024)
	for {
		nbytesActuallyRead, err := rc.Read(buffer)
		fullContent = append(fullContent, buffer[:nbytesActuallyRead]...)
		if err == io.EOF {
			break
		}
		Check(err)
		if nbytesActuallyRead == 0 {
			break
		}
	}

	// Return bytes.
	return fullContent
}
