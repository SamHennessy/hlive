package hlive

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

const LevelTrace = slog.Level(-8)

// Logger is a global logger used when a logger is not available
var Logger *slog.Logger

// LoggerDev is a global logger needed for developer warnings to avoid the need for panics
var LoggerDev *slog.Logger

func init() {
	Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	LoggerDev = slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func callerFrame(skip int) (runtime.Frame, bool) {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(skip+2, rpc[:])
	if n < 1 {
		return runtime.Frame{}, false
	}
	frame, _ := runtime.CallersFrames(rpc).Next()

	return frame, frame.PC != 0
}

func CallerStackStr() string {
	skip := 0
	stack := ""
	prefix := ""

startFrame:

	skip++
	frame, ok := callerFrame(skip)
	if ok {
		stack = fmt.Sprintf("%s:%v%s%s", frame.Function, frame.Line, prefix, stack)

		prefix = " > "
		goto startFrame
	}

	return stack
}
