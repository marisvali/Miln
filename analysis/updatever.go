package main

import (
	"database/sql"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"time"
)

func deserializePlaythrough(data []byte) (p Playthrough) {
	// ..
	return
}

func SerializedPlaythrough(seed Int, history []PlayerInput) []byte {
	// ..
	return nil
}

func UpdateVersion() {
	start := "2024-07-14 11:48:56"
	end := "2024-07-14 12:11:02"

	startTime, err := time.Parse(time.DateTime, start)
	Check(err)
	endTime, err := time.Parse(time.DateTime, end)
	Check(err)

	db := ConnectToDbSql()
	rows, err := db.Query("SELECT id, playthrough FROM playthroughs WHERE start_moment BETWEEN ? and ?", startTime, endTime)
	Check(err)
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)

	dbRows := []dbRow{}
	for rows.Next() {
		row := dbRow{}
		err = rows.Scan(&row.id, &row.data)
		Check(err)
		dbRows = append(dbRows, row)
	}

	for i := range dbRows {
		p := deserializePlaythrough(dbRows[i].data)
		newData := SerializedPlaythrough(p.Seed, p.History)

		_, err = db.Exec("UPDATE playthroughs SET version = ?, playthrough = ? WHERE id = ?", SimulationVersion, newData, dbRows[i].id)
		Check(err)
	}
}
