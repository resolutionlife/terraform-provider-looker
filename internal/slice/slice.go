package slice

import (
	"strings"

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

// Delete removes the elements in toDelete from slice s.
func Delete(s []string, toDelete []string) (str []string) {
	for i := range s {
		if !Contains(toDelete, s[i]) {
			str = append(str, s[i])
		}
	}
	return
}

// Contains returns true is the element v is contained in the slice s.
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

// Predicate defines the expected filter functions Filter accepts
type Predicate[T constraints.Ordered] func(elem T) bool

// StartsWithString returns true if the predicate string starts with the provided prefix
func StartsWithString(prefix string) Predicate[string] {
	return func(elem string) bool {
		return strings.HasPrefix(elem, prefix)
	}
}

// Filter takes a generic slice and a predicate func and returns a slice of all elements in the original slice that satisfy the predicate
func Filter[T constraints.Ordered](slice []T, predicateFunc Predicate[T]) []T {
	filtered := make([]T, 0)
	for _, elem := range slice {
		if predicateFunc(elem) {
			filtered = append(filtered, elem)
		}
	}
	return filtered
}
