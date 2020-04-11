package tool

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

//CheckStatus checks http.Response.Status.
func CheckStatus(resp *http.Response) {
	if resp.Status == "404 NOT FOUND" {
		log.Fatal("Response.CheckStatus == ", resp.Status, ", which means no file satisfy your filter.")
	} else if resp.Status != "200 OK" {
		log.Fatal("Response.CheckStatus == ", resp.Status)
	}
	return
}

//Post posts request with payload in some format (e.g. json).
func Post(s0Url string, b1Body []byte, s0ContentType string) (resp *http.Response) {
	requ, e := http.NewRequest("POST", s0Url, bytes.NewReader(b1Body))
	requ.Header.Set("Content-Type", s0ContentType)
	client := &http.Client{}
	resp, e = client.Do(requ)
	CheckError(e)
	return
}

//ResponseToB1 turns *http.Response.Body into []byte.
func ResponseToB1(resp *http.Response) (b1Body []byte) {
	b1Body, e := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	CheckError(e)
	return
}
