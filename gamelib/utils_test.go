package gamelib

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestZip(t *testing.T) {
	data1 := []byte("some kind of data1 aaaaaaaaaaaaaaaaaaaaaaa")
	data2 := []byte("some kind of data1 baaaaaaaaaaaaaaaaaaaaaa")
	data3 := []byte("some kind of data1 baaaaaaaaaaaaaaaaaaaaaa")
	zippedData1 := Zip(data1)
	unzippedData1 := Unzip(zippedData1)
	zippedData2 := Zip(data2)
	unzippedData2 := Unzip(zippedData2)
	zippedData3 := Zip(data3)
	assert.Equal(t, data1, unzippedData1)
	assert.Equal(t, data2, unzippedData2)
	assert.NotEqual(t, zippedData1, zippedData2)
	assert.Equal(t, zippedData2, zippedData3)
}

func TestMatrixFromString(t *testing.T) {
	mapping := map[byte]Int{'x': ONE}

	expected1 := NewMatrix[Int](IPt(3, 2))
	expected1.Set(IPt(0, 1), ONE)
	result1 := MatrixFromString("\nabc\nxyz", mapping)
	assert.Equal(t, expected1, result1)

	expected2 := NewMatrix[Int](IPt(7, 3))
	expected2.Set(IPt(4, 0), ONE)
	expected2.Set(IPt(2, 1), ONE)
	result2 := MatrixFromString(`
----x--
--x----
-------
`, mapping)
	assert.Equal(t, expected2, result2)
}

func TestConnectedPositions(t *testing.T) {
	var m1 MatBool
	m1.Matrix = MatrixFromString(`
x---x--
--x-xx-
--x----
`, map[byte]bool{'x': true})

	var expected1 MatBool
	expected1.Matrix = MatrixFromString(`
x---x--
--x-xx-
--x----
`, map[byte]bool{'x': true, '-': false})

	result1 := m1.ConnectedPositions(IPt(1, 0))
	assert.Equal(t, expected1, result1)

	var expected2 MatBool
	expected2.Matrix = MatrixFromString(`
x---a--
--x-aa-
--x----
`, map[byte]bool{'x': true, 'a': false})
	result2 := m1.ConnectedPositions(IPt(5, 1))
	assert.Equal(t, expected2, result2)
}

func TestDbSql(t *testing.T) {
	db := ConnectToDbSql()
	id := uuid.New()
	InitializeIdInDbSql(db, id)
	UploadDataToDbSql(db, id, []byte("what do you mean"))
	InspectDataFromDbSql(db)
	assert.True(t, true)
}

func TestDbHttp(t *testing.T) {
	id := uuid.New()
	// id, err := uuid.Parse("550e8400-e29b-41d4-a716-446655440002")
	// Check(err)
	InitializeIdInDbHttp("test-user", 19, id)
	UploadDataToDbHttp("test-user", 19, id, []byte("mele 1"))
	UploadDataToDbHttp("test-user", 19, id, []byte("mele 2"))
	UploadDataToDbHttp("test-user", 19, id, []byte("mele totusi, da -------"))

	SetUserDataHttp("test-user1", "test-data1")
	data := GetUserDataHttp("test-user1")
	assert.Equal(t, "test-data1", data)

	SetUserDataHttp("test-user1", "test-data2")
	data = GetUserDataHttp("test-user1")
	assert.Equal(t, "test-data2", data)
}

func TestSeralizationInt(t *testing.T) {
	var x Int
	buf := new(bytes.Buffer)
	x = I64(103)
	Serialize(buf, x)
	var y Int
	Deserialize(buf, &y)
	assert.Equal(t, x, y)
}

func TestSeralizationStruct(t *testing.T) {
	type Struct struct {
		A Int
		B Int
	}
	var x Struct
	x.A = I64(103)
	x.B = I64(93772)
	buf := new(bytes.Buffer)
	Serialize(buf, x)
	var y Struct
	Deserialize(buf, &y)
	assert.Equal(t, x, y)
}

func TestSeralizationSlice(t *testing.T) {
	var x []int64
	x = make([]int64, 3)
	x[0], x[1], x[2] = 3, 12, 9
	buf := new(bytes.Buffer)
	Serialize(buf, x)
	var y []int64
	Deserialize(buf, &y)
	assert.True(t, slices.Equal(x, y))
}

func TestSeralizationSliceOfSlice(t *testing.T) {
	type T1 struct {
		X []int64
	}
	type T2 struct {
		X []T1
		Y int64
	}
	var x T2
	x.X = make([]T1, 2)
	x.X[0].X = make([]int64, 2)
	x.X[0].X[0] = 39
	x.X[0].X[1] = 927
	x.X[1].X = make([]int64, 2)
	x.X[1].X[0] = 3333
	x.X[1].X[1] = -3
	x.Y = 931

	buf := new(bytes.Buffer)
	Serialize(buf, x)
	var y T2
	Deserialize(buf, &y)
	assert.Equal(t, x, y)
}
