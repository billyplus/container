package container

type Stack[T comparable] []T

func NewStack[T comparable](items ...T) Stack[T] {
	set := make(Stack[T], 0, len(items))
	for _, item := range items {
		set.Push(item)
	}
	return set
}

func NewStackWithCap[T comparable](maxCap int) Stack[T] {
	set := make(Stack[T], 0, maxCap)
	return set
}

func (s *Stack[T]) Push(item T) {
	for _, v := range *s {
		if v == item {
			return
		}
	}
	*s = append(*s, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(*s) == 0 {
		var t T
		return t, false
	}
	t := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return t, true
}

// 获取stack的长度
func (s *Stack[T]) Len() int {
	return len(*s)
}
