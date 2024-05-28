package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

var (
	path   = flag.String("path", "", "target path where all structs will be injected; default is ./")
	output = flag.String("output", "struct_injector.go", "output file name; default srcdir/struct_injector.go")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of stringer:\n")
	fmt.Fprintf(os.Stderr, "\tinjector -path=[path] -output=[filename] \n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://github.com/alikWu/gotools/injector\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(*path) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	g := &Generator{}
	g.ParsePackage(*path)
	g.InjectAllTypes()
	g.WriteFile(*output)
}

type Generator struct {
	pkg *Package
	buf bytes.Buffer
}

func (g *Generator) ParsePackage(pkgPath string) {
	g.pkg = NewPackage(pkgPath)
}

func (g *Generator) InjectAllTypes() {
	pkgTypesMap := g.pkg.GetPublicTypeNames()

	targetPkg := os.Getenv("GOPACKAGE")

	// Print the header and package clause.
	g.print("// Code generated by injector %s; DO NOT EDIT.\n\n", strings.Join(os.Args[1:], " "))
	g.print("package %s\n\n", targetPkg)

	//import
	g.print("import (\n")
	for pkg, _ := range pkgTypesMap {
		g.print("\t\"%s\"\n", pkg)
	}
	g.print(")\n\n")

	g.print("var beanFactory = make(map[string]interface{})\n\n")

	g.print("func Init() {\n")
	for pkg, types := range pkgTypesMap {
		for _, tn := range types {
			g.print(fmt.Sprintf("\tbeanFactory[\"%s\"] = new(%s.%s)\n", pkg, pkg[strings.LastIndex(pkg, "/")+1:], tn))
		}
		g.print("\n")
	}
	g.print("}\n")

	g.print("func GetObject(structName string) interface{} {\n")
	g.print("\treturn beanFactory[structName]\n")
	g.print("}")
}

func (g *Generator) print(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) WriteFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	_, err = file.Write(g.buf.Bytes())
	return err
}

type Package struct {
	path    string
	files   []*File
	subPkgs []*Package
	curPkg  *packages.Package
}

func NewPackage(path string) *Package {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests: false,
		//BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages matching %v", len(pkgs), path)
	}

	p := &Package{
		path:   path,
		curPkg: pkgs[0],
	}

	var files []*File

	for _, goFile := range p.curPkg.GoFiles {
		files = append(files, NewFile(goFile, p))
	}
	p.files = files

	var subPkgs []*Package
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			subPkgs = append(subPkgs, NewPackage(path+"/"+fileInfo.Name()))
		}
	}
	p.subPkgs = subPkgs

	return p
}

func (p *Package) GetPublicTypeNames() map[string][]string {
	allTypeNames := make(map[string][]string)
	for _, file := range p.files {
		names := file.GetPublicTypeNames()
		allTypeNames[p.GetPkg()] = append(allTypeNames[p.GetPkg()], names...)
	}

	for _, subPkg := range p.subPkgs {
		curTypeNames := subPkg.GetPublicTypeNames()
		for pkg, typeNames := range curTypeNames {
			allTypeNames[pkg] = typeNames
		}
	}
	return allTypeNames
}

func (p *Package) GetPkg() string {
	return p.curPkg.PkgPath
}

type File struct {
	fileName string
	pkg      *Package

	mutex sync.Mutex
	file  *ast.File
}

func NewFile(fileName string, pkg *Package) *File {
	return &File{fileName: fileName, pkg: pkg}
}

func (f *File) GetPublicTypeNames() []string {
	f.parse()

	var publicTypeNames []string
	for _, decl := range f.file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			_, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			typeName := typeSpec.Name.Name
			if len(typeName) > 0 && (typeName[0] >= 'A' && typeName[0] <= 'Z') {
				publicTypeNames = append(publicTypeNames, typeName)
			}
		}
	}

	return publicTypeNames
}

func (f *File) parse() {
	if f.file != nil {
		return
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.file != nil {
		return
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, f.fileName, nil, 0)
	if err != nil {
		log.Fatalln(err)
	}

	f.file = file
}
