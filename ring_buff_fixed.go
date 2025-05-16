package container

// 固定大小的环形数据表，插入新元素时自动覆盖最早插入的数据，保持总数据长度不变，多协程写不安全
type RingBuffFixed[T any] struct {
	buff  []T
	head  int32
	tail  int32
	cap   int32
	len   int32
	zeroV T
}

// 固定大小的环形数据表，插入新元素时自动覆盖最早插入的数据，保持总数据长度不变，多协程写不安全
// 用来于保存最新N个数据的场景
func NewRingBuffFixed[T any](size int32) *RingBuffFixed[T] {
	buf := RingBuffFixed[T]{
		buff: make([]T, size),
		cap:  size,
	}
	return &buf
}

// 当元素满了覆盖最前面的
func (q *RingBuffFixed[T]) Push(val T) {
	q.buff[q.tail] = val
	q.tail = (q.tail + 1) % q.cap
	if q.len == q.cap {
		// 满了
		q.head = q.tail
	} else {
		q.len += 1
	}
}

func (q *RingBuffFixed[T]) Len() int {
	return int(q.len)
}

func (q *RingBuffFixed[T]) Empty() bool {
	return q.Len() == 0
}

func (q *RingBuffFixed[T]) GetAll() []T {
	if q.Empty() {
		return nil
	}
	count := int32(q.Len())

	buff := make([]T, 0, count)
	var pos int32
	for i := int32(0); i < count; i++ {
		pos = (q.head + i) % q.cap
		buff = append(buff, q.buff[pos])
	}
	return buff
}

// 获取最后插入的数据，
func (q *RingBuffFixed[T]) Last() (T, bool) {
	if q.Empty() {
		return q.zeroV, false
	}
	tail := (q.tail + q.cap - 1) % q.cap

	return q.buff[tail], true
}
