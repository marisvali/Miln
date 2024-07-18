package main

import (
	. "github.com/marisvali/miln/gamelib"
	"time"
)

func UpdateVersion(start string, end string, version string) {
	startTime, err := time.Parse(time.DateTime, start)
	Check(err)
	endTime, err := time.Parse(time.DateTime, end)
	Check(err)

	db := ConnectToDbSql()
	_, err = db.Exec("UPDATE playthroughs SET version = ? WHERE start_moment BETWEEN ? and ?", version, startTime, endTime)
	Check(err)
}
