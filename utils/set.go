package utils

type Set[T string | int] map[T]struct{}

func NewSet[T string | int](size int) Set[T] {
	return make(Set[T], size)
}

func (s Set[T]) Contain(key T) bool {
	_, ok := s[key]
	return ok
}

func (s Set[T]) Add(key T) {
	s[key] = struct{}{}
}
