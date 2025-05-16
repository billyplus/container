package timewheel

import (
	"fmt"
	"iter"
	"sync/atomic"

	"github.com/billyplus/container"
)

const (
	// MaxMilliSecondDelay 毫秒级的最大延迟
	MaxMilliSecondDelay = 24 * 3600 * 1000
	Tick16              = 16
	miliInterval        = 5
	miliCount           = 1000 / miliInterval
)

// HH HH HH MM SS MS MS
// ff ff ff 3c 3c 03 E8

type wheelTaskFunc func(ms int64)

type wheelTask struct {
	fn wheelTaskFunc
	ms int32
	ss int32
	mm int32
}

type queueTask struct {
	end int64
	fn  wheelTaskFunc
	// next atomic.Pointer[queueTask]
}

type wheelRing struct {
	static    []int
	expand    []int
	staticCap int
}

// type wheelHandValue int64

// func (h wheelHandValue) mm() int32 {
// 	return int32(h >> 24 & 0xff)
// }

// func (h wheelHandValue) ss() int32 {
// 	return int32(h >> 16 & 0xff)
// }

// func (h wheelHandValue) ms() int32 {
// 	return int32(h >> 8 & 0xff)
// }

// func (h wheelHandValue) dur() int32 {
// 	return int32(h & 0xff)
// }

// func (h *wheelHandValue) setValue(mm, ss, ms, dur int32) {
// 	*h = wheelHandValue((int64(mm) << 24) | (int64(ss) << 16) | (int64(ms) << 8) | (int64(dur)))
// }

// func (h wheelHandValue) String() string {
// 	return fmt.Sprintf("%d:%d:%d:%d", h.mm(), h.ss(), h.ms(), h.dur())
// }

func (wr *wheelRing) add(idx int) {
	if len(wr.static) < wr.staticCap {
		wr.static = append(wr.static, idx)
	} else {
		wr.expand = append(wr.expand, idx)
	}
}

func (wr *wheelRing) reset() {
	wr.static = wr.static[:0]
	wr.expand = nil
}

func (wr *wheelRing) len() int {
	if wr == nil {
		return 0
	}
	return len(wr.static) + len(wr.expand)
}

func (wr *wheelRing) all() iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, idx := range wr.static {
			if !yield(idx) {
				return
			}
		}
		for _, idx := range wr.expand {
			if !yield(idx) {
				return
			}
		}
	}
}

type TimeWheelMilliSecond struct {
	ringPool  container.Pool[*wheelRing]
	hand      int32
	realHand  int64
	taskCount int32
	tasks     []wheelTask
	available container.Stack[int]
	ring      []*wheelRing
	toRun     []wheelTaskFunc
	queue     *container.MISOQueue[*queueTask]
	opt       TimeWheelOption
}

type TimeWheelOption struct {
	Logger container.IErrorLogger
}

func NewTimeWheelMilliSecond(opt ...TimeWheelOption) *TimeWheelMilliSecond {
	wh := &TimeWheelMilliSecond{}
	if len(opt) > 0 {
		wh.opt = opt[0]
	}
	wh.tasks = make([]wheelTask, 512)
	wh.available = make(container.Stack[int], 0, 512)
	for i := range 512 {
		wh.available.Push(i)
	}
	wh.ring = make([]*wheelRing, 30+60+miliCount)
	wh.ringPool = container.Pool[*wheelRing](container.NewPool(func() *wheelRing {
		return &wheelRing{
			staticCap: 32,
			static:    make([]int, 0, 32),
		}
	}))
	wh.toRun = make([]wheelTaskFunc, 0, 256)
	wh.queue = container.NewMISOQueue[*queueTask]()

	// now := time.Now()
	// wh.lastTick = now.UnixMilli()
	// mili := wh.lastTick % 1000
	// mili = mili / 20
	// wh.hand = mili<<8 | wh.lastTick%20
	// fmt.Println("init hand", wh.hand>>8&0xff, wh.hand&0xff, wh.lastTick)
	return wh
}

func (wh *TimeWheelMilliSecond) Insert(tickTime int64, task wheelTaskFunc) {
	// if delay > MaxMilliSecondDelay {
	// 	panic("超出最大延迟范围")
	// }

	// ms := time.Now().UnixMilli()

	// hand := atomic.LoadInt64(&wh.hand)

	// fmt.Println("insert", tickTime, hand>>32, hand>>24&0xff, hand>>16&0xff, hand>>8&0xff, hand&0xff)

	atomic.AddInt32(&wh.taskCount, 1)

	wh.queue.Enqueue(&queueTask{
		end: tickTime,
		fn:  task,
	})
}

func (wh *TimeWheelMilliSecond) addToWheel(tidx int) {
	task := &wh.tasks[tidx]
	hand := wh.hand
	mm := hand >> 24 & 0xff
	ss := hand >> 16 & 0xff
	// ms := hand >> 8 & 0xff
	idx := task.ms
	if task.mm != int32(mm) {
		idx = task.mm + miliCount + 60
		// wh.minuteList[task.minute] = append(wh.minuteList[task.minute], task)
	} else if task.ss != int32(ss) {
		idx = task.ss + miliCount
		// wh.secondList[task.second] = append(wh.secondList[task.second], task)
		// } else {
		// wh.msList[task.ms] = append(wh.msList[task.ms], task)
	}
	fmt.Println("add to wheel", idx)
	if wh.ring[idx] == nil {
		wh.ring[idx] = wh.ringPool.Get()
	}

	wh.ring[idx].add(tidx)
}

func (wh *TimeWheelMilliSecond) expand() {
	l := len(wh.tasks)
	lst := make([]wheelTask, 2*l)
	copy(lst, wh.tasks)
	lst = lst[:l]
	wh.tasks = lst
	wh.available = container.NewStackWithCap[int](2 * l)
	for i := l; i < 2*l; i++ {
		wh.available.Push(i)
	}
}

// Update move time tick forward. and handle task in current ms list
func (wh *TimeWheelMilliSecond) Update(ms int64) {
	if wh == nil {
		return
	}
	fmt.Println("update ms=", ms, "hand=", wh.realHand)
	if wh.realHand == 0 {
		wh.realHand = ms
		mili := wh.realHand % 1000 / miliInterval
		wh.hand = int32(mili<<8 | wh.realHand%miliInterval)
		// wh.lastTick = ms - ms%miliInterval + miliInterval
		fmt.Println("init hand", wh.hand>>8&0xff, wh.hand&0xff)
		return
	}
	// if ms-wh.lastTick <= miliInterval {
	// 	// 16ms一个tick
	// 	return
	// }
	// 处理队列
	wh.dequeue(ms)
	// 更新时间
	wh.updateHand(ms)

	// if ms-wh.lastRun16 >= Tick16 {
	// 执行任务
	wh.runTask16(ms)

	// fmt.Println("update lastRun16", ms, wh.re, ms%miliInterval)
	// 	wh.lastRun16 = ms - ms%miliInterval + miliInterval
	// }
}

func (wh *TimeWheelMilliSecond) dequeue(ms int64) {
	for {
		task, ok := wh.queue.Dequeue()
		if !ok {
			// fmt.Println("dequeue empty")
			break
		}
		fmt.Println("dequeue", task.end-ms)
		if task.end <= ms {
			// 到时了，马上运行
			wh.toRun = append(wh.toRun, task.fn)
			continue
		}

		if wh.available.Len() == 0 {
			wh.expand()
		}

		hand := wh.hand

		mm := hand >> 24 & 0xff
		ss := hand >> 16 & 0xff
		mili := hand >> 8 & 0xff
		delay := int32(task.end-wh.realHand) + hand&0xff
		fmt.Println("before add task", wh.realHand, mm, ss, mili, delay)

		tt := wheelTask{
			fn: task.fn,
			ms: int32(mili),
			ss: int32(ss),
			mm: int32(mm),
		}
		// if delay < 20 {
		// 	delay += 20
		// }

		tt.ms += (delay % 1000) / miliInterval // 20ms per tick

		delay = delay / 1000 // 秒数
		if delay >= 1 {
			// atomic.AddInt32(&task.second, (delay % 60))
			tt.ss += (delay % 60)
		}
		delay = delay / 60 // 分钟
		if delay >= 1 {
			tt.mm += (delay % 60)
		}

		if tt.ms >= miliCount {
			tt.ms -= miliCount
			tt.ss++
		}
		if tt.ss >= 60 {
			tt.ss -= 60
			tt.mm++
		}
		if tt.mm >= 30 {
			tt.mm -= 30
		}

		tidx, ok := wh.available.Pop()
		if !ok {
			wh.expand()
			tidx, _ = wh.available.Pop()
		}
		wh.tasks[tidx] = tt
		fmt.Println("add task", tt.mm, tt.ss, tt.ms)

		wh.addToWheel(tidx)
	}
}

func (wh *TimeWheelMilliSecond) updateHand(ms int64) {
	// hand := atomic.LoadInt64(&wh.hand)
	hand := wh.hand
	mm := hand >> 24 & 0xff
	ss := hand >> 16 & 0xff
	mili := hand >> 8 & 0xff
	dur := hand & 0xff

	dur += int32(ms - wh.realHand)
	wh.realHand = ms
	for dur >= miliInterval {
		fmt.Println("updateTimer", ms, mm, ss, mili, dur)
		// run ms list task
		list := wh.ring[mili]
		if list.len() > 0 {
			wh.hand = (mm << 24) | (ss << 16) | (mili << 8) | dur
			fmt.Println("to run ms task", ms, mm, ss, mili, dur)
			// log.Debug().Int64("now", ms).Int("sec", wheel.second).Int("ms", wheel.ms).Msg("run ms task")
			for _, task := range list.static {
				tt := wh.tasks[task]
				wh.toRun = append(wh.toRun, tt.fn)
				wh.available.Push(task)
			}
			for _, task := range list.expand {
				tt := wh.tasks[task]
				wh.toRun = append(wh.toRun, tt.fn)
				wh.available.Push(task)
			}
			// clear list
			list.reset()
			wh.ring[mili] = nil
			wh.ringPool.Put(list)
		}
		// move forward ms hand
		dur -= miliInterval
		mili++
		fmt.Println("move mili", mili, dur)

		if mili >= miliCount {
			// move forward second hand
			mili -= miliCount
			ss++
			if ss >= 60 {
				// move forward minute hand
				ss -= 60
				mm++
				if mm >= 30 {
					// move forward hour hand
					mm -= 30
				}
				// deal with minute list
				list = wh.ring[mm+miliCount+60]
				if list.len() > 0 {
					wh.hand = (mm << 24) | (ss << 16) | (mili << 8) | dur
					for taskid := range list.all() {
						wh.addToWheel(taskid)
					}

					// clear list
					list.reset()
					wh.ring[mm+miliCount+60] = nil
					wh.ringPool.Put(list)
				}
			}
			// deal with second list
			list = wh.ring[ss+miliCount]
			// log.Debug().Int64("now", ms).Int("sec", wheel.second).Int("ms", wheel.ms).Msg("run ms task")
			if list.len() > 0 {
				wh.hand = (mm << 24) | (ss << 16) | (mili << 8) | dur
				fmt.Println("to run ss task", ss)
				for taskid := range list.all() {
					wh.addToWheel(taskid)
				}
				// clear list
				list.reset()
				wh.ring[ss+miliCount] = nil
				wh.ringPool.Put(list)
			}

		}

	}
	// update hand
	wh.hand = (mm << 24) | (ss << 16) | (mili << 8) | dur
	fmt.Println("update hand", ms, wh.hand, mm, ss, mili, dur)

	// atomic.StoreInt64(&wh.hand, (hh<<32)|(mm<<24)|(ss<<16)|(mili<<8)|dur)
}

func (wh *TimeWheelMilliSecond) runTask16(ms int64) {
	if len(wh.toRun) > 0 {
		hand := wh.hand
		mm := hand >> 24 & 0xff
		ss := hand >> 16 & 0xff
		mili := hand >> 8 & 0xff
		dur := hand & 0xff
		fmt.Println("runTask", ms, mm, ss, mili, dur)
		for _, task := range wh.toRun {
			wh.runOneTask(ms, task)
			atomic.AddInt32(&wh.taskCount, -1)
		}
		wh.toRun = wh.toRun[:0]
	}
}

func (wh *TimeWheelMilliSecond) runOneTask(ms int64, fn func(int64)) {
	defer func() {
		// 收集错误
		if e := recover(); e != nil {
			err, ok := e.(error)
			if !ok {
				err = fmt.Errorf("msg from panic: %v", e)
			}
			if wh.opt.Logger != nil {
				wh.opt.Logger.Error(err, "recover from wheel task")
			}
		}
	}()
	fn(ms)
}
