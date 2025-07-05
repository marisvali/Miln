package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: analysis.exe download")
		return
	}
	action := os.Args[1]
	if action == "download" {
		DownloadRecordings()
	}
}
