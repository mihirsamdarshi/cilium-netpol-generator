package main

import "golang.org/x/exp/maps"

// Set is a struct representing a set of strings.
// It is implemented as a map[string]struct{}, since Go does not have a built-in set type.
type Set struct {
	_Inner map[string]struct{}
}

func NewSet() *Set {
	return &Set{_Inner: make(map[string]struct{})}
}

func (s *Set) AddLists(lists ...[]string) *Set {
	var exists struct{}
	for _, list := range lists {
		for _, item := range list {
			s._Inner[item] = exists
		}
	}
	return s
}

func (s *Set) ToList() []string {
	return maps.Keys(s._Inner)
}
