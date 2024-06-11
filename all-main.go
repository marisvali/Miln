package main

//
//import (
//	"github.com/hajimehoshi/ebiten/v2"
//	"os"
//	"playful-patterns.com/miln/utils"
//)
//
//func main() {
//	if len(os.Args) == 1 {
//		RunGuiFusedPlay(utils.GetNewRecordingFile())
//	} else {
//		RunGuiFusedPlayback(os.Args[1])
//		//RunGuiFusedPlayback("d:/gms/bakoko/recordings/recorded-inputs-2024-03-20-000000")
//	}
//}
//
//func RunGuiFusedPlay(recordingFile string) {
//	var worldRunner WorldRunner
//	var player2Ai PlayerAI
//	worldRunner.Initialize(recordingFile, true)
//
//	var g Gui
//	g.Init(nil, &worldRunner, &player2Ai, "", []string{})
//
//	// Start the game.
//	err := ebiten.RunGame(&g)
//	utils.Check(err)
//}
//
//func RunGuiFusedPlayback(recordingFile string) {
//	var worldRunner WorldRunner
//	var player2Ai PlayerAI
//	worldRunner.Initialize("", false)
//
//	var g Gui
//	g.Init(nil, &worldRunner, &player2Ai, recordingFile, []string{})
//
//	// Start the game.
//	err := ebiten.RunGame(&g)
//	utils.Check(err)
//}
