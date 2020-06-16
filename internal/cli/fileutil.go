package cli

import (
	"fmt"
	"io"
	"os"
)

func isSymlink(path string, errs io.Writer) bool {
	link, err := os.Readlink(path)
	if link != "" {
		if err != nil {
			// Just in case
			fmt.Fprintf(errs, "Error! Reading link failed. Path = %s\n", path)
		}
		return true
	}
	return false
}
