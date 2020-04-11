package tool

import (
	"log"
)

//CheckError checks the error.
func CheckError(e error) {
	if e != nil {
		log.Fatal(e)
	}
	return
}
