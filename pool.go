package container

import "sync"

type Pool[T any] struct {
	sync.Pool
}

func NewPool[T any](fn func() T) Pool[T] {
	return Pool[T]{
		Pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.Pool.Get().(T)
}

func (p *Pool[T]) Put(t T) {
	p.Pool.Put(t)
}
