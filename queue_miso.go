package container

import (
	"sync/atomic"
)

type iMISOQueueNodePool[T any] interface {
	Get() *node[T]
	Put(n *node[T])
}

// type MISOQueueNodePool[T any] struct {
// 	sync.Pool
// }

// func (p *MISOQueueNodePool[T]) Get() *node[T] {
// 	return p.Pool.Get().(*node[T])
// }

// func (p *MISOQueueNodePool[T]) Put(n *node[T]) {
// 	p.Pool.Put(n)
// }

// 多进单出的队列，可以在多个goruntine中push，只能在一个goruntine中pop
type MISOQueue[T any] struct {
	pool Pool[*node[T]]
	head atomic.Pointer[node[T]]
	tail atomic.Pointer[node[T]]
	zero T // 0值
}

func NewMISOQueue[T any](opt ...MISOQueueOption[T]) *MISOQueue[T] {
	q := MISOQueue[T]{}
	for _, o := range opt {
		o(&q)
	}
	q.pool = NewPool(func() *node[T] {
		return &node[T]{}
	})

	n := q.pool.Get()
	q.head.Store(n)
	q.tail.Store(n)
	return &q
}

type MISOQueueOption[T any] func(*MISOQueue[T])

// func WithMISOQueueNodePool[T any](pool iMISOQueueNodePool[T]) MISOQueueOption[T] {
// 	return func(m *MISOQueue[T]) {
// 		m.pool = pool
// 	}
// }

// type defMISOQueueNodePool[T any] struct{}

// func (p *defMISOQueueNodePool[T]) Get() *node[T] {
// 	return &node[T]{}
// }

// func (p *defMISOQueueNodePool[T]) Put(*node[T]) {}

type node[T any] struct {
	next atomic.Pointer[node[T]]
	val  T
}

// Enqueue 将val推入到队列尾部.
// Enqueue 可以在多个goruntine中并发调用
func (q *MISOQueue[T]) Enqueue(val T) {
	n := q.pool.Get()
	n.val = val
	prev := q.tail.Swap(n)
	prev.next.Store(n)
}

// Dequeue从队列头将元素弹出，若队列为空，会返回false
// 只能在同一个goruntime调用Dequeue
func (q *MISOQueue[T]) Dequeue() (T, bool) {
	head := q.head.Load()
	next := head.next.Load()

	var v T
	if next != nil {
		q.head.Store(next)
		v = next.val
		next.val = q.zero

		head.next.Store(nil)
		q.pool.Put(head)
		return v, true
	}
	return v, false
}

func (q *MISOQueue[T]) IsEmpty() bool {
	head := q.head.Load()
	return head.next.Load() == nil
}
