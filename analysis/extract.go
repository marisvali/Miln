package main

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"os"
	"path/filepath"
)

func Extract(dir string) {
	entries, err := os.ReadDir(dir)
	Check(err)

	// Create a new CSV file
	file, err := os.Create("output.csv")
	Check(err)
	defer file.Close()

	_, err = file.WriteString("level_name,obstacles,portals,gremlins,hounds,ultrahounds,average_portal_cooldown,health\n")
	Check(err)

	// for _, e := range entries {
	// 	if e.Name() == "20240709-110909-002.mln" {
	// 		entries = []os.DirEntry{e}
	// 		break
	// 	}
	// }

	// Write data to the CSV file line by line
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		data := ReadFile(fullPath)
		if len(data) > 0 { // Games that never start are null.
			playthrough := DeserializePlaythrough(data)

			w := NewWorld(playthrough.Seed)
			for _, input := range playthrough.History {
				w.Step(input)
			}

			playerDead := !w.Player.Health.IsPositive()
			allEnemiesDead := len(w.Enemies) == 0
			gameWon := allEnemiesDead && !playerDead
			gameLost := playerDead && !allEnemiesDead
			gameUnfinished := !playerDead && !allEnemiesDead

			fmt.Printf("%s  %010d  ", entry.Name(), playthrough.Seed)
			if gameWon {
				fmt.Printf("won, life leftover: %d", w.Player.Health.ToInt())
			} else if gameLost {
				fmt.Printf("DIED")
			} else if gameUnfinished {
				fmt.Println("UNfinished")
				continue
			} else {
				Check(fmt.Errorf("what"))
			}
			fmt.Println()
		}
	}
}
