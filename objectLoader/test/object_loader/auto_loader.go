//Code generated by objectLoader DO NOT EDIT 
package object_loader

import (
	"github.com/alikWu/gotools/objectLoader/test/path1"
	"github.com/alikWu/gotools/objectLoader/test/path1/path2"
)

var beanFactory = make(map[string]interface{})

func Init() {
	beanFactory["github.com/alikWu/gotools/objectLoader/test/path1.A"] = new(path1.A)
	beanFactory["github.com/alikWu/gotools/objectLoader/test/path1/path2.B"] = new(path2.B)
}

func GetObject(structName string) interface{} {
	return beanFactory[structName]
}