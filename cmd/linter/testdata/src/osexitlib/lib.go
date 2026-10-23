package osexitlib

import "os"

func main() {
	os.Exit(1) // want `os.Exit outside the main function of the main package`
}
