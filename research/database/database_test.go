package database

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	var db *sql.DB
	// Capture connection properties.

	cfg := mysql.Config{
		User:                 "",
		Passwd:               "",
		Net:                  "tcp",
		Addr:                 "",
		DBName:               "",
		AllowNativePasswords: true,
	}

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var c1, c2, c3, c4 int
		if err := rows.Scan(&c1, &c2, &c3, &c4); err != nil {
			log.Fatal(err)
		}
		println(c1, c2, c3, c4)
	}

	_, errIns := db.Exec("INSERT INTO test (c1, c2, c3) VALUES (?, ?, ?)", 999, 99, 89)
	if errIns != nil {
		log.Fatal(errIns)
	}

	assert.True(t, true)
}

func TestDatabaseBytes(t *testing.T) {
	var db *sql.DB

	// Capture connection properties.
	cfg := mysql.Config{
		User:                 os.Getenv("MILN_DBUSER"),
		Passwd:               os.Getenv("MILN_DBPASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("MILN_DBADDR"),
		DBName:               os.Getenv("MILN_DBNAME"),
		AllowNativePasswords: true,
	}

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	data := []byte("Kyle Kinane")
	_, errIns := db.Exec("INSERT INTO test3 (recording) VALUES (?)", data)
	if errIns != nil {
		log.Fatal(errIns)
	}

	rows, err := db.Query("SELECT * FROM test3")
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var data []byte
		var id int
		if err := rows.Scan(&id, &data); err != nil {
			log.Fatal(err)
		}
		println(id, string(data))
	}

	assert.True(t, true)
}

type dbRow struct {
	startMoment time.Time
	endMoment   time.Time
	user        string
	version     int64
	id          uuid.UUID
	data        []byte
}

func TestDatabaseSaveRecordings(t *testing.T) {
	db := ConnectToDbSql()
	rows, err := db.Query("SELECT start_moment, COALESCE(end_moment, start_moment), user, version, id, playthrough FROM playthroughs")
	Check(err)
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)

	dbRows := []dbRow{}
	for rows.Next() {
		row := dbRow{}
		err := rows.Scan(&row.startMoment, &row.endMoment, &row.user, &row.version, &row.id, &row.data)
		Check(err)
		dbRows = append(dbRows, row)
	}

	for i := range dbRows {
		dir := dbRows[i].user
		_ = os.Mkdir(dir, os.ModeDir)
		m := dbRows[i].startMoment
		filename := fmt.Sprintf("%s/%d%02d%02d-%02d%02d%02d-%03d.mln", dir, m.Year(),
			m.Month(), m.Day(), m.Hour(), m.Minute(), m.Second(), dbRows[i].version)
		WriteFile(filename, dbRows[i].data)
	}

	assert.True(t, true)
}
