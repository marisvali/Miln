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
		Extract()
	} else if action == "updatever" {
		UpdateVersion()
	}
}
