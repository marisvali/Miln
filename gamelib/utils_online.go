//go:build online

package gamelib

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

// makeHttpRequest makes a POST HTTP request to an endpoint and returns the
// body of the response as a string.
func makeHttpRequest(url string, fields map[string]string, files map[string][]byte) string {
	// Create a buffer to write our multipart form data.
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	for k, v := range fields {
		err := writer.WriteField(k, v)
		Check(err)
	}
	for k, v := range files {
		part, err := writer.CreateFormFile(k, k)
		Check(err)
		_, err = part.Write(v)
		Check(err)
	}
	err := writer.Close()
	Check(err)

	// Create a POST request with the multipart form data.
	request, err := http.NewRequest("POST", url, &requestBody)
	Check(err)
	request.Header.Set("content-type", writer.FormDataContentType())

	// Perform the request.
	client := &http.Client{}
	response, err := client.Do(request)
	Check(err)
	if response.StatusCode != 200 {
		Check(fmt.Errorf("http request failed: %d", response.StatusCode))
	}
	data, err := io.ReadAll(response.Body)
	Check(err)
	return string(data)
}

func InitializeIdInDbHttp(user string, version int64, id uuid.UUID) {
	url := "https://playful-patterns.com/submit-playthrough.php"
	makeHttpRequest(url,
		map[string]string{
			"user":    user,
			"version": strconv.FormatInt(version, 10),
			"id":      id.String()},
		map[string][]byte{})
}

func UploadDataToDbHttp(user string, version int64, id uuid.UUID, data []byte) {
	url := "https://playful-patterns.com/submit-playthrough.php"
	makeHttpRequest(url,
		map[string]string{
			"user":    user,
			"version": strconv.FormatInt(version, 10),
			"id":      id.String()},
		map[string][]byte{"playthrough": data})
}

func SetUserDataHttp(user string, data string) {
	url := "https://playful-patterns.com/set-user-data.php"
	makeHttpRequest(url,
		map[string]string{"user": user, "data": data},
		map[string][]byte{})
}

func GetUserDataHttp(user string) string {
	url := "https://playful-patterns.com/get-user-data.php"
	return makeHttpRequest(url,
		map[string]string{"user": user},
		map[string][]byte{})
}

func ConnectToDbSql() *sql.DB {
	cfg := mysql.Config{
		User:                 os.Getenv("MILN_DBUSER"),
		Passwd:               os.Getenv("MILN_DBPASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("MILN_DBADDR"),
		DBName:               os.Getenv("MILN_DBNAME"),
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	Check(err)
	err = db.Ping()
	Check(err)
	return db
}

func InitializeIdInDbSql(db *sql.DB, id uuid.UUID) {
	_, err := db.Exec("INSERT INTO playthroughs (id) VALUES (?)", id.String())
	Check(err)
}

func UploadDataToDbSql(db *sql.DB, id uuid.UUID, data []byte) {
	_, err := db.Exec("UPDATE playthroughs SET playthrough = ? WHERE id = ?", data, id.String())
	Check(err)
}

func DownloadDataFromDbSql(db *sql.DB, id uuid.UUID) (data []byte) {
	rows, err := db.Query("SELECT playthrough FROM playthroughs WHERE id = ?", id.String())
	Check(err)
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)
	if !rows.Next() {
		Check(fmt.Errorf("id not found: %s", id.String()))
	}
	err = rows.Scan(&data)
	Check(err)
	return
}

func InspectDataFromDbSql(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM playthroughs")
	Check(err)
	defer func(rows *sql.Rows) { Check(rows.Close()) }(rows)

	for rows.Next() {
		var data []byte
		err := rows.Scan(&data)
		Check(err)
		println(len(data))
	}
}
