package util

import (
	"fmt"
	"os"
)

func Exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func CheckError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, err.Error())
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}
