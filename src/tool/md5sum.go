package tool

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
)

//Md5sumByte calculates md5 checksum of []byte.
func Md5sumByte(b1File []byte) (s0Md5sum string) {
	s0Md5sum = fmt.Sprintf("%x", md5.Sum(b1File))
	return
}

//Md5sumFile calculates md5 checksum of file.
func Md5sumFile(s0FilePath string) (s0Md5sum string) {
	f, e := os.Open(s0FilePath)
	defer f.Close()
	CheckError(e)
	b1File, e := ioutil.ReadAll(f)
	CheckError(e)
	s0Md5sum = Md5sumByte(b1File)
	return
}
