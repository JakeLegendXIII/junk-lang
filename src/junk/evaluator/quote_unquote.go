package evaluator

import (
	"junk/ast"
	"junk/object"
)

func quote(node ast.Node) object.Object {
	return &object.Quote{Node: node}
}
