package slice

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/exp/constraints"
)

// Diff checks if s in contained within t and if t is contained within s. Diff returns all elements that are not in the intersection of s and t.
func Diff(s, t []string) []string {
	var diff = []string{}

	for _, sval := range s {
		if !Contains(t, sval) {
			diff = append(diff, sval)
		}
	}
	for _, tval := range t {
		if !Contains(s, tval) {
			diff = append(diff, tval)
		}
	}

	return diff
}

// LeftDiff checks if t is contained within s and returns all elements not is s.
func LeftDiff(s, t []string) []string {
	var diff = []string{}

	for _, tval := range t {
		if !Contains(s, tval) {
			diff = append(diff, tval)
		}
	}

	return diff
}

func Delete(s []string, toDelete []string) (str []string) {
	for i := range s {
		if !Contains(toDelete, s[i]) {
			str = append(str, s[i])
		}
	}
	return
}

func Contains[T comparable](s []T, v T) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}

/*
UnorderedEqual determines if two slices of the same type contain equal elements ignoring ordering
the ordered constraint ensures only ints, floats and strings are usable with this function.
*/
func UnorderedEqual[T constraints.Ordered](a, b []T) bool {
	return cmp.Equal(a, b, cmpopts.SortSlices(func(a, b T) bool { return a < b }))
}
