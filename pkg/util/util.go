package util

import (
	"fmt"
	"os"
	"strings"
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

func ParseLabels(labels string) map[string]string {
	labels = strings.TrimSpace(labels)

	parsedEdgeLabels := make(map[string]string)
	for _, label := range strings.Split(labels, ",") {
		parts := strings.SplitN(label, "=", 1)
		switch len(parts) {
		case 1:
			parsedEdgeLabels[parts[0]] = ""
		case 2:
			parsedEdgeLabels[parts[0]] = parts[1]
		default:
		}
	}

	return parsedEdgeLabels
}
