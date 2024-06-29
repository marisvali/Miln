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

func TestDatabaseUtils(t *testing.T) {
	db := ConnectToDB()
	id := uuid.New()
	InitializeIdInDB(db, id)
	UploadDataToDB(db, id, []byte("what do you mean"))
	InspectDataFromDB(db)
	assert.True(t, true)
}
