package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"

	"github.com/alikWu/gotools/objectLoader/internal"
)

const template = `
package curPackage

import (
)

var beanFactory = make(map[string]interface{})

func Init() {
}

func GetObject(structName string) interface{} {
	return beanFactory[structName]
}
`

type injectStructsAstGenerator struct {
	targetPackage string
	targetDir     string
	curPackage    string
}

func NewInjectStructsAstGenerator(targetPackage string, targetDir string, curPackage string) internal.AstGenerator {
	return &injectStructsAstGenerator{targetPackage: targetPackage, targetDir: targetDir, curPackage: curPackage}
}

func (i injectStructsAstGenerator) Generate() (*token.FileSet, *ast.File, error) {
	allGoFiles, err := getAllGoFiles(i.targetDir)
	if err != nil {
		return nil, nil, err
	}

	package2StructsMap := make(map[string][]string)
	for dir, fileNames := range allGoFiles {
		var structs []string
		for _, fileName := range fileNames {
			curStructs, err := getAllStructs(fileName)
			if err != nil {
				return nil, nil, err
			}
			structs = append(structs, curStructs...)
		}
		fullPackage := strings.Replace(dir, i.targetDir, i.targetPackage, 1)
		package2StructsMap[fullPackage] = structs
	}

	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, "", template, 0)
	if err != nil {
		return nil, nil, err
	}
	f.Name = ast.NewIdent(i.curPackage)

	//add import
	var importSpecs []ast.Spec
	for packageName, _ := range package2StructsMap {
		importSpec := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "\"" + packageName + "\"",
			},
		}
		f.Imports = append(f.Imports, importSpec)
		importSpecs = append(importSpecs, importSpec)
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		genDecl.Specs = append(genDecl.Specs, importSpecs...)
	}

	//inject objects
	beanFactory := f.Scope.Objects["beanFactory"]

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "Init" {
			continue
		}

		for packageName, structNames := range package2StructsMap {
			for _, structName := range structNames {
				x := &ast.Ident{
					Name: "beanFactory",
					Obj:  beanFactory,
				}
				funcDecl.Body.List = append(funcDecl.Body.List, assignObjects(x, packageName, structName))
			}
		}
	}
	return fileSet, f, nil
}

func assignObjects(x *ast.Ident, packageName string, structName string) *ast.AssignStmt {
	assignStmt := &ast.AssignStmt{
		Lhs: nil,
		Tok: token.ASSIGN,
		Rhs: nil,
	}

	lhsExpr := &ast.IndexExpr{
		X: x,
		Index: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("\"%s.%s\"", packageName, structName),
		},
	}
	assignStmt.Lhs = append(assignStmt.Lhs, lhsExpr)

	index := strings.LastIndex(packageName, "/")
	briefPackageName := packageName[index+1:]
	rhsExpr := &ast.CallExpr{
		Fun:  ast.NewIdent("new"),
		Args: []ast.Expr{ast.NewIdent(fmt.Sprintf("%s.%s", briefPackageName, structName))},
	}
	assignStmt.Rhs = append(assignStmt.Rhs, rhsExpr)
	return assignStmt
}

func getAllStructs(fileName string) ([]string, error) {
	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, fileName, nil, 0)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			_, isStruct := typeSpec.Type.(*ast.StructType)
			if !isStruct {
				continue
			}
			res = append(res, typeSpec.Name.Name)
		}
	}
	return res, nil
}

func getAllGoFiles(dir string) (map[string][]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	res := make(map[string][]string)
	for _, file := range files {
		fullName := dir + "/" + file.Name()
		if file.IsDir() {
			curRes, err := getAllGoFiles(fullName)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			for k, v := range curRes {
				res[k] = v
			}
			continue
		}
		if strings.Contains(file.Name(), ".go") {
			res[dir] = append(res[dir], fullName)
		}
	}
	return res, nil
}
