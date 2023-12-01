package hlive

import (
	"fmt"
	"os"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger is a global logger used when a logger is not available
var Logger zerolog.Logger

// LoggerDev is a global logger needed for developer warnings to avoid the need for panics
var LoggerDev zerolog.Logger

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	LoggerDev = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
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
