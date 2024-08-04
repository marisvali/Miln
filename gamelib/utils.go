package gamelib

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"io"
	"io/fs"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var CheckCrashes = true
var CheckFailed error

func Check(e error) {
	if e != nil {
		CheckFailed = e
		if CheckCrashes {
			panic(e)
		}
	}
}

func CloseFile(f *os.File) {
	Check(f.Close())
}

func WriteFile(name string, data []byte) {
	err := os.WriteFile(name, data, 0644)
	Check(err)
}

func ReadFile(name string) []byte {
	data, err := os.ReadFile(name)
	Check(err)
	return data
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err == nil {
		return true
	}
	return false
}

func GetNewRecordingFile() string {
	if !FileExists("recordings") {
		return ""
	}
	date := time.Now()
	for i := 0; i < 1000000; i++ {
		filename := fmt.Sprintf("recordings/recorded-inputs-%04d-%02d-%02d-%06d.mln",
			date.Year(), date.Month(), date.Day(), i)
		if !FileExists(filename) {
			return filename
		}
	}
	panic("Cannot record, no available filename found.")
}

func GetLatestRecordingFile() string {
	dir := "recordings"
	if !FileExists(dir) {
		return ""
	}
	entries, err := os.ReadDir(dir)
	Check(err)

	candidates := []string{}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".mln") {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return ""
	}

	slices.Sort(candidates)
	return dir + "/" + candidates[len(candidates)-1]
}

//
// func TouchFile(name string) {
//	name = "e:/" + name
//	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
//	Check(err)
//	err = file.Close()
//	Check(err)
// }
//
// func FileExists(name string) bool {
//	name = "e:/" + name
//	if _, err := os.Stat(name); err == nil {
//		return true
//	}
//	return false
// }
//
// func WaitForFile(name string) {
//	name = "e:/" + name
//	for {
//		if _, err := os.Stat(name); err == nil {
//			for {
//				// Attempt to remove the file until the attempt succeeds.
//				err = os.Remove(name)
//				if err == nil {
//					return
//				}
//			}
//		} else if errors.Is(err, os.ErrNotExist) {
//			// name does not exist
//		} else {
//			Check(err)
//		}
//	}
// }
//
// func DeleteFile(name string) {
//	name = "e:/" + name
//	err := os.Remove(name)
//	if !errors.Is(err, os.ErrNotExist) {
//		Check(err)
//	}
// }

func Serialize(w io.Writer, data any) {
	err := binary.Write(w, binary.LittleEndian, data)
	Check(err)
}

func Deserialize(r io.Reader, data any) {
	err := binary.Read(r, binary.LittleEndian, data)
	Check(err)
}

func SerializeSlice[T any](buf *bytes.Buffer, s []T) {
	Serialize(buf, int64(len(s)))
	Serialize(buf, s)
}

func DeserializeSlice[T any](buf *bytes.Buffer, s *[]T) {
	var lenSlice int64
	Deserialize(buf, &lenSlice)
	*s = make([]T, lenSlice)
	Deserialize(buf, *s)
}

type TimedFunction func()

func Duration(function TimedFunction) float64 {
	start := time.Now()
	function()
	return time.Since(start).Seconds()
}

func ReadAllText(filename string) string {
	file, err := os.Open(filename)
	Check(err)
	data, err := io.ReadAll(file)
	Check(err)
	return string(data)
}

func LoadJSON(filename string, v any) {
	file, err := os.Open(filename)
	Check(err)
	data, err := io.ReadAll(file)
	Check(err)
	err = json.Unmarshal(data, v)
	Check(err)
}

type FolderWatcher struct {
	Folder string
	times  []time.Time
}

func (f *FolderWatcher) FolderContentsChanged() bool {
	if f.Folder == "" {
		return false
	}

	files, err := os.ReadDir(f.Folder)
	Check(err)
	if len(files) != len(f.times) {
		f.times = make([]time.Time, len(files))
	}
	changed := false
	for idx, file := range files {
		info, err := file.Info()
		Check(err)
		if f.times[idx] != info.ModTime() {
			changed = true
			f.times[idx] = info.ModTime()
		}
	}
	return changed
}

func HomeFolder() string {
	ex, err := os.Executable()
	Check(err)
	return filepath.Dir(ex)
}

func Home(relativePath string) string {
	return path.Join(HomeFolder(), relativePath)
}

func Unzip(data []byte) []byte {
	// Get a bytes.Reader, which implements the io.ReaderAt interface required
	// by the zip.NewReader() function.
	bytesReader := bytes.NewReader(data)

	// Open a zip archive for reading.
	r, err := zip.NewReader(bytesReader, int64(len(data)))
	Check(err)

	// We assume there's exactly 1 file in the zip archive.
	if len(r.File) != 1 {
		Check(errors.New(fmt.Sprintf("expected exactly one file in zip archive, got: %d", len(r.File))))
	}

	// Get a reader for that 1 file.
	f := r.File[0]
	rc, err := f.Open()
	Check(err)
	defer func(rc io.ReadCloser) { Check(rc.Close()) }(rc)

	// Keep reading bytes, 1024 bytes at a time.
	buffer := make([]byte, 1024)
	fullContent := make([]byte, 0, 1024)
	for {
		nbytesActuallyRead, err := rc.Read(buffer)
		fullContent = append(fullContent, buffer[:nbytesActuallyRead]...)
		if err == io.EOF {
			break
		}
		Check(err)
		if nbytesActuallyRead == 0 {
			break
		}
	}

	// Return bytes.
	return fullContent
}

func UnzipFromFile(filename string) []byte {
	return Unzip(ReadFile(filename))
}

func Zip(data []byte) []byte {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Create a single file inside it called "recorded-inputs".
	f, err := w.Create("recorded-inputs")
	Check(err)

	// Write/compress the data to the file inside the zip.
	_, err = f.Write(data)
	Check(err)

	// Make sure to check the error on Close.
	err = w.Close()
	Check(err)

	return buf.Bytes()
}

func ZipToFile(filename string, data []byte) {
	// Actually write the zip to disk.
	WriteFile(filename, Zip(data))
}

func LoadImage(str string) *ebiten.Image {
	file, err := os.Open(str)
	defer func(file *os.File) { Check(file.Close()) }(file)
	Check(err)

	img, _, err := image.Decode(file)
	Check(err)
	if err != nil {
		return nil
	}

	return ebiten.NewImageFromImage(img)
}

func LoadImageEmbedded(str string, efs *embed.FS) *ebiten.Image {
	file, err := efs.Open(str)
	defer func(file fs.File) { Check(file.Close()) }(file)
	Check(err)

	img, _, err := image.Decode(file)
	Check(err)
	if err != nil {
		return nil
	}

	return ebiten.NewImageFromImage(img)
}

func EqualFloats(f1, f2 float64) bool {
	return math.Abs(f1-f2) < 0.000001
}

func HexToColor(hexVal int) color.Color {
	if hexVal < 0x000000 || hexVal > 0xFFFFFF {
		panic(fmt.Sprintf("Invalid HEX value for color: %d", hexVal))
	}
	r := uint8(hexVal & 0xFF0000 >> 16)
	g := uint8(hexVal & 0x00FF00 >> 8)
	b := uint8(hexVal & 0x0000FF)
	return color.RGBA{
		R: r,
		G: g,
		B: b,
		A: 255,
	}
}

// Remove modifies the underlying array, which may be what you want, or
// may not be what you want.
func Remove[S ~[]E, E any](s S, i int) S {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func ComputeSpriteMask(img *ebiten.Image) *ebiten.Image {
	mask := ebiten.NewImageFromImage(img)
	sz := mask.Bounds().Size()
	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				mask.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
	return mask
}

func sendDataToDbHttp(user string, version int64, id uuid.UUID, data []byte) {
	url := "https://playful-patterns.com/submit-playthrough.php"

	// Create a buffer to write our multipart form data.
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	err := writer.WriteField("user", user)
	Check(err)
	err = writer.WriteField("version", strconv.FormatInt(version, 10))
	Check(err)
	err = writer.WriteField("id", id.String())
	Check(err)
	if data != nil {
		part, err := writer.CreateFormFile("playthrough", "rima")
		Check(err)
		_, err = part.Write(data)
		Check(err)
	}
	err = writer.Close()
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
}

func InitializeIdInDbHttp(user string, version int64, id uuid.UUID) {
	sendDataToDbHttp(user, version, id, nil)
}

func UploadDataToDbHttp(user string, version int64, id uuid.UUID, data []byte) {
	sendDataToDbHttp(user, version, id, data)
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

func Directions8() []Pt {
	// This order is needed so that straight lines get priority in pathfinding.
	return []Pt{
		// left/right/up/down
		{I(1).Negative(), I(0)},
		{I(1), I(0)},
		{I(0), I(1).Negative()},
		{I(0), I(1)},

		// diagonals
		{I(1).Negative(), I(1).Negative()},
		{I(1), I(1).Negative()},
		{I(1).Negative(), I(1)},
		{I(1), I(1)},
	}
}

func MatrixFromString[T comparable](str string, vals map[byte]T) (m Matrix[T]) {
	row := -1
	col := 0
	maxCol := 0
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c == '\n' {
			maxCol = col
			col = 0
			row++
			continue
		}
		col++
	}
	// If the string does not end with an empty line, count the last row.
	if col > 0 {
		row++
	}
	m = NewMatrix[T](IPt(maxCol, row))

	row = -1
	col = 0
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c == '\n' {
			col = 0
			row++
			continue
		} else if val, ok := vals[c]; ok {
			m.Set(IPt(col, row), val)
		}
		col++
	}
	return
}

// HashBytes receives a byte array and returns its SHA-256 hash as a hex string.
func HashBytes(input []byte) string {
	// Create a new SHA-256 hash
	hash := sha256.New()

	// Write the byte slice to the hash
	hash.Write(input)

	// Get the resulting hash as a byte slice
	hashBytes := hash.Sum(nil)

	// Convert the byte slice to a hex string
	hashHex := hex.EncodeToString(hashBytes)

	return hashHex
}

func SplitInLines(content []byte) (lines []string) {
	lastI := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			line := string(content[lastI:i])
			lines = append(lines, line)
			lastI = i + 1
		}
	}
	if len(content) > lastI {
		line := string(content[lastI:])
		lines = append(lines, line)
	}
	return
}
