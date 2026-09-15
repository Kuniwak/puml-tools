package sets

import "sort"

type Set map[string]struct{}

func New(items ...string) Set {
	s := make(Set, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func (s Set) Add(item string) {
	s[item] = struct{}{}
}

func (s Set) AddAll(other Set) {
	for item := range other {
		s[item] = struct{}{}
	}
}

func (s Set) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s Set) Sorted() []string {
	out := make([]string, 0, len(s))
	for item := range s {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func Intersects(a, b Set) bool {
	small, large := a, b
	if len(large) < len(small) {
		small, large = large, small
	}
	for item := range small {
		if large.Contains(item) {
			return true
		}
	}
	return false
}
