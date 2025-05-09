package main

import (
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"os"
	"strconv"
)

func intFromBool(b bool) Int {
	if b {
		return ONE
	} else {
		return ZERO
	}
}

func Extract() {
	// Create a new CSV file
	file, err := os.Create("output.csv")
	Check(err)
	defer CloseFile(file)

	inputFile := "d:\\gms\\Miln\\analysis\\tools\\playthroughs\\denis\\20250319-170648.mln010"
	playthrough := DeserializePlaythrough(ReadFile(inputFile))

	_, err = file.WriteString("frame_idx,mouse_x,mouse_y,left_clicked,right_clicked\n")
	Check(err)

	for i := 1; i < len(playthrough.History); i++ {
		input := playthrough.History[i]
		// Write data to the CSV file line by line
		nums := []Int{I(i + 1), input.MousePt.X, input.MousePt.Y, intFromBool(input.LeftButtonPressed), intFromBool(input.RightButtonPressed)}
		for _, num := range nums {
			_, err = file.WriteString(strconv.Itoa(num.ToInt()) + ",")
			Check(err)
		}
		_, err = file.WriteString("\n")
		Check(err)
	}
}
