package container

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestFixedRingBuff(t *testing.T) {
	suite.Run(t, new(suiteRingBuffFixed))
}

type suiteRingBuffFixed struct {
	suite.Suite
}

func (st *suiteRingBuffFixed) TestPushPop() {
	q := NewRingBuffFixed[string](10)
	last, ok := q.Last()
	st.False(ok, "last not exist")
	st.Equal("", last, "check last")
	q.Push("hello")
	last, ok = q.Last()
	st.True(ok, "last exist")
	st.Equal("hello", last, "check last")
	q.Push("hello2")
	last, ok = q.Last()
	st.True(ok, "last exist")
	st.Equal("hello2", last, "check last")
	q.Push("hello3")
	last, ok = q.Last()
	st.True(ok, "last exist")
	st.Equal("hello3", last, "check last")
	st.Equal(3, q.Len())
	all := q.GetAll()
	st.Equal([]string{"hello", "hello2", "hello3"}, all, "")
}

func (st *suiteRingBuffFixed) TestPushPopMany() {
	q := NewRingBuffFixed[int](10)
	for i := 0; i < 22; i++ {
		q.Push(i)
	}
	st.Equal(10, q.Len())
	all := q.GetAll()
	st.Equal([]int{12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, all, "")
}
