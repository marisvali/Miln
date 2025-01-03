package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type A struct {
	A1 int
	A2 int
}

type Wave struct {
	SecondsSincePreviousWave int
	NGremlins                int
	NHounds                  int
	NBau                     int
}

type SpawnPortal struct {
	Waves []Wave
}

type TestStruct struct {
	Param1  int
	Param2  int
	ArrInt  []int
	ArrObj  []A
	Portals []SpawnPortal
}

func main() {
	var t TestStruct
	LoadJSON("test.json", &t)
	fmt.Printf("%+v", t)
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func LoadJSON(filename string, v any) {
	file, err := os.Open(filename)
	Check(err)
	data, err := io.ReadAll(file)
	Check(err)
	err = json.Unmarshal(data, v)
	Check(err)
}
