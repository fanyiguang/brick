package Go

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

var writer io.Writer = os.Stderr

func SetWriter(w any) error {
	if ww, ok := w.(io.Writer); ok {
		writer = ww
	} else {
		return errors.New("not io.Writer")
	}
	return nil
}

func Go(fun func()) {
	go func(f func()) {
		defer func() {
			if err := recover(); err != nil {
				Recover(err)
			}
		}()

		f()
	}(fun)
}

func Recover(err interface{}) {
	callers := make([]uintptr, 15)
	_ = runtime.Callers(3, callers)
	for k, _ := range callers {
		frame, _ := runtime.CallersFrames(callers[k : k+1]).Next()
		_, _ = writer.Write([]byte(fmt.Sprintf("runtime.CallersFrames failed, err: %v file: %v line: %v func: %v\n", err, frame.File, frame.Line, frame.Function)))
	}
	_, _ = writer.Write([]byte(strings.Repeat("-", 50) + "\n"))
}
