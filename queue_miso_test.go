package container

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMISOQueue(t *testing.T) {
	suite.Run(t, new(suiteMISOQueue))
}

type suiteMISOQueue struct {
	suite.Suite
}

func (st *suiteMISOQueue) TestPoolQueue_PushPop() {
	st.testQueue_PushPop()
}

func (st *suiteMISOQueue) TestQueue_PushPop() {
	st.testQueue_PushPop()
}

func (st *suiteMISOQueue) testQueue_PushPop() {
	q := NewMISOQueue[int]()

	q.Enqueue(1)
	q.Enqueue(2)
	v, ok := q.Dequeue()
	st.Equal(true, ok, "pop 1")
	st.Equal(1, v, "pop 1")
	v, ok = q.Dequeue()
	st.Equal(true, ok, "pop 2")
	st.Equal(2, v, "pop 2")
	st.Equal(true, q.IsEmpty(), "queue is empty")
	_, ok = q.Dequeue()
	st.False(ok, "queue is empty")
}

func (st *suiteMISOQueue) TestQueue_Empty() {
	q := NewMISOQueue[uint64]()
	st.True(q.IsEmpty())
	q.Enqueue(1)
	st.False(q.IsEmpty())
}

func (st *suiteMISOQueue) TestPoolQueue_PushPopOneProducer() {
	st.testQueue_PushPopOneProducer()
}

func (st *suiteMISOQueue) TestQueue_PushPopOneProducer() {
	st.testQueue_PushPopOneProducer()
}

func (st *suiteMISOQueue) testQueue_PushPopOneProducer() {
	expCount := 100

	var wg sync.WaitGroup
	wg.Add(1)
	q := NewMISOQueue[uint64]()
	go func() {
		i := 0
		for {
			v, ok := q.Dequeue()
			if !ok {
				runtime.Gosched()
				continue
			}
			st.Equal(uint64(i), v, "pop value")
			i++
			if i == expCount {
				wg.Done()
				return
			}
		}
	}()

	// var val interface{} = "foo"

	for i := 0; i < expCount; i++ {
		q.Enqueue(uint64(i))
	}

	wg.Wait()
}

func (st *suiteMISOQueue) TestPoolMpscQueueConsistency() {
	st.testMpscQueueConsistency()
}

func (st *suiteMISOQueue) TestNoPoolMpscQueueConsistency() {
	st.testMpscQueueConsistency()
}

func (st *suiteMISOQueue) testMpscQueueConsistency() {
	c := 100
	max := 1000000
	max = max / c * c
	var wg sync.WaitGroup
	wg.Add(1)
	q := NewMISOQueue[string]()
	go func() {
		i := 0
		seen := make(map[string]string)
		for {
			r, ok := q.Dequeue()
			if !ok {
				runtime.Gosched()

				continue
			}
			i++
			s := r
			_, present := seen[s]
			if present {
				st.FailNow("item have already been seen %v", s)
			}
			seen[s] = s
			if i == max {
				wg.Done()
				return
			}
		}
	}()

	for j := 0; j < c; j++ {
		jj := j
		cmax := max / c
		go func() {
			for i := 0; i < cmax; i++ {
				if rand.Intn(10) == 0 {
					runtime.Gosched()
					// time.Sleep(time.Duration(rand.Intn(1000)))
				}
				q.Enqueue(fmt.Sprintf("%v_%v", jj, i))
			}
		}()
	}

	wg.Wait()
	// time.Sleep(50 * time.Millisecond)
	// queue should be empty
	for i := 0; i < 100; i++ {
		r, ok := q.Dequeue()
		if ok {
			st.FailNow("unexpected result %+v", r)
		}
	}
}

func benchmarkPushPop(count, c int) {
	var wg sync.WaitGroup
	wg.Add(1)
	q := NewMISOQueue[string]()
	go func() {
		i := 0
		for {
			_, ok := q.Dequeue()
			if !ok {
				runtime.Gosched()
				continue
			}
			i++
			if i == count {
				wg.Done()
				return
			}
		}
	}()

	val := "foo"

	for i := 0; i < c; i++ {
		go func(n int) {
			for n > 0 {
				q.Enqueue(val)
				n--
			}
		}(count / c)
	}

	wg.Wait()
}

func BenchmarkPushPop(b *testing.B) {
	benchmarks := []struct {
		count       int
		concurrency int
	}{
		{
			count:       32000,
			concurrency: 1,
		},
		{
			count:       32000,
			concurrency: 4,
		},
		{
			count:       32000,
			concurrency: 8,
		},
		{
			count:       32000,
			concurrency: 16,
		},
		{
			count:       32000,
			concurrency: 32,
		},
	}
	for _, bm := range benchmarks {
		b.Run(fmt.Sprintf("%d_%d", bm.count, bm.concurrency), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchmarkPushPop(bm.count, bm.concurrency)
			}
		})
	}
}
