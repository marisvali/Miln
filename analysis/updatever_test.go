package main

import (
	"github.com/google/uuid"
	"github.com/marisvali/miln/analysis/oldworld"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func CI(i oldworld.Int) Int {
	return I64(i.Val)
}

func CPT(p oldworld.Pt) Pt {
	return Pt{I64(p.X.Val), I64(p.Y.Val)}
}

func TestUpdateFrom16To17(t *testing.T) {
	inputDir := "d:\\Miln\\playthroughs\\gabi-experiment3"
	outputDir := "d:\\Miln\\playthroughs\\gabi-experiment3-17-17"
	_ = os.Mkdir(outputDir, os.ModeDir)
	inputFiles := GetFiles(os.DirFS(inputDir).(FS), ".", "*.mln016")
	for idx := range inputFiles {
		inputFiles[idx] = inputDir + inputFiles[idx][1:]
	}

	for _, inputFile := range inputFiles {
		p := oldworld.DeserializePlaythrough(ReadFile(inputFile))
		var pn Playthrough
		pn.ReleaseVersion = I(17)
		pn.SimulationVersion = I(SimulationVersion)
		pn.InputVersion = I(InputVersion)

		pn.Level.Boardgame = p.Level.Boardgame
		pn.Level.UseAmmo = p.Level.UseAmmo
		pn.Level.AmmoLimit = CI(p.Level.AmmoLimit)
		pn.Level.EnemyMoveCooldownDuration = CI(p.Level.EnemyMoveCooldownDuration)
		pn.Level.EnemiesAggroWhenVisible = p.Level.EnemiesAggroWhenVisible
		pn.Level.SpawnPortalCooldownMin = CI(p.Level.SpawnPortalCooldownMin)
		pn.Level.SpawnPortalCooldownMax = CI(p.Level.SpawnPortalCooldownMax)
		pn.Level.HoundMaxHealth = CI(p.Level.HoundMaxHealth)
		pn.Level.HoundMoveCooldownMultiplier = CI(p.Level.HoundMoveCooldownMultiplier)
		pn.Level.HoundPreparingToAttackCooldown = CI(p.Level.HoundPreparingToAttackCooldown)
		pn.Level.HoundAttackCooldownMultiplier = CI(p.Level.HoundAttackCooldownMultiplier)
		pn.Level.HoundHitCooldownDuration = CI(p.Level.HoundHitCooldownDuration)
		pn.Level.HoundHitsPlayer = p.Level.HoundHitsPlayer
		pn.Level.HoundAggroDistance = CI(p.Level.HoundAggroDistance)

		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				pn.Obstacles.Matrix.Set(IPt(x, y),
					p.Obstacles.Get(oldworld.Pt{
						oldworld.Int{int64(x)}, oldworld.Int{int64(y)}}))
			}
		}

		pn.SpawnPortalsParams.N = int64(len(p.SpawnPortalsParams))
		for i := range p.SpawnPortalsParams {
			pn.SpawnPortalsParams.V[i].SpawnPortalCooldown = CI(p.SpawnPortalsParams[i].SpawnPortalCooldown)
			pn.SpawnPortalsParams.V[i].Pos = CPT(p.SpawnPortalsParams[i].Pos)
			pn.SpawnPortalsParams.V[i].Waves.N = int64(len(p.SpawnPortalsParams[i].Waves))
			for j := range p.SpawnPortalsParams[i].Waves {
				pn.SpawnPortalsParams.V[i].Waves.V[j].SecondsAfterLastWave = CI(p.SpawnPortalsParams[i].Waves[j].SecondsAfterLastWave)
				pn.SpawnPortalsParams.V[i].Waves.V[j].NHounds = CI(p.SpawnPortalsParams[i].Waves[j].NHounds)
			}
		}

		pn.Id = p.Id
		pn.Seed = I64(p.Seed.Val)
		pn.History = make([]PlayerInput, len(p.History))
		for i := range p.History {
			pn.History[i].MousePt = CPT(p.History[i].MousePt)
			pn.History[i].LeftButtonPressed = p.History[i].LeftButtonPressed
			pn.History[i].RightButtonPressed = p.History[i].RightButtonPressed
			pn.History[i].Move = p.History[i].Move
			pn.History[i].MovePt = CPT(p.History[i].MovePt)
			pn.History[i].Shoot = p.History[i].Shoot
			pn.History[i].ShootPt = CPT(p.History[i].ShootPt)
		}

		inputFile = filepath.Base(inputFile)
		var extension = filepath.Ext(inputFile)
		var name = inputFile[0 : len(inputFile)-len(extension)]
		WriteFile(outputDir+"\\"+name+".mln017-017", pn.Serialize())
	}
}

func TestUpload17ToDatabase(t *testing.T) {
	dir := "d:\\Miln\\playthroughs\\gabi-experiment3-17-17"
	_ = os.Mkdir(dir, os.ModeDir)
	inputFiles := GetFiles(os.DirFS(dir).(FS), ".", "*.mln017-017")
	for idx := range inputFiles {
		inputFiles[idx] = dir + inputFiles[idx][1:]
	}

	for _, inputFile := range inputFiles {
		p := DeserializePlaythrough(ReadFile(inputFile))
		p.Id = uuid.New()

		InitializeIdInDbHttp("gabi-experiment3",
			p.ReleaseVersion.ToInt64(),
			p.SimulationVersion.ToInt64(),
			p.InputVersion.ToInt64(),
			p.Id)

		UploadDataToDbHttp("gabi-experiment3",
			p.ReleaseVersion.ToInt64(),
			p.SimulationVersion.ToInt64(),
			p.InputVersion.ToInt64(),
			p.Id,
			p.Serialize())

		time.Sleep(2 * time.Second)
	}
}
