package container

import (
	"sync"
	"sync/atomic"
)

// 数组实现的环形链表
type RingBuffMIMO[T any] struct {
	buff  []T
	head  int32
	tail  int32
	cap   int32
	len   int32
	zeroV T
	mut   sync.RWMutex
}

func NewRingBuffMIMO[T any](size int32) *RingBuffMIMO[T] {
	buf := RingBuffMIMO[T]{
		buff: make([]T, size),
		cap:  size,
	}
	return &buf
}

// push到最后
func (q *RingBuffMIMO[T]) Push(val T) {
	q.mut.Lock()
	defer q.mut.Unlock()

	q.tail = (q.tail + 1) % q.cap
	if q.head == q.tail {
		// need more memory
		buf := make([]T, q.cap*2)
		copy(buf[:], q.buff[q.head:])
		copy(buf[q.cap-q.head:], q.buff[:q.tail])
		q.buff = buf
		q.head = 0
		q.tail = q.cap
		q.cap = int32(len(buf))
	}
	atomic.AddInt32(&q.len, 1)
	q.buff[q.tail] = val
}

func (q *RingBuffMIMO[T]) Len() int32 {
	return atomic.LoadInt32(&q.len)
}

func (q *RingBuffMIMO[T]) Empty() bool {
	return q.Len() == 0
}

func (q *RingBuffMIMO[T]) IsFull() bool {
	return q.Len() == atomic.LoadInt32(&q.cap)
}

// 弹出最前面的一个
func (q *RingBuffMIMO[T]) Pop() (T, bool) {
	if q.Empty() {
		return q.zeroV, false
	}
	q.mut.Lock()
	defer q.mut.Unlock()

	q.head = (q.head + 1) % q.cap
	v := q.buff[q.head]
	q.buff[q.head] = q.zeroV
	atomic.AddInt32(&q.len, -1)
	return v, true
}
