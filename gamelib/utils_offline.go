//go:build offline

package gamelib

import (
	"database/sql"
	"github.com/google/uuid"
)

func InitializeIdInDbHttp(user string,
	releaseVersion int64,
	simulationVersion int64,
	inputVersion int64,
	id uuid.UUID) {
}

func UploadDataToDbHttp(user string,
	releaseVersion int64,
	simulationVersion int64,
	inputVersion int64,
	id uuid.UUID, data []byte) {
}

func SetUserDataHttp(user string, data string) {
}

func GetUserDataHttp(user string) string {
	return ""
}

func ConnectToDbSql() *sql.DB {
	return nil
}

func InitializeIdInDbSql(db *sql.DB, id uuid.UUID) {
}

func UploadDataToDbSql(db *sql.DB, id uuid.UUID, data []byte) {
}

func DownloadDataFromDbSql(db *sql.DB, id uuid.UUID) (data []byte) {
	return
}

func InspectDataFromDbSql(db *sql.DB) {
}
