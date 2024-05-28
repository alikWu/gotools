package main

import "fmt"

//go:generate stringer -type=KILL ./hello.go

type KILL int

const (
	K1 KILL = iota
	K2
	K3
	K4
	K5
	K6
	K7
	K8
	K9
	K10
	K11
)

func main() {
	name := "John"
	str := fmt.Sprintf("Hello, %v!", name)
	fmt.Println(str)
}

func _() {
	var x [1]struct{}
	_ = x[K2-1]
}
