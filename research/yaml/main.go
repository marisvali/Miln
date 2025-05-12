package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/marisvali/miln/gamelib"
	"io"
	"os"
)

type A struct {
	A1 int `yaml:"A1"`
	A2 int `yaml:"A2"`
}

type Wave struct {
	SecondsSincePreviousWave int `yaml:"SecondsSincePreviousWave"`
	NGremlins                int `yaml:"NGremlins"`
	NHounds                  int `yaml:"NHounds"`
	NBau                     int `yaml:"NBau"`
}

type SpawnPortal struct {
	Waves []Wave `yaml:"Waves"`
}

type TestStruct struct {
	TestInt gamelib.Int   `yaml:"TestInt"`
	Param1  int           `yaml:"Param1"`
	Param2  int           `yaml:"Param2"`
	ArrInt  []int         `yaml:"ArrInt"`
	ArrObj  []A           `yaml:"ArrObj"`
	Portals []SpawnPortal `yaml:"Portals"`
}

func main() {
	var t TestStruct
	LoadYAML("test.yaml", &t)
	fmt.Printf("%+v", t)
	SaveYAML("output.yaml", &t)
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func LoadYAML(filename string, v any) {
	file, err := os.Open(filename)
	Check(err)
	data, err := io.ReadAll(file)
	Check(err)
	err = yaml.Unmarshal(data, v)
	Check(err)
}

func SaveYAML(filename string, v any) {
	data, err := yaml.Marshal(v)
	err = os.WriteFile(filename, data, 0644)
	Check(err)
}
