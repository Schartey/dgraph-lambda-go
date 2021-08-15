package graphql

import (
	"errors"
	"fmt"
	goTypes "go/types"
	"reflect"
	"strings"

	"github.com/dgraph-io/dgraph/types"
	"github.com/vektah/gqlparser/v2/ast"
)

var inbuiltTypeToDgraph = map[string]string{
	"ID":           "uid",
	"Boolean":      "bool",
	"Int":          "int",
	"Int64":        "int",
	"Float":        "float",
	"String":       "string",
	"DateTime":     "dateTime",
	"Password":     "password",
	"Point":        "geo",
	"Polygon":      "geo",
	"MultiPolygon": "geo",
}

type GoTypeDefinition struct {
	TypeName string
	PkgName  string
}

// Not supported: PointList
func SchemaTypeToGoType(name string, astType *ast.Type) (*GoTypeDefinition, error) {
	dgraphTypeName := strings.ToLower(inbuiltTypeToDgraph[astType.Name()])
	typeId, ok := types.TypeForName(dgraphTypeName)
	if !ok {
		return nil, errors.New("TypeId not found")
	}
	val := types.ValueForType(typeId)
	reflectType := reflect.Indirect(reflect.ValueOf(val.Value)).Type()

	goType := reflectType.String()

	// Is Array
	if strings.HasPrefix(astType.String(), "[") && (strings.HasSuffix(astType.String(), "]") || strings.HasSuffix(astType.String(), "]!")) {
		goType = fmt.Sprintf("[]%s", goType)
	}

	return &GoTypeDefinition{TypeName: goType, PkgName: reflectType.PkgPath()}, nil
}

// Not supported: PointList
func SchemaDefToGo(def *ast.Definition) (*goTypes.Package, error) {
	dgraphTypeName := strings.ToLower(inbuiltTypeToDgraph[def.Name])
	typeId, ok := types.TypeForName(dgraphTypeName)
	if !ok {
		return nil, errors.New("TypeId not found")
	}
	val := types.ValueForType(typeId)
	reflectType := reflect.Indirect(reflect.ValueOf(val.Value)).Type()

	return goTypes.NewPackage(reflectType.PkgPath(), reflectType.String()), nil
}

func SchemaDefToGoDef(def *ast.Definition) (pkgPath string, typeName string, err error) {
	dgraphTypeName := strings.ToLower(inbuiltTypeToDgraph[def.Name])
	typeId, ok := types.TypeForName(dgraphTypeName)

	if !ok {
		return pkgPath, typeName, errors.New("TypeId not found")
	}
	val := types.ValueForType(typeId)
	reflectType := reflect.Indirect(reflect.ValueOf(val.Value)).Type()

	pkgPath = reflectType.PkgPath()
	typeName = reflectType.Name()

	return pkgPath, typeName, nil
}

func IsArray(name string) bool {
	return strings.HasPrefix(name, "[") && (strings.HasSuffix(name, "]") || strings.HasSuffix(name, "]!"))
}

func IsDgraphType(name string) bool {
	return inbuiltTypeToDgraph[name] != ""
}
