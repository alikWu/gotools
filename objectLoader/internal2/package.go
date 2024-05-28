package internal2

import (
	"io/ioutil"
	"log"

	"golang.org/x/tools/go/packages"
)

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
