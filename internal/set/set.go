package set

import "fmt"

type Set[T comparable] struct {
	m map[T]struct{}
}

func New[T comparable](vals ...T) *Set[T] {
	t := &Set[T]{
		m: make(map[T]struct{}),
	}

	for _, val := range vals {
		t.Add(val)
	}

	return t
}

func (s Set[T]) Add(t T) {
	s.m[t] = struct{}{}
}

func (s Set[T]) Contains(t T) bool {
	_, ok := s.m[t]
	return ok
}

func (s Set[T]) Size() int {
	return len(s.m)
}

func (s Set[T]) Values() []string {
	vals := []string{}
	for k := range s.m {
		vals = append(vals, fmt.Sprint(k))
	}
	return vals
}
