package main

import (
	"errors"
	"fmt"
	"html"
	"log"
	"os"
	"time"
)

func WriteToTempFile(b []byte) (string, error) {
	// convert HTML escaped
	//    cmd.Stdout = &amp;stdout
	//  to
	//   cmd.Stdout = &stdout
	unscaped := html.UnescapeString(string(b))

	// using nano second to avoid filename collision in highly concurrent requests
	id := fmt.Sprintf("%v", time.Now().UnixNano())
	err := os.WriteFile(codeFile(id), []byte(unscaped), 0777)

	if err != nil && errors.Is(err, os.ErrNotExist) {
		// may be the folder absent. so trying to create it
		err = os.Mkdir(CodeStorageFolder, os.ModePerm)
		if err != nil {
			log.Println("failed creating folder:", CodeStorageFolder)
			return id, err
		}
		log.Println("created folder:", CodeStorageFolder)
		// second attempt
		err = os.WriteFile(codeFile(id), []byte(unscaped), 0777)
	}

	return id, err
}

func codeFile(id string) string {
	return fmt.Sprintf("%v/%v.go", CodeStorageFolder, id)
}
