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

func chdir(cmd runner, path string) (pwd string, err error) {
	pwd, err = os.Getwd()
	if err != nil {
		fmt.Fprintln(cmd.getErrs(), "Error! Can't get current directory")
		return
	}
	if err = os.Chdir(path); err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! Directory change failed. Path = %s\n", path)
		return
	}
	return
}
