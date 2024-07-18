package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: analysis.exe <download/extract>")
		return
	}
	action := os.Args[1]
	if action == "download" {
		DownloadRecordings()
	} else if action == "extract" {
		dir := os.Args[2]
		Extract(dir)
	} else if action == "updatever" {
		// start := os.Args[2]
		// end := os.Args[3]
		// version := os.Args[4]
		start := "2024-07-14 10:01:21"
		end := "2024-07-14 10:17:16"
		version := "4"
		UpdateVersion(start, end, version)
	}
}
