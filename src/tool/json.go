package tool

import (
	"github.com/json-iterator/go"
	"log"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

//CheckJson valids json format []byte.
func CheckJson(b1Json []byte) {
	if !json.Valid(b1Json) {
		log.Fatal("json.Valid return False")
	}
	return
}
