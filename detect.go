package detectlock

import (
	"fmt"
	"sort"
	"strings"
)

var debug bool

// Enable enable lock detection
func Enable() {
	reset()
	debug = true
}

// Disable disable lock detection
func Disable() {
	debug = false
	clear()
}

// DetectAcquired detect goroutine that locker acquired.
// 返回已获得了锁的goroutine列表
func DetectAcquired(items map[int64]LockerStateList) (l GoroutineLockerList) {
	for gid, lockers := range items {
		lockerAcquired := false
		for _, locker := range lockers {
			if locker.Status == StatusAcquired {
				lockerAcquired = true
				break
			}
		}
		if lockerAcquired {
			l = append(l, GoroutineLocker{GoroutineID: gid, Lockers: lockers})
		}
	}
	sort.Sort(l)
	return l
}

// DetectWaitingOnEachOther detect goroutine that locked each other.
// 核心思想：如果一个 goroutine 既持有锁又在等待锁，就可能存在死锁风险（仅是死锁的必要不充分条件）。
// 检测 goroutine 之间是否存在相互等待锁的情况，返回有相互等待锁的 goroutine 列表
// 处理普通锁死锁和读写锁死锁。其中普通锁可以视作读写锁的写锁。
// 普通锁可能死锁的情况：一个协程已经获得了一把锁，但还要等待另一把锁
// 读写锁可能死锁的情况：
// - 情形1：普通锁(写锁)死锁
// - 情形2：读锁和写锁间死锁：一个协程已持有读锁，另一个协程已持有写锁，且两个协程均还在等待另一把锁
// REVIEW: 这个函数没有理解
func DetectWaitingOnEachOther(items map[int64]LockerStateList) (l GoroutineLockerList) {
	for gid, lockers := range items {
		lockerAcquired := false               // 当前协程是否已获得非读锁（即普通锁或写锁）
		acquiredLockers := make([]uintptr, 0) // 当前协程已获得的锁列表
		waiting := false                      // 是否有等待中的锁
		for _, locker := range lockers {
			if locker.Status == StatusAcquired {
				acquiredLockers = append(acquiredLockers, locker.LockerPtr)
				if !locker.RLocker { // 写锁获取了就是真的获取了
					lockerAcquired = true
				} else { // 读锁只是对于第一个获取该锁的协程才是真正获取了，因此需要检查其他协程
				loop:
					for ogid, olockers := range items {
						// 跳过自己
						if ogid == gid {
							continue
						}
						lockerWaitByOtherGoroutine := false     // 是否出现了一个协程加了读锁，另一个协程等待写锁的情况
						lockerAcquiredByOtherGoroutine := false // 是否出现了两个协程分别加了锁的情况（读锁或普通锁）
						for _, olocker := range olockers {
							// 检查是否有其他协程使用写锁来等待这把读写锁
							if olocker.LockerPtr == locker.LockerPtr && olocker.Status == StatusWaitting && !olocker.RLocker {
								lockerWaitByOtherGoroutine = true
							}
							// 检查是否其他协程也已经持有了某些锁
							if olocker.Status == StatusAcquired { // FIXME: 这里是不是不太对啊，没有考虑olocker也是读写锁的情况，这里似乎直接就认为olocker就是普通锁了？
								lockerAcquiredByOtherGoroutine = true
							}
							// 满足持有该读写锁的写锁+还持有其他锁，则可能满足循环等待的必要不充分条件
							if lockerWaitByOtherGoroutine && lockerAcquiredByOtherGoroutine {
								lockerAcquired = true
								break loop
							}
						}
					}
				}
			} else {
				// 如果locker.LockerPtr不在acquiredLockers中则置waiting为true
				existsInAcquiredLockers := false
				for _, acuqiredLocker := range acquiredLockers {
					if acuqiredLocker == locker.LockerPtr {
						existsInAcquiredLockers = true
						break
					}
				}
				if !existsInAcquiredLockers {
					waiting = true
				}
			}
		}
		// 已获取后还在等待，说明有可能发生循环依赖。加入返回列表
		if lockerAcquired && waiting {
			l = append(l, GoroutineLocker{GoroutineID: gid, Lockers: lockers})
		}
	}
	sort.Sort(l)
	return l
}

// DetectReentry detect goroutine that reentry locker occurred.
func DetectReentry(items map[int64]LockerStateList) (l GoroutineLockerList) {
	for gid, lockers := range items {
		acquiredLockers := make([]uintptr, 0)
		for _, locker := range lockers {
			if locker.Status == StatusAcquired {
				acquiredLockers = append(acquiredLockers, locker.LockerPtr)
			} else {
				existsInAcquiredLockers := false
				for _, acuqiredLocker := range acquiredLockers {
					if acuqiredLocker == locker.LockerPtr {
						existsInAcquiredLockers = true
						break
					}
				}
				// 说明是尝试等待一个已经自己获得的锁
				if existsInAcquiredLockers {
					l = append(l, GoroutineLocker{GoroutineID: gid, Lockers: lockers})
				}
			}
		}
	}
	sort.Sort(l)
	return l
}

// GoroutineLocker goroutine with lockers.
type GoroutineLocker struct {
	GoroutineID int64
	Lockers     LockerStateList
}

// String of GoroutineLocker, format: goroutine <gid>: [(<locker-id>, <locker-status>), ...]\n, like:
//
// goroutine 53: [(0xc000014080, acquired), (0xc000014088, wait)]
func (l GoroutineLocker) String() string {
	return fmt.Sprintf("goroutine %d: %s\n", l.GoroutineID, l.Lockers)
}

// GoroutineLockerList the list of GoroutineLocker
type GoroutineLockerList []GoroutineLocker

func (l GoroutineLockerList) Len() int {
	return len(l)
}

func (l GoroutineLockerList) Less(i, j int) bool {
	return l[i].GoroutineID < l[j].GoroutineID
}

func (l GoroutineLockerList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// String of GoroutineLockerList, format: goroutine <gid>: [(<locker-id>, <locker-status>), ...]\n..., like:
//
// goroutine 53: [(0xc000014080, acquired), (0xc000014088, wait)]
//
// goroutine 54: [(0xc000014088, acquired), (0xc000014080, wait)]
func (l GoroutineLockerList) String() string {
	if len(l) == 0 {
		return ""
	}
	sb := &strings.Builder{}
	for _, glocker := range l {
		sb.WriteString(glocker.String())
	}
	return sb.String()
}
