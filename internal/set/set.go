package set

type Set[T comparable] struct {
	m map[T]struct{}
}

func New[T comparable]() *Set[T] {
	return &Set[T]{
		m: make(map[T]struct{}),
	}
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
