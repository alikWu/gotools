package internal

import (
	"go/ast"
	"go/token"
)

type AstGenerator interface {
	Generate() (*token.FileSet, *ast.File, error)
}
