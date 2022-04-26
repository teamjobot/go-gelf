package gelf

import (
	"runtime"
	"strings"
)

// getCaller returns the filename and the line info of a function
// further down in the call stack.  Passing 0 in as callDepth would
// return info on the function calling getCallerIgnoringLog, 1 the
// parent function, and so on.  Any suffixes passed to getCaller are
// path fragments like "/pkg/log/log.go", and functions in the call
// stack from that file are ignored.
func getCaller(callDepth int, suffixesToIgnore ...string) (file string, line int) {
	// bump by 1 to ignore the getCaller (this) stackframe
	callDepth++
outer:
	for {
		var ok bool
		_, file, line, ok = runtime.Caller(callDepth)
		if !ok {
			file = "???"
			line = 0
			break
		}

		if strings.Contains(file, "github.com/op/go-logging") {
			callDepth++
			continue outer
		}

		for _, s := range suffixesToIgnore {
			if strings.HasSuffix(file, s) {
				callDepth++
				continue outer
			}
		}
		break
	}
	return
}

func getCallerIgnoringLogMulti(callDepth int) (string, int) {
	// the +1 is to ignore this (getCallerIgnoringLogMulti) frame
	// Can't easily do just suffix w/o regex now since we use modules now with varying version number:
	// /home/ubuntu/gocode/pkg/mod/github.com/op/go-logging@v0.0.0-20160315200505-970db520ece7/log_nix.go
	return getCaller(
		callDepth+1,
		"/pkg/log/log.go",
		"src/log/log.go",
		"/pkg/io/multi.go",
		"go-logging/multi.go",
		"go-logging/log_nix.go",
		"go-logging/format.go",
		"go-logging/level.go",
		"go-logging/logger.go",
		"log_nix.go")
}
