package graphql

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/twpayne/go-geom"
	"github.com/vektah/gqlparser/v2/ast"
)

var inbuiltTypeToDgraph = map[string]string{
	// ID is interally int64, but is represented by hex string
	//	"ID":           "uid",
	"ID":           "string",
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

var defaultReturnForType = map[string]string{
	"string":  "\"\"",
	"int":     "0",
	"int64":   "0",
	"bool":    "false",
	"float":   "0.0",
	"float64": "0.0",
}

var bytes []byte
var i int64
var f float64
var b bool
var t time.Time
var s string
var g geom.T
var ui uint64

var typeNameMap = map[string]interface{}{
	"default":  &s,
	"binary":   &bytes,
	"int":      &i,
	"float":    &f,
	"bool":     &b,
	"datetime": &t,
	"geo":      &g,
	"uid":      &ui,
	"string":   &s,
	"password": &s,
}

type GoTypeDefinition struct {
	TypeName string
	PkgName  string
}

func SchemaDefToGoDef(def *ast.Definition) (pkgPath string, typeName string, err error) {
	dgraphTypeName := strings.ToLower(inbuiltTypeToDgraph[def.Name])
	t, ok := typeNameMap[dgraphTypeName]

	if !ok {
		return pkgPath, typeName, errors.New("TypeId not found")
	}
	reflectType := reflect.Indirect(reflect.ValueOf(t)).Type()

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

func GetDefaultStringValueForType(name string) (string, error) {
	if val, ok := defaultReturnForType[name]; !ok {
		return "", errors.New("Could not find default value")
	} else {
		return val, nil
	}
}
