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
	startMoment       time.Time
	endMoment         time.Time
	user              string
	releaseVersion    int64
	simulationVersion int64
	inputVersion      int64
	id                uuid.UUID
	data              []byte
}

func DownloadRecordings() {
	db := ConnectToDbSql()
	rows, err := db.Query("SELECT " +
		"start_moment, " +
		"COALESCE(end_moment, start_moment), " +
		"user, " +
		"release_version, " +
		"COALESCE(simulation_version, -1), " +
		"COALESCE(input_version, -1), " +
		"id, " +
		"playthrough " +
		"FROM playthroughs")
	Check(err)
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)

	dbRows := []dbRow{}
	for rows.Next() {
		row := dbRow{}
		err = rows.Scan(&row.startMoment, &row.endMoment, &row.user,
			&row.releaseVersion, &row.simulationVersion, &row.inputVersion,
			&row.id, &row.data)
		Check(err)
		dbRows = append(dbRows, row)
	}

	for i := range dbRows {
		dir := dbRows[i].user
		_ = os.Mkdir(dir, os.ModeDir)
		m := dbRows[i].startMoment
		var filename string
		if dbRows[i].simulationVersion == -1 || dbRows[i].inputVersion == -1 {
			// -1 values mean the fields were NULL (check the SQL query above).
			// If the simulation or input version is NULL it means we are
			// dealing with a playthrough recorded before splitting the version
			// into release, simulation and input versions.
			// Use the old extension system (e.g. .mln016).
			filename = fmt.Sprintf("%s/%d%02d%02d-%02d%02d%02d.mln%03d", dir,
				m.Year(), m.Month(), m.Day(), m.Hour(), m.Minute(), m.Second(),
				dbRows[i].releaseVersion)
		} else {
			// Use the extension system that includes both simulation and
			// input versions: .mln019-012
			filename = fmt.Sprintf("%s/%d%02d%02d-%02d%02d%02d.mln%03d-%03d", dir,
				m.Year(), m.Month(), m.Day(), m.Hour(), m.Minute(), m.Second(),
				dbRows[i].simulationVersion, dbRows[i].inputVersion)
		}
		WriteFile(filename, dbRows[i].data)
	}
}
