package timewheel

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeWheelMs(t *testing.T) {
	mslist := []int64{172223430, 172223436, 172223440, 172223441, 172223439, 1747312310251, time.Now().UnixMilli()}
	// mslist := []int64{1747321355416}
	// wheel.hand = mili
	// wheel.Update(nowMs)
	// nowMs += 32 * 60 * 1000
	// wheel.Update(nowMs)
	delaymap := map[int64]int{
		3:             2,
		29:            2,
		49:            2,
		50:            3,
		51:            2,
		59:            3,
		200:           2,
		300:           1,
		1100:          1,
		3223:          64,
		4210:          2,
		59010:         2,
		5 * 60 * 1000: 2,
	}

	for _, ms := range mslist {
		result := make(map[int64]int)
		wheel := NewTimeWheelMilliSecond()
		maxDelay := int64(0)
		for delay, count := range delaymap {
			if maxDelay < delay {
				maxDelay = delay
			}
			for range count {
				addTestCase(wheel, t, ms, delay, &result)
			}
		}
		nowMs := ms
		wheel.Update(nowMs)
		for range maxDelay/3 + 5 {
			nowMs += 3
			wheel.Update(nowMs)
		}
		for delay, count := range delaymap {
			assert.Equalf(t, count, result[delay], "ms %d delay %d should excute %d times", ms, delay, count)
		}
	}
}

func checkHand(hand int64, hh, mm, ss, ms int32) bool {
	fmt.Println("actual ", int32(hand>>32), int32(hand>>24&0xff), int32(hand>>16&0xff), int32(hand>>8&0xff))
	fmt.Println("expect ", hh, mm, ss, ms)
	return int32(hand>>32) == hh && int32(hand>>24&0xff) == mm && int32(hand>>16&0xff) == ss && int32(hand>>8&0xff) == ms
}

func addTestCase(wheel *TimeWheelMilliSecond, t *testing.T, startTime, delay int64, result *map[int64]int) {
	tickTime := startTime + delay
	wheel.Insert(tickTime, func(ms int64) {
		fmt.Println("trigger", ms, delay)
		assert.Lessf(t, ms-tickTime, int64(Tick16), "[%d+%d]间隔小于%dms", startTime, delay, Tick16)
		assert.GreaterOrEqualf(t, ms-tickTime, int64(0), "[%d+%d]间隔大于0", startTime, delay)
		(*result)[delay] = (*result)[delay] + 1
	})
}

func TestTimeWheelMsExpand(t *testing.T) {
	mslist := []int64{time.Now().UnixMilli()}
	// mslist := []int64{1747321355416}
	// wheel.hand = mili
	// wheel.Update(nowMs)
	// nowMs += 32 * 60 * 1000
	// wheel.Update(nowMs)
	delaymap := map[int64]int{
		200: 1030,
	}

	for _, ms := range mslist {
		result := make(map[int64]int)
		wheel := NewTimeWheelMilliSecond()
		maxDelay := int64(0)
		for delay, count := range delaymap {
			if maxDelay < delay {
				maxDelay = delay
			}
			for range count {
				addTestCase(wheel, t, ms, delay, &result)
			}
		}
		nowMs := ms
		wheel.Update(nowMs)
		for range maxDelay/3 + 5 {
			nowMs += 3
			wheel.Update(nowMs)
		}
		for delay, count := range delaymap {
			assert.Equalf(t, count, result[delay], "ms %d delay %d should excute %d times", ms, delay, count)
		}
	}
}
