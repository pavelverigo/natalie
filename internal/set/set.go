package set

type Set[T comparable] struct {
	m map[T]struct{}
}

func New[T comparable](cap int) *Set[T] {
	m := make(map[T]struct{}, cap)

	return &Set[T]{
		m: m,
	}
}

func (s *Set[T]) Contains(key T) bool {
	_, ok := s.m[key]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.m)
}

func (s *Set[T]) Set(key T) {
	s.m[key] = struct{}{}
}

func (s *Set[T]) Delete(key T) {
	delete(s.m, key)
}

func (s *Set[T]) Keys() []T {
	res := make([]T, len(s.m))
	i := 0
	for k := range s.m {
		res[i] = k
		i += 1
	}
	return res
}
