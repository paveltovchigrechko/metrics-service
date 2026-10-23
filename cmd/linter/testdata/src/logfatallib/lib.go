package logfatallib

import "log"

func main() {
	log.Fatal("boom") // want `log.Fatal outside the main function of the main package`
}
