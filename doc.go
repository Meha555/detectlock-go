// Package detectlock for detect dead locker
//
// 1. replace mutex locker:
//
// 1) replace "var locker *sync.Mutex = &sync.Mutex{}" to "var locker *detectlock.Mutex = &detectlock.Mutex{}"
//
// 2) replace "var locker *sync.RWMutex = &sync.RWMutex{}" to "var locker *detectlock.RWMutex = &detectlock.RWMutex{}"
//
// 2. enable detection on startup
//
// detectlock.Enable()" or disable it by "detectlock.Disable()"
//
// 3. detect dead locker
//
// records := detectlock.Records()
//
// detect reentry locker: "detectlock.DetectReentry(records)"
//
// detect locked each other: "detectlock.DetectWaitingOnEachOther(records)"
//
// detect acquired owners: "detectlock.DetectAcquired(records)"
package detectlock
