package tool

import (
	//"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

//TransB1ToS1 transforms 1d-byte ([]byte) into 0d-string (string) (utf-8).
func TransB1ToS1(b1d []byte) (s0d string) {
	b1 := b1d
	s1 := make([]string, 0)
	for len(b1) > 0 {
		r0, i0Size := utf8.DecodeRune(b1)
		s0 := strconv.QuoteRune(r0)
		s1 = append(s1, strings.Trim(s0, "'"))
		b1 = b1[i0Size:]
	}
	s0d = strings.Join(s1, "")
	return
}

//TransS0ToS2 transforms 0d-string (string) into 2d-string ([][]string).
func TransS0ToS2(s0d string, sep1, sep2 string) (s2d [][]string) {
	s1d := strings.Split(s0d, sep1)
	s2d = make([][]string, 0)
	for _, v := range s1d {
		if v != "" {
			s2d = append(s2d, strings.Split(v, sep2))
		}
	}
	return
}

//TransB1ToS2 transforms 1d-byte ([]byte) into 2d-string ([][]string).
func TransB1ToS2(b1d []byte, sep1, sep2 string) (s2d [][]string) {
	s1d := strings.Split(string(b1d), sep1)
	s2d = make([][]string, 0)
	for _, v := range s1d {
		if v != "" {
			s2d = append(s2d, strings.Split(strings.Trim(v, "\r"), sep2)) // "\r" for windows
		}
	}
	return
}

//TransS2ToB1 transforms 2d-string ([][]string) into 1d-byte ([]byte).
func TransS2ToB1(s2d [][]string, sep1, sep2 string) (b1d []byte) {
	s1d := make([]string, len(s2d))
	for k, v := range s2d {
		s1d[k] = strings.Join(v, sep2)
	}
	b1d = []byte(strings.Join(s1d, sep1))
	return
}

//TransS2ToM1 transforms 2d-string ([][]string) into 1d-map (map[string][]string).
func TransS2ToM1(s2d [][]string, i0Key int) (m1d map[string][]string) {
	m1d = make(map[string][]string)
	for _, s1d := range s2d {
		m1d[s1d[i0Key]] = s1d[:i0Key] // not append(s1d[:i0Key], s1d[i0Key+1:]...), which changes s1d
		m1d[s1d[i0Key]] = append(m1d[s1d[i0Key]], s1d[i0Key+1:]...)
		//fmt.Println("m1d:", m1d)
	}
	return
}

//TransS2 swaps 1st and 2nd dimension of 2d-string ([][]string).
func TransS2(s2d [][]string) (s2dTrans [][]string) {
	s2dTrans = make([][]string, len(s2d[0]))
	for k2, _ := range s2d[0] {
		s1d := make([]string, len(s2d))
		for k1, _ := range s2d {
			s1d[k1] = s2d[k1][k2]
		}
		s2dTrans[k2] = s1d
	}
	return
}

//TransM1S0ToS1 transforms 1d-map (map[string]string) into 1d-string ([]string) (in keys order).
func TransM1S0ToS1(m1d map[string]string) (s1d []string) {
	s1Key := make([]string, len(m1d))
	for k, _ := range m1d {
		s1Key = append(s1Key, k)
	}
	sort.Strings(s1Key)
	s1d = make([]string, len(s1Key))
	for i, v := range s1Key {
		s1d[i] = m1d[v]
	}
	return
}

//TransM2S1 swaps 1st and 2nd dimension of 2d-map (map[string]map[string][]string).
func TransM2S1(m12 map[string]map[string][]string) (m21 map[string]map[string][]string) {
	s1Key1 := make([]string, 0)
	s1Key2 := make([]string, 0)
	s0Sep := "-=-"
	s1Key12 := make([]string, 0)
	for k1, v1 := range m12 { // find keys from all dimensions
		if !FindS1(s1Key1, k1) {
			s1Key1 = append(s1Key1, k1)
		}
		for k2, _ := range v1 {
			if !FindS1(s1Key2, k2) {
				s1Key2 = append(s1Key2, k2)
			}
			s0Key12 := strings.Join([]string{k1, k2}, s0Sep)
			s1Key12 = append(s1Key12, s0Key12)
		}
	}
	sort.Strings(s1Key1)
	sort.Strings(s1Key2)
	sort.Strings(s1Key12)
	m21 = make(map[string]map[string][]string)
	for _, k2 := range s1Key2 {
		m1 := make(map[string][]string)
		for _, k1 := range s1Key1 {
			s0Key12 := strings.Join([]string{k1, k2}, s0Sep)
			if FindS1(s1Key12, s0Key12) {
				m1[k1] = m12[k1][k2]
			}
		}
		m21[k2] = m1
	}
	return
}

//TransM2S0 swaps 1st and 2nd dimension of (filled or unfilled) 2d-map (map[string]map[string]string).
func TransM2S0(m12 map[string]map[string]string) (m21 map[string]map[string]string) {
	s1Key1 := make([]string, 0)
	s1Key2 := make([]string, 0)
	s0Sep := "-=-"
	s1Key12 := make([]string, 0)
	for k1, v1 := range m12 { // find keys from all dimensions
		if !FindS1(s1Key1, k1) {
			s1Key1 = append(s1Key1, k1)
		}
		for k2, _ := range v1 {
			if !FindS1(s1Key2, k2) {
				s1Key2 = append(s1Key2, k2)
			}
			s0Key12 := strings.Join([]string{k1, k2}, s0Sep)
			s1Key12 = append(s1Key12, s0Key12)
		}
	}
	sort.Strings(s1Key1)
	sort.Strings(s1Key2)
	sort.Strings(s1Key12)
	m21 = make(map[string]map[string]string)
	for _, k2 := range s1Key2 {
		m1 := make(map[string]string)
		for _, k1 := range s1Key1 {
			s0Key12 := strings.Join([]string{k1, k2}, s0Sep)
			if FindS1(s1Key12, s0Key12) {
				m1[k1] = m12[k1][k2]
			}
		}
		m21[k2] = m1
	}
	return
}

//TransM3S0Filled changes dimension order of filled 3d-map (map[string]map[string]map[string]string).
func TransM3S0Filled(m123 map[string]map[string]map[string]string, s0DimOrder string) (m3d map[string]map[string]map[string]string) {
	key1 := make([]string, 0)
	key2 := make([]string, 0)
	key3 := make([]string, 0)
	for k1, v1 := range m123 { // find keys from all dimensions
		if !FindS1(key1, k1) {
			key1 = append(key1, k1)
		}
		for k2, v2 := range v1 {
			if !FindS1(key2, k2) {
				key2 = append(key2, k2)
			}
			for k3, _ := range v2 {
				if !FindS1(key3, k3) {
					key3 = append(key3, k3)
				}
			}
		}
	}
	sort.Strings(key1)
	sort.Strings(key2)
	sort.Strings(key3)
	m3d = make(map[string]map[string]map[string]string)
	switch s0DimOrder {
	case "123": //1st <> 2nd
		m3d = m123
	case "213": //1st <> 2nd
		for _, k2 := range key2 {
			m13 := make(map[string]map[string]string)
			for _, k1 := range key1 {
				m13[k1] = m123[k1][k2]
			}
			m3d[k2] = m13
		}
	case "132": // 2nd <> 3rd
		for _, k1 := range key1 {
			m32 := make(map[string]map[string]string)
			for _, k3 := range key3 {
				m2 := make(map[string]string)
				for _, k2 := range key2 {
					m2[k2] = m123[k1][k2][k3]
				}
				m32[k3] = m2
			}
			m3d[k1] = m32
		}
	case "231": // 1st <> 2nd <> 3rd
		for _, k2 := range key2 {
			m31 := make(map[string]map[string]string)
			for _, k3 := range key3 {
				m1 := make(map[string]string)
				for _, k1 := range key1 {
					m1[k1] = m123[k1][k2][k3]
				}
				m31[k3] = m1
			}
			m3d[k2] = m31
		}
	default:
		log.Fatal("wrong dimension order")
	}
	return
}

//TransM2S0ToS2 transforms 2d-map (map[key1]map[key2]cell) into 2d-string ([][]string):
//if s0DimOrder=="12", return ([[id, key2a, key2b, ...], [key1a, cell11, cell12, ...], [key1b, cell21, cell22, ...], ...]);
//if s0DimOrder=="21", return ([[id, key1a, key1b, ...], [key2a, cell11, cell12, ...], [key2b, cell21, cell22, ...], ...]);
//where cell will be "NA" if without such key.
func TransM2S0ToS2(m2d map[string]map[string]string, s0Id string, s0DimOrder string) (s2d [][]string) {
	key1 := make([]string, 0)
	key2 := make([]string, 0)
	for k1, v1 := range m2d { // find keys from all dimensions
		key1 = append(key1, k1)
		for k2, _ := range v1 {
			if !FindS1(key2, k2) {
				key2 = append(key2, k2)
			}
		}
	}
	sort.Strings(key1)
	sort.Strings(key2)
	for _, k1 := range key1 { // fulfill map[key1][key2]cell ("NA" for missing cell)
		key2nd := make([]string, 0)
		for k2nd, _ := range m2d[k1] {
			key2nd = append(key2nd, k2nd)
		}
		for _, k2 := range key2 {
			if !FindS1(key2nd, k2) { // without such key
				m2d[k1][k2] = "NA"
			}
		}
	}
	switch s0DimOrder {
	case "12":
		s2d = make([][]string, len(key1)+1)
		s2d[0] = append([]string{s0Id}, key2...)
		for i1, k1 := range key1 {
			s1d := make([]string, len(key2)+1)
			s1d[0] = k1
			for i2, k2 := range key2 {
				s1d[i2+1] = m2d[k1][k2]
			}
			s2d[i1+1] = s1d
		}
	case "21":
		s2d = make([][]string, len(key2)+1)
		s2d[0] = append([]string{s0Id}, key1...)
		for i2, k2 := range key2 {
			s1d := make([]string, len(key1)+1)
			s1d[0] = k2
			for i1, k1 := range key1 {
				s1d[i1+1] = m2d[k1][k2]
			}
			s2d[i2+1] = s1d
		}
	}
	return
}

//CheckLengthS2 checks length of 2nd-dim of 2d-string ([1st-dim][2nd-dim]string).
func CheckLengthS2(s2d [][]string) {
	iLength := len(s2d[0])
	for _, v := range s2d[1:] {
		if len(v) != iLength {
			log.Fatal("different length of 2nd-dim in 2-slice")
		}
	}
	return
}

//TransS2ToM2S0 transforms 2d-string ([][]string) into 2d-map (map[key1]map[key2]cell):
//if s0DimOrder=="12", from ([[id, key2a, key2b, ...], [key1a, cell11, cell12, ...], [key1b, cell21, cell22, ...], ...]);
//if s0DimOrder=="21", from ([[id, key1a, key1b, ...], [key2a, cell11, cell12, ...], [key2b, cell21, cell22, ...], ...]).
func TransS2ToM2S0(s2d [][]string, s0DimOrder string) (m2d map[string]map[string]string) {
	CheckLengthS2(s2d)
	key1 := make([]string, len(s2d)-1)
	key2 := s2d[0][1:]
	for i, v := range s2d[1:] {
		key1[i] = v[0]
	}
	m2d = make(map[string]map[string]string)
	if s0DimOrder == "21" {
		key1, key2 = key2, key1
	}
	for i1, k1 := range key1 {
		m1d := make(map[string]string)
		for i2, k2 := range key2 {
			switch s0DimOrder {
			case "12":
				m1d[k2] = s2d[i1+1][i2+1]
			case "21":
				m1d[k2] = s2d[i2+1][i1+1]
			default:
				log.Fatal("dimension-order should be \"12\" or \"21\"")
			}
		}
		m2d[k1] = m1d
	}
	return
}

//TransS2Skip skips some 1st-dim and 2nd-dim in 2d-string ([1st-dim][2nd-dim]string).
func TransS2Skip(s2d [][]string, s1Skip1st, s1Skip2nd []string) (s2dSkip [][]string) {
	CheckLengthS2(s2d)
	s1Name1st := make([]string, len(s2d))
	for i, v := range s2d {
		s1Name1st[i] = v[0]
	}
	s1Name2nd := s2d[0]
	i1Index1st := make([]int, 0)
	for _, v := range s1Skip1st {
		i1Index := IndexS1(s1Name1st, v)
		i1Index1st = append(i1Index1st, i1Index...)
	}
	i1Index2nd := make([]int, 0)
	for _, v := range s1Skip2nd {
		i1Index := IndexS1(s1Name2nd, v)
		i1Index2nd = append(i1Index2nd, i1Index...)
	}
	s2dSkip = make([][]string, 0)
	for i1, v1 := range s2d {
		if FindI1(i1Index1st, i1) {
			continue
		}
		s1d := make([]string, 0)
		for i2, v2 := range v1 {
			if FindI1(i1Index2nd, i2) {
				continue
			}
			s1d = append(s1d, v2)
		}
		s2dSkip = append(s2dSkip, s1d)
	}
	return
}

//SwapM1S0KeyValue swaps the key-value in 1d-map (map[key]value) with unique-value.
func SwapM1S0KeyValue(m1d map[string]string) (m1dTrans map[string][]string) {
	s1Key := make([]string, 0)
	s1Val := make([]string, 0)
	for k, v := range m1d {
		if !FindS1(s1Key, k) {
			s1Key = append(s1Key, k)
		}
		if !FindS1(s1Val, v) {
			s1Val = append(s1Val, v)
		}
	}
	m1dTrans = make(map[string][]string)
	for _, s0Val := range s1Val {
		s1Key2 := make([]string, 0)
		for k, v := range m1d {
			if v == s0Val {
				if !FindS1(s1Key2, k) {
					s1Key2 = append(s1Key2, k)
				}
			}
		}
		m1dTrans[s0Val] = s1Key2
	}
	return
}

//SetUnionM1S1ToS1 gets the union of all values from 1d-map ([ key1:[value1, value2, ...], key2:[value2, value3, ...] ]).
func SetUnionM1S1ToS1(m1d map[string][]string) (s1d []string) {
	s1d = make([]string, 0)
	for _, v := range m1d {
		s1d = append(s1d, v...)
	}
	s1d = SetS1(s1d)
	return
}

//SetInterM1S1ToS1 gets the intersection of all values from different keys of 1d-map ([ key1:[value1, value2, ...], key2:[value2, value3, ...] ]).
func SetInterM1S1ToS1(m1d map[string][]string) (s1d []string) {
	s1Value := make([]string, 0)
	for _, v := range m1d {
		s1Value = append(s1Value, v...)
	}
	s1Value = SetS1(s1Value)
	m1ValueCount := make(map[string]int)
	for _, v := range s1Value {
		m1ValueCount[v] = 0
	}
	for _, v1 := range m1d {
		for _, v2 := range v1 {
			m1ValueCount[v2] = m1ValueCount[v2] + 1
		}
	}
	i0Count := len(m1d)
	s1d = make([]string, 0)
	for k, v := range m1ValueCount {
		if v == i0Count {
			s1d = append(s1d, k)
		}
	}
	sort.Strings(s1d)
	return
}

//KeyM1S0 gets keys of 1d-map (map[string]string).
func KeyM1S0(m1 map[string]string) (s1 []string) {
	s1 = make([]string, 0)
	for k, _ := range m1 {
		s1 = append(s1, k)
	}
	return
}

//KeyM1S1 gets keys of 1d-map (map[string][]string).
func KeyM1S1(m1 map[string][]string) (s1 []string) {
	s1 = make([]string, 0)
	for k, _ := range m1 {
		s1 = append(s1, k)
	}
	sort.Strings(s1)
	return
}

//SubS0 gets substring from string.
func SubS0(s0d string, i0Start, i0End int) (s0Sub string) {
	s0Sub = strings.Join(strings.Split(s0d, "")[i0Start:i0End], "")
	return
}

//SubS1 gets substrings from 1d-string ([]string).
func SubS1(s1d []string, i0Start, i0End int) (s1Sub []string) {
	s1Sub = make([]string, 0)
	for _, s0d := range s1d {
		s0Sub := SubS0(s0d, i0Start, i0End)
		if !FindS1(s1Sub, s0Sub) {
			s1Sub = append(s1Sub, s0Sub)
		}
	}
	return
}

//SubS1ToM1S1 transforms 1d-string ([]string) to map[sub-string-key]1d-sub-string-value.
func SubS1ToM1S1(s1d []string, i0KeyStart, i0KeyEnd, i0ValueStart, i0ValueEnd int) (m1KeyValue map[string][]string) {
	m1StringKey := make(map[string]string)
	for _, s0d := range s1d {
		s0Key := SubS0(s0d, i0KeyStart, i0KeyEnd)
		m1StringKey[s0d] = s0Key
	}
	m1KeyString := SwapM1S0KeyValue(m1StringKey)
	m1KeyValue = make(map[string][]string)
	for s0Key, s1d := range m1KeyString {
		m1KeyValue[s0Key] = SubS1(s1d, i0ValueStart, i0ValueEnd)
	}
	return
}
