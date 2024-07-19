package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	. "github.com/marisvali/miln/gamelib"
	"os"
	"time"
)

type dbRow struct {
	startMoment time.Time
	endMoment   time.Time
	user        string
	version     int64
	id          uuid.UUID
	data        []byte
}

func DownloadRecordings() {
	db := ConnectToDbSql()
	rows, err := db.Query("SELECT start_moment, COALESCE(end_moment, start_moment), user, version, id, playthrough FROM playthroughs")
	Check(err)
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)

	dbRows := []dbRow{}
	for rows.Next() {
		row := dbRow{}
		err = rows.Scan(&row.startMoment, &row.endMoment, &row.user, &row.version, &row.id, &row.data)
		Check(err)
		dbRows = append(dbRows, row)
	}

	for i := range dbRows {
		dir := dbRows[i].user
		_ = os.Mkdir(dir, os.ModeDir)
		m := dbRows[i].startMoment
		filename := fmt.Sprintf("%s/%d%02d%02d-%02d%02d%02d.mln%03d", dir, m.Year(),
			m.Month(), m.Day(), m.Hour(), m.Minute(), m.Second(), dbRows[i].version)
		WriteFile(filename, dbRows[i].data)
	}
}
