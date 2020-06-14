// +build debug

package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ShowTrace shows call trace logs
func ShowTrace() {
	i := 1
	for {
		pt, file, line, ok := runtime.Caller(i)
		if !ok || !strings.Contains(file, "shelp") {
			break
		}
		funcName := strings.TrimPrefix(runtime.FuncForPC(pt).Name(), "github.com/progrhyme/")
		fmt.Fprintf(os.Stderr, "  at %s:L%d, func=%v\n", file, line, funcName)
		i++
	}
	fmt.Fprintln(os.Stderr, "")
}
