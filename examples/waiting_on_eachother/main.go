package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/meha555/detectlock-go"
)

var Locker1 sync.Locker
var Locker2 sync.Locker

func init() {
	Locker1 = &detectlock.Mutex{}
	Locker2 = &detectlock.Mutex{}
}

func A() {
	defer Locker1.Unlock()
	Locker1.Lock()
	time.Sleep(time.Millisecond * 300)
	defer Locker2.Unlock()
	Locker2.Lock()

	time.Sleep(time.Millisecond * 300)
}

func B() {
	defer Locker2.Unlock()
	Locker2.Lock()
	time.Sleep(time.Millisecond * 400)
	defer Locker1.Unlock()
	Locker1.Lock()

	time.Sleep(time.Millisecond * 400)
}

func main() {
	detectlock.Enable()
	for i := 0; i < 10; i++ {
		go A()
		go B()
	}

	time.Sleep(time.Second * 2)

	records := detectlock.Records()

	fmt.Println("--- DetectAcquired ---")
	fmt.Println(detectlock.DetectAcquired(records))

	fmt.Println("--- DetectWaitingOnEachOther ---")
	fmt.Println(detectlock.DetectWaitingOnEachOther(records))

	// fmt.Println("--- all stack ---")
	// b := make([]byte, 102400)
	// b = b[:runtime.Stack(b, true)]
	//fmt.Println(string(b))
	// fmt.Println("------- end --------")

	detectlock.Disable()
}
