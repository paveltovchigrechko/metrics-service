// Package buildinfo prints the build information of the service binaries.
package buildinfo

import (
	"fmt"
	"io"
)

const notAvailable = "N/A"

// Print writes the build version, date and commit to w, one per line.
// Empty values are printed as "N/A".
func Print(w io.Writer, version, date, commit string) {
	fmt.Fprintf(w, "Build version: %s\n", valueOrNA(version))
	fmt.Fprintf(w, "Build date: %s\n", valueOrNA(date))
	fmt.Fprintf(w, "Build commit: %s\n", valueOrNA(commit))
}

func valueOrNA(s string) string {
	if s == "" {
		return notAvailable
	}
	return s
}
