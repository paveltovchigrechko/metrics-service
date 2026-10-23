package main

import "log"

var _ = func() int {
	log.Fatal("boom") // want `log.Fatal outside the main function of the main package`
	return 0
}()

func init() {
	log.Fatal("boom") // want `log.Fatal outside the main function of the main package`
}

func helper() {
	log.Fatal("boom")         // want `log.Fatal outside the main function of the main package`
	log.Fatalf("boom: %d", 1) // want `log.Fatalf outside the main function of the main package`
	log.Fatalln("boom")       // want `log.Fatalln outside the main function of the main package`
}

func loggerMethod() {
	l := log.Default()
	l.Fatal("boom") // want `log.Fatal outside the main function of the main package`
}

type app struct{}

func (app) main() {
	log.Fatal("boom") // want `log.Fatal outside the main function of the main package`
}

type logger struct{}

func (logger) Fatal(...any) {}

func notReported() {
	logger{}.Fatal("not the log package")
	log.Println("not fatal")
}

func main() {
	log.Fatal("allowed in main")
	log.Fatalf("allowed in main: %d", 1)
	func() {
		log.Fatal("a closure inside main is still main")
	}()
}
