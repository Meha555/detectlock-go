package detectlock

import (
	"sync"
	"unsafe"
)

// Mutex wrapped from sync.RWMutex
type RWMutex struct {
	rwm sync.RWMutex
}

func (l *RWMutex) RLock() {
	if debug {
		// 这里只是想要获取l的内存首地址，因此直接强制转换即可，没必要用反射
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		acquire(lockerPtr, true, l.rwm.RLock)
	} else {
		l.rwm.RLock()
	}
}

func (l *RWMutex) TryRLock() bool {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		return tryAcquire(lockerPtr, true, l.rwm.TryRLock)
	} else {
		return l.rwm.TryRLock()
	}
}

func (l *RWMutex) RUnlock() {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		release(lockerPtr, true, l.rwm.RUnlock)
	} else {
		l.rwm.RUnlock()
	}
}

func (l *RWMutex) Lock() {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		acquire(lockerPtr, false, l.rwm.Lock)
	} else {
		l.rwm.Lock()
	}
}

func (l *RWMutex) TryLock() bool {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		return tryAcquire(lockerPtr, false, l.rwm.TryLock)
	} else {
		return l.rwm.TryLock()
	}
}

func (l *RWMutex) Unlock() {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		release(lockerPtr, false, l.rwm.Unlock)
	} else {
		l.rwm.Unlock()
	}
}
