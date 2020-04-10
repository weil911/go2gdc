package tool

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//WalkDir finds all files in the dir and sub-dir.
func WalkDir(s0Dir string) (s1FilePath []string) {
	s1FilePath = make([]string, 0)
	e := filepath.Walk(s0Dir, func(s0FilePath string, fileInfo os.FileInfo, e error) error {
		CheckError(e)
		if !fileInfo.IsDir() {
			s1FilePath = append(s1FilePath, s0FilePath)
		}
		return nil
	})
	CheckError(e)
	return
}

//MakeDir makes dir with path of file.
func MakeDir(s0FilePath string) {
	e := os.MkdirAll(string(path.Dir(s0FilePath)), 0755)
	CheckError(e)
	return
}

//MakeFile creates file, and create dir first if necessary.
func MakeFile(s0FilePath string) (f *os.File) {
	MakeDir(s0FilePath)
	f, e := os.Create(s0FilePath)
	CheckError(e)
	return
}

//SaveFile writes []byte into file.
func SaveFile(s0FilePath string, b1Body []byte) {
	MakeDir(s0FilePath)
	e := ioutil.WriteFile(s0FilePath, b1Body, 0644)
	CheckError(e)
	return
}

//ReadByteGz reads gz format []byte and writes it into unzipped []byte.
func ReadByteGz(b1File []byte) (b1d []byte) {
	readerGz, e := gzip.NewReader(bytes.NewReader(b1File))
	defer readerGz.Close()
	CheckError(e)
	b1d, e = ioutil.ReadAll(readerGz)
	CheckError(e)
	return
}

//ReadFile reads unzipped file into map[UnzippedFilePath][]byte (len=1).
func ReadFile(s0FilePath string) (m1FileByte map[string][]byte) {
	f, e := os.Open(s0FilePath)
	defer f.Close()
	CheckError(e)
	b1File, e := ioutil.ReadAll(f)
	CheckError(e)
	m1FileByte = make(map[string][]byte)
	m1FileByte[s0FilePath] = b1File
	return
}

//ReadFiles reads unzipped file into map[UnzippedFilePath][]byte (len>1).
func ReadFiles(s1FilePath []string) (m1FileByte map[string][]byte) {
	m1FileByte = make(map[string][]byte)
	for _, s0FilePath := range s1FilePath {
		f, e := os.Open(s0FilePath)
		defer f.Close()
		CheckError(e)
		b1File, e := ioutil.ReadAll(f)
		CheckError(e)
		m1FileByte[s0FilePath] = b1File
	}
	return
}

//ReadFileTarGz reads tar.gz file into map[UnzippedFilePath][]byte (len>=1).
func ReadFileTarGz(s0FilePath string) (m1FileByte map[string][]byte) {
	if path.Ext(s0FilePath) != ".gz" || path.Ext(strings.TrimSuffix(s0FilePath, path.Ext(s0FilePath))) != ".tar" {
		log.Fatal("filename does not end with \".tar.gz\".")
	}
	fileGz, e := os.Open(s0FilePath)
	defer fileGz.Close()
	CheckError(e)
	readerGz, e := gzip.NewReader(fileGz)
	defer readerGz.Close()
	CheckError(e)
	readerTar := tar.NewReader(readerGz)
	m1FileByte = make(map[string][]byte)
	for tarHeader, e := readerTar.Next(); e != io.EOF; tarHeader, e = readerTar.Next() {
		CheckError(e)
		b1File, e := ioutil.ReadAll(readerTar)
		CheckError(e)
		m1FileByte[tarHeader.Name] = b1File
	}
	return
}

//ReadByteTarGz reads tar.gz format []byte into map[UnzippedFilePath][]byte (len>=1).
func ReadByteTarGz(b1File []byte) (m1FileByte map[string][]byte) {
	readerGz, e := gzip.NewReader(bytes.NewReader(b1File))
	defer readerGz.Close()
	CheckError(e)
	readerTar := tar.NewReader(readerGz)
	m1FileByte = make(map[string][]byte)
	for tarHeader, e := readerTar.Next(); e != io.EOF; tarHeader, e = readerTar.Next() {
		CheckError(e)
		b1File, e := ioutil.ReadAll(readerTar)
		CheckError(e)
		m1FileByte[tarHeader.Name] = b1File
	}
	return
}

//Untar untars tar.gz file and returns a slice of UnzippedFilePath.
func Untar(s0TarGzPath string) (s1FilePath []string) {
	if path.Ext(s0TarGzPath) != ".gz" || path.Ext(strings.TrimSuffix(s0TarGzPath, path.Ext(s0TarGzPath))) != ".tar" {
		log.Fatal("filename does not end with \".tar.gz\".")
	}
	fileGz, e := os.Open(s0TarGzPath)
	defer fileGz.Close()
	CheckError(e)
	readerGz, e := gzip.NewReader(fileGz)
	defer readerGz.Close()
	CheckError(e)
	readerTar := tar.NewReader(readerGz)
	s1FilePath = make([]string, 0)
	for tarHeader, e := readerTar.Next(); e != io.EOF; tarHeader, e = readerTar.Next() {
		CheckError(e)
		s0FilePath := strings.TrimSuffix(s0TarGzPath, ".tar.gz") + "/" + tarHeader.Name
		s1FilePath = append(s1FilePath, s0FilePath)
		f := MakeFile(s0FilePath)
		defer f.Close()
		_, e = io.Copy(f, readerTar)
		CheckError(e)
		os.Chmod(s0FilePath, tarHeader.FileInfo().Mode().Perm())
	}
	return
}

//CsvRead reads csv file.
func CsvRead(s0FilePath string, r0Comment rune, r0Comma rune, i0Field int) (s2d [][]string) {
	f, e := os.Open(s0FilePath)
	defer f.Close()
	CheckError(e)
	readerCsv := csv.NewReader(f)
	readerCsv.Comment = r0Comment
	readerCsv.Comma = r0Comma
	readerCsv.FieldsPerRecord = i0Field
	s2d, e = readerCsv.ReadAll()
	CheckError(e)
	return
}

//CsvSave writes csv file.
func CsvSave(s0FilePath string, s2d [][]string, r0Comma rune) {
	f, e := os.OpenFile(s0FilePath, os.O_RDWR|os.O_CREATE, 0666)
	defer f.Close()
	CheckError(e)
	writerCsv := csv.NewWriter(f)
	writerCsv.Comma = r0Comma
	e = writerCsv.WriteAll(s2d)
	CheckError(e)
	return
}
