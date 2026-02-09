package detectlock

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// 分桶数量
const shardCount int64 = 1 << 16

// Status of locker.
const (
	StatusWaitting byte = iota // wait or r-wait
	StatusAcquired             // acquired or r-acquired
)

// 链地址法分片哈希的locker state桶，分片依据和键是goroutine id。
var buckets cacheBuckets

// LockerState the state of locker.
type LockerState struct {
	LockerPtr uintptr
	Status    byte
	RLocker   bool // is read lock or not
	Caller    *runtime.Frame
}

// String of LockerState, format: (<locker-id>, <locker-status>)
func (l LockerState) String() string {
	status := "waitting"
	if l.Status == StatusAcquired {
		status = "acquired"
	}
	if l.RLocker {
		status = "r-" + status
	}

	stackInfo := ""
	if l.Caller != nil {
		stackInfo = fmt.Sprintf("%s (file: %s:%d)", l.Caller.Function, l.Caller.File, l.Caller.Line)
	}
	return fmt.Sprintf("(%#x, %s, %s)", l.LockerPtr, status, stackInfo)
}

// LockerStateList the list of LockerState
type LockerStateList []LockerState

func (l LockerStateList) Len() int {
	return len(l)
}

func (l LockerStateList) Less(i, j int) bool {
	return l[i].Status > l[j].Status
}

func (l LockerStateList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// String of LockerStateList, format: [(<locker-id>, <locker-status>), ...]
func (l LockerStateList) String() string {
	if len(l) == 0 {
		return "[]"
	}
	sb := &strings.Builder{}
	sb.WriteString("[")
	llen := len(l)
	for i, locker := range l {
		sb.WriteString(locker.String())
		if i < llen-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// 获取参与锁竞争的goroutine的LockerState
// 非并发安全
func Records() map[int64]LockerStateList {
	items := make(map[int64]LockerStateList)
	for _, bucket := range buckets {
		func() {
			defer bucket.locker.Unlock()
			bucket.locker.Lock()
			for k, v := range bucket.items {
				lockers := make(LockerStateList, len(v))
				for i := 0; i < len(v); i++ {
					lockers[i] = *v[i] // 值拷贝，因此这里不会受其他协程修改锁状态的影响（只需要确保Items调用时进行加锁）
				}
				items[k] = lockers
			}
		}()
	}
	return items
}

type cacheBuckets []*cacheBucket

type cacheBucket struct {
	locker sync.RWMutex
	items  map[int64][]*LockerState
}

func clear() {
	buckets = make(cacheBuckets, shardCount)
}

func reset() {
	buckets = make(cacheBuckets, shardCount)
	var i int64 = 0
	for ; i < shardCount; i++ {
		buckets[i] = &cacheBucket{items: make(map[int64][]*LockerState)}
	}
}
