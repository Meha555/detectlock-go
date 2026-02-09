package detectlock

import (
	"sync"
	"unsafe"
)

// Mutex wrapped from sync.Mutex
type Mutex struct {
	m sync.Mutex
}

func (l *Mutex) Lock() {
	if debug {
		// 这里只是想要获取l的内存首地址，因此直接强制转换即可，没必要用反射
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		acquire(lockerPtr, false, l.m.Lock)
	} else {
		l.m.Lock()
	}
}

func (l *Mutex) TryLock() bool {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		return tryAcquire(lockerPtr, false, l.m.TryLock)
	} else {
		return l.m.TryLock()
	}
}

func (l *Mutex) Unlock() {
	if debug {
		// lockerPtr := reflect.ValueOf(l).Pointer()
		lockerPtr := uintptr(unsafe.Pointer(l))
		release(lockerPtr, false, l.m.Unlock)
	} else {
		l.m.Unlock()
	}
}
