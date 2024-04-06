package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func CleanOldCode(ageThreshold time.Duration) {
	// keep files which lie inside threshold
	threshold := time.Now().Add(-ageThreshold)

	files, err := os.ReadDir(CodeStorageFolder)
	if err != nil {
		log.Println("[cleanup] failed traversing 'code' dir:", err)
	}

	count := 0
	for _, f := range files {
		// filepath: /tmp/code/1712425357672220917.go
		path := CodeStorageFolder + "/" + f.Name()
		ca, err := codeCreatedAt(f.Name())
		if err != nil {
			log.Printf("failed to get created date from file/dir '%v': %v", path, err)
			continue
		}
		if ca.Before(threshold) {
			err := os.RemoveAll(path)
			if err != nil {
				log.Printf("failed to delete file/dir '%v': %v", path, err)
			} else {
				count++
			}
		}
	}

	log.Println("[cleanup] deleted files:", count)
}

// fname = 1712425357672220917.go
func codeCreatedAt(fname string) (time.Time, error) {
	t := strings.Split(fname, ".go")[0]
	unixT, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(unixT / 1000), nil
}
