// Copyright 2020-present woodsshin. All rights reserved.
// Use of this source code is governed by MIT license.

package galanodes

import (
	"fmt"
	"log"
	"os"
	"time"
)

func WriteLog(str string) {
	// open output file
	fo, err := os.OpenFile("log.txt", os.O_APPEND, 0755)
	if err != nil {
		fo, err = os.Create("log.txt")
		if err != nil {
			log.Println(err)
			return
		}
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			log.Println(err)
		}
	}()

	line := fmt.Sprintf("%v %v\n", time.Now().UTC().Format("2006-01-02 15:04:05.000"), str)
	if _, err := fo.WriteString(line); err != nil {
		log.Println(err)
	}
}
