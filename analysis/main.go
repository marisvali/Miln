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
		start := "2024-07-14 08:31:44"
		end := "2024-07-14 09:12:26"
		version := "3"
		UpdateVersion(start, end, version)
	}
}
