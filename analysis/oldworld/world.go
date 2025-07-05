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

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Serialize(w io.Writer, data any) {
	enc := gob.NewEncoder(w)
	err := enc.Encode(data)
	Check(err)
}

func Deserialize(r io.Reader, data any) {
	dec := gob.NewDecoder(r)
	err := dec.Decode(data)
	Check(err)
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

type Pt struct {
	X Int
	Y Int
}

type Int struct {
	// This is made public for the sake of serializing and deserializing
	// using the encoding/binary package.
	// Don't access it otherwise.
	Val int64
}

type Matrix[T any] struct {
	cells []T
	size  Pt
}

type MatBool struct {
	Matrix[bool]
}

func (m *Matrix[T]) GobDecode(b []byte) error {
	// See GobEncode for tips on efficiency.
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&m.size); err != nil {
		return err
	}
	if err := decoder.Decode(&m.cells); err != nil {
		return err
	}
	return nil
}

func (m *Matrix[T]) Get(pos Pt) T {
	return m.cells[pos.Y.Val*m.size.X.Val+pos.X.Val]
}

type Wave struct {
	SecondsAfterLastWave Int `yaml:"SecondsAfterLastWave"`
	NHounds              Int `yaml:"NHounds"`
}

func DeserializePlaythrough(data []byte) (p Playthrough) {
	buf := bytes.NewBuffer(Unzip(data))
	var token int64
	Deserialize(buf, &token)
	Deserialize(buf, &p)
	return
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
