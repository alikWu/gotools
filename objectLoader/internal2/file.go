package internal2

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"sync"
)

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
