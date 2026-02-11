package detectlock

import "log"

// acquire lock
// lockerPtr: uintptr of locker
// rLocker: is read lock?
// doLock: implementation of lock operation
func acquire(lockerPtr uintptr, rLocker bool, doLock func()) {
	if doLock == nil {
		panic("doLock is nil")
	}
	log.Printf("trace> lock %#x acquiring", lockerPtr)

	// locate a shard
	gid := getGoroutineID()
	shardKey := gid % shardCount
	bucket := buckets[shardKey]
	var locker *LockerState

	func() {
		defer bucket.locker.Unlock()
		bucket.locker.Lock()
		var lockers []*LockerState
		if exists, ok := bucket.items[gid]; ok {
			lockers = exists
		} else {
			lockers = make([]*LockerState, 0, 16)
		}
		locker = &LockerState{LockerPtr: lockerPtr, Status: StatusWaitting, RLocker: rLocker}
		lockers = append(lockers, locker)
		bucket.items[gid] = lockers
	}()

	locker.Caller = getCaller(4)

	doLock() // 无条件执行
	// 修改上锁状态必须在doLock之后
	locker.Status = StatusAcquired
	log.Printf("trace> lock %#x acquired ✔", lockerPtr)
	// log.Println(locker.String())
}

// try acquire lock
// lockerPtr: uintptr of locker
// rLocker: is read lock?
// tryLock: implementation of try lock operation
func tryAcquire(lockerPtr uintptr, rLocker bool, tryLock func() bool) bool {
	if tryLock == nil {
		panic("tryLock is nil")
	}
	log.Printf("trace> lock %#x try-acquiring", lockerPtr)

	// 无条件执行
	if tryLock() {
		gid := getGoroutineID()
		shardKey := gid % shardCount
		bucket := buckets[shardKey]
		var locker *LockerState

		func() {
			defer bucket.locker.Unlock()
			bucket.locker.Lock()
			var lockers []*LockerState
			if exists, ok := bucket.items[gid]; ok {
				lockers = exists
			} else {
				lockers = make([]*LockerState, 0, 16)
			}
			locker = &LockerState{LockerPtr: lockerPtr, Status: StatusAcquired, RLocker: rLocker}
			lockers = append(lockers, locker)
			bucket.items[gid] = lockers
		}()

		locker.Caller = getCaller(4)
		return true
	} else {
		return false
	}
}

// release lock
// lockerPtr: uintptr of locker
// rLocker: is read lock?
// doUnlock: implementation of unlock operation
func release(lockerPtr uintptr, rLocker bool, doUnlock func()) {
	if doUnlock == nil {
		panic("doUnlock is nil")
	}
	log.Printf("trace> lock %#x releasing", lockerPtr)

	// 无条件执行
	doUnlock()

	gid := getGoroutineID()
	shardKey := gid % shardCount
	bucket := buckets[shardKey]

	defer bucket.locker.Unlock()
	bucket.locker.Lock()
	if lockers, ok := bucket.items[gid]; ok {
		removeIndex := -1
		llen := len(lockers)
		for i := 0; i < llen; i++ {
			l := lockers[i]
			// 这里要确保是正确的解锁逻辑，这样一旦实际程序使用的方式错误，发生了死锁之类的问题，这里就会没有真正完成解锁，于是就能发现问题
			if l.LockerPtr == lockerPtr && l.Status == StatusAcquired && l.RLocker == rLocker {
				removeIndex = i
				break
			}
		}
		if removeIndex < 0 {
			return
		}
		// 从切片中移除remoteIndex对应的locker
		if llen == 1 {
			lockers = nil
		} else {
			lockers = append(lockers[:removeIndex], lockers[removeIndex+1:]...)
		}
		if len(lockers) == 0 {
			delete(bucket.items, gid)
		} else {
			bucket.items[gid] = lockers
		}
	}
}
