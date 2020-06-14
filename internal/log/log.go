package log

import (
	"fmt"
	"os"
)

func Debugf(format string, params ...interface{}) {
	if os.Getenv("SHELP_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format, params...)
	}
}
