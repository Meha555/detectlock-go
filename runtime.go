package detectlock

import (
	"bytes"
	"io"
	"runtime"
	"strconv"
	"strings"
	_ "unsafe"
)

var getGoroutineID func() int64

func init() {
	SetGetGoroutineIDFunc(func() int64 {
		buf := make([]byte, 64)
		buf = buf[0:runtime.Stack(buf, false)]
		index := bytes.Index(buf, []byte{'['})
		buf = buf[0:index]
		buf = bytes.TrimLeft(buf, "goroutine")
		buf = bytes.TrimSpace(buf)
		gid, _ := strconv.ParseInt(string(buf), 10, 64)
		return gid
	})
}

//go:linkname SetDebugOutput log.SetOutput
func SetDebugOutput(w io.Writer)

// SetGetGoroutineIDFunc to set how to get goroutine id.
// e.g.: SetGetGoroutineIDFunc(goid.Get) from github.com/petermattis/goid
func SetGetGoroutineIDFunc(f func() int64) {
	if f != nil {
		getGoroutineID = f
	}
}

func getCaller(skip int) *runtime.Frame {
	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, 10)
	depth := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, more := frames.Next(); more; f, more = frames.Next() {
		// If the caller isn't part of this package, we're done
		if !strings.Contains(f.Function, "github.com/meha555/detectlock-go.") {
			return &f //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}
