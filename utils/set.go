package utils

// Set 实现集合（线程不安全）
type Set interface {
	Add(key string)
	Contains(key string) bool
	Len() int
	List() []string
	For(func(key string) string) []string
}

type stringSet struct {
	set map[string]struct{}
}

func (s *stringSet) For(f func(key string) string) []string {
	slice := make([]string, 0, len(s.set))
	for k, _ := range s.set {
		slice = append(slice, f(k))
	}
	return slice
}

func (s *stringSet) List() []string {
	slice := make([]string, 0, len(s.set))
	for k, _ := range s.set {
		slice = append(slice, k)
	}
	return slice
}

func NewStringSet(size int) Set {
	stringSet := &stringSet{
		set: make(map[string]struct{}, size),
	}
	return stringSet
}

func (s *stringSet) Add(key string) {
	s.set[key] = struct{}{}
}

func (s *stringSet) Contains(key string) bool {
	_, ok := s.set[key]
	return ok
}

func (s *stringSet) Len() int {
	return len(s.set)
}
