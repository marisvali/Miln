package main

import (
	"bytes"
	"database/sql"
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"time"
)

func deserializePlaythrough(data []byte) (p Playthrough) {
	buf := bytes.NewBuffer(Unzip(data))
	var token int64
	Deserialize(buf, &token)
	if token != Version && token != 2 {
		Check(fmt.Errorf("this code can't update this playthrough "+
			"correctly - we are version %d and playthrough was generated "+
			"with version %d",
			Version, token))
	}
	Deserialize(buf, &token)
	p.Seed = I64(token)
	DeserializeSlice(buf, &p.History)
	return
}

func SerializedPlaythrough(seed Int, history []PlayerInput) []byte {
	buf := new(bytes.Buffer)
	Serialize(buf, int64(Version))
	Serialize(buf, seed.ToInt64())
	SerializeSlice(buf, history)
	return Zip(buf.Bytes())
}

func UpdateVersion() {
	start := "2024-07-14 10:01:21"
	end := "2024-07-14 10:17:16"

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

		_, err = db.Exec("UPDATE playthroughs SET version = ?, playthrough = ? WHERE id = ?", Version, newData, dbRows[i].id)
		Check(err)
	}
}
