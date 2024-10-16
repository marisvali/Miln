package world

import . "github.com/marisvali/miln/gamelib"

func Level1() string {
	return `
      
 xxxxx
      
  1xxx
   2 x
     x
`
}

func LevelFromString(level string) (m MatBool, pos1 []Pt, pos2 []Pt) {
	row := -1
	col := 0
	maxCol := 0
	for i := 0; i < len(level); i++ {
		c := level[i]
		if c == '\n' {
			maxCol = col
			col = 0
			row++
			continue
		}
		col++
	}
	// If the string does not end with an empty line, count the last row.
	if col > 0 {
		row++
	}
	m = NewMatBool(IPt(maxCol, row))

	row = -1
	col = 0
	for i := 0; i < len(level); i++ {
		c := level[i]
		if c == '\n' {
			col = 0
			row++
			continue
		} else if c == 'x' {
			m.Set(IPt(col, row))
		} else if c == '1' {
			pos1 = append(pos1, IPt(col, row))
		} else if c == '2' {
			pos2 = append(pos2, IPt(col, row))
		}
		col++
	}
	return
}
