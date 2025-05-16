package container

type Set[T comparable] []T

func NewSet[T comparable](items ...T) Set[T] {
	set := make(Set[T], 0)
	for _, item := range items {
		set.Add(item)
	}
	return set
}

func (s *Set[T]) Add(item T) {
	for _, v := range *s {
		if v == item {
			return
		}
	}
	*s = append(*s, item)
}

func (s *Set[T]) Remove(item T) {
	for i, v := range *s {
		if v == item {
			*s = append((*s)[:i], (*s)[i+1:]...)
			return
		}
	}
}

func (s *Set[T]) Contains(item T) bool {
	for _, v := range *s {
		if v == item {
			return true
		}
	}
	return false
}
