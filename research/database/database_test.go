package database

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"log"
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
	defer rows.Close()
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

func TestDatabase22(t *testing.T) {
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
	defer rows.Close()
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
