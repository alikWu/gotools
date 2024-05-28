package main

import (
	"os"

	"github.com/alikWu/gotools/objectLoader/internal2"
)

func main() {
	os.Setenv("GOPACKAGE", "main")
	g := &internal2.Generator{}
	g.ParsePackage("./objectLoader/test/path1/.")
	g.InjectAllTypes()
	g.WriteFile("test11.go")
}
