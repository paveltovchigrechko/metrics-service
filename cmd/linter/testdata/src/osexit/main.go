package main

import (
	"os"
	sys "os"
)

func init() {
	os.Exit(1) // want `os.Exit outside the main function of the main package`
}

func helper() {
	os.Exit(1) // want `os.Exit outside the main function of the main package`
}

func aliasedImport() {
	sys.Exit(1) // want `os.Exit outside the main function of the main package`
}

type app struct{}

func (app) main() {
	os.Exit(1) // want `os.Exit outside the main function of the main package`
}

type process struct{}

func (process) Exit(int) {}

func notReported() {
	process{}.Exit(1)
	_ = os.Getpid()
}

func main() {
	defer func() {
		os.Exit(2)
	}()
	os.Exit(0)
}
