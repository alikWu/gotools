package internal

import "os"

type FileGenerator interface {
	Generate(astGenerator AstGenerator) (*os.File, error)
}
