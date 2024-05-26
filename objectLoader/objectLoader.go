package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alikWu/gotools/objectLoader/internal/ast"
	"github.com/alikWu/gotools/objectLoader/internal/file"
)

func main() {
	var targetPackage string
	var targetDir string
	//github.com/alikWu/gotools/objectLoader/test/path1
	flag.StringVar(&targetPackage, "targetPackage", "", "target package where you want to inject all struct objects ")
	//./objectLoader/test/path1
	flag.StringVar(&targetDir, "targetDir", "", "target directory where you want to inject all struct objects ")
	flag.Parse()
	if len(targetPackage) == 0 {
		panic("please specify your target package")
	}
	if len(targetDir) == 0 {
		panic("please specify your target directory")
	}

	curPackage := os.Getenv("GOPACKAGE")
	fileGenerator := file.NewInjectStructsFileGenerator("./auto_loader.go")
	_, err := fileGenerator.Generate(ast.NewInjectStructsAstGenerator(targetPackage, targetDir, curPackage))
	if err != nil {
		fmt.Println(err)
	}
}
