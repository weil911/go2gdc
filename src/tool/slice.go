package tool

import (
	"log"
	"sort"
	"strings"
)

//RepeatPickI1 picks "all", "one" or "none" of the same element in []int.
func RepeatPickI1(i1Repeat []int, s0PickType string) (i1Picked []int) {
	i1Picked = make([]int, 0)
	switch s0PickType {
	case "all", "":
		i1Picked = i1Repeat
	case "one":
		for _, v := range i1Repeat {
			if !FindI1(i1Picked, v) {
				i1Picked = append(i1Picked, v)
			}
		}
	case "none":
		i1Exclusion := make([]int, 0)
		i1Set := make([]int, 0)
		for _, v := range i1Repeat {
			if !FindI1(i1Set, v) {
				i1Set = append(i1Set, v)
			} else {
				i1Exclusion = append(i1Exclusion, v)
			}
		}
		for _, v := range i1Set {
			if !FindI1(i1Exclusion, v) {
				i1Picked = append(i1Picked, v)
			}
		}
	default:
		log.Fatalf("wrong pick type: %s, should be one of \"all\" (\"\"), \"one\" and \"none\".", s0PickType)
	}
	return
}

//RepeatPickS1 picks "all", "one" or "none" of the same element in []string.
func RepeatPickS1(s1Repeat []string, s0PickType string) (s1Picked []string) {
	s1Picked = make([]string, 0)
	switch s0PickType {
	case "all", "":
		s1Picked = s1Repeat
	case "one":
		for _, v := range s1Repeat {
			if !FindS1(s1Picked, v) {
				s1Picked = append(s1Picked, v)
			}
		}
	case "none":
		s1Exclusion := make([]string, 0)
		s1Set := make([]string, 0)
		for _, v := range s1Repeat {
			if !FindS1(s1Set, v) {
				s1Set = append(s1Set, v)
			} else {
				s1Exclusion = append(s1Exclusion, v)
			}
		}
		for _, v := range s1Set {
			if !FindS1(s1Exclusion, v) {
				s1Picked = append(s1Picked, v)
			}
		}
	default:
		log.Fatalf("wrong pick type: %s, should be one of \"all\" (\"\"), \"one\" and \"none\".", s0PickType)
	}
	return
}

//EqualI1 compares 2 []int.
func EqualI1(x, y []int) (b bool) {
	if len(x) == len(y) {
		b = true
		for i, v := range x {
			if v != y[i] {
				b = false
				break
			}
		}
	}
	return
}

//EqualS1 compares 2 []string.
func EqualS1(x, y []string) (b bool) {
	if len(x) == len(y) {
		b = true
		for i, v := range x {
			if v != y[i] {
				b = false
				break
			}
		}
	}
	return
}

//FindI1 finds element in []int.
func FindI1(x1d []int, x int) (b bool) {
	for _, v := range x1d {
		if v == x {
			b = true
			break
		}
	}
	return
}

//FindS1 finds element in []string.
func FindS1(x1d []string, x string) (b bool) {
	for _, v := range x1d {
		if v == x {
			b = true
			break
		}
	}
	return
}

//IndexI1 gets indexes of element in []int.
func IndexI1(x1d []int, x int) (i1Index []int) {
	i1Index = make([]int, 0)
	for i, v := range x1d {
		if v == x {
			i1Index = append(i1Index, i)
		}
	}
	return
}

//IndexS1 gets indexes of element in []string.
func IndexS1(x1d []string, x string) (i1Index []int) {
	i1Index = make([]int, 0)
	for i, v := range x1d {
		if v == x {
			i1Index = append(i1Index, i)
		}
	}
	return
}

//UniteS1 unites elements of []string into 1st element (e.g. for multi-columns-id).
func UniteS1(s1d []string, i1Index []int, s0Sep string) (s1dUnite []string) {
	s1Id := make([]string, len(i1Index))
	for k, v := range i1Index {
		s1Id[k] = s1d[v]
	}
	s1dUnite = make([]string, 0)
	s1dUnite = append(s1dUnite, strings.Join(s1Id, s0Sep))
	for k, v := range s1d {
		if !FindI1(i1Index, k) {
			s1dUnite = append(s1dUnite, v)
		}
	}
	return
}

//UniteS2 unites elements of every []string into 1st element (in 2nd dimension) (e.g. for multi-columns-id).
func UniteS2(s2d [][]string, i1Index []int, s0Sep string) (s2dUnite [][]string) {
	s2dUnite = make([][]string, len(s2d))
	for k, s1d := range s2d {
		s2dUnite[k] = UniteS1(s1d, i1Index, s0Sep)
	}
	return
}

//SetS1 gets unique members of []string.
func SetS1(s1d []string) (s1dSet []string) {
	s1dSet = make([]string, 0)
	for _, v := range s1d {
		if !FindS1(s1dSet, v) {
			s1dSet = append(s1dSet, v)
		} else {
			continue
		}
	}
	sort.Strings(s1dSet)
	return
}
