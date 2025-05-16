package container

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestRingBuff(t *testing.T) {
	suite.Run(t, new(suiteRingBuffMIMO))
}

type suiteRingBuffMIMO struct {
	suite.Suite
}

func (st *suiteRingBuffMIMO) TestPushPop() {
	q := NewRingBuffMIMO[string](10)
	q.Push("hello")
	res, ok := q.Pop()
	st.True(ok, "has value")
	st.Equal("hello", res)
	st.True(q.Empty())
}

func (st *suiteRingBuffMIMO) TestPushPopRepeated() {
	q := NewRingBuffMIMO[string](10)
	for i := 0; i < 100; i++ {
		q.Push("hello")
		res, ok := q.Pop()
		st.True(ok, "has value")
		st.Equal("hello", res)
		st.True(q.Empty())
	}
}

func (st *suiteRingBuffMIMO) TestPushPopMany() {
	q := NewRingBuffMIMO[string](10)
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("hello%v", i)
		q.Push(item)
		res, ok := q.Pop()
		st.True(ok, "has value")
		st.Equal(item, res)
	}
	st.True(q.Empty())
}

func (st *suiteRingBuffMIMO) TestPushPopMany2() {
	q := NewRingBuffMIMO[string](10)
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("hello%v", i)
		q.Push(item)
	}
	for i := 0; i < 10000; i++ {
		item := fmt.Sprintf("hello%v", i)
		res, ok := q.Pop()
		st.True(ok, "has value")
		st.Equal(item, res)
	}
	st.True(q.Empty())
}

func (st *suiteRingBuffMIMO) TestLfQueueConsistency() {
	max := 1000000
	c := 100
	var wg sync.WaitGroup
	wg.Add(1)
	q := NewRingBuffMIMO[string](2)
	go func() {
		i := 0
		seen := make(map[string]string)
		for {
			s, ok := q.Pop()
			if !ok {
				runtime.Gosched()

				continue
			}
			i++
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
				}
				q.Push(fmt.Sprintf("%v %v", jj, i))
			}
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)
	// queue should be empty
	for i := 0; i < 100; i++ {
		r, ok := q.Pop()
		if ok {
			st.FailNow("unexpected result %+v", r)
		}
	}
}
