package tools

import (
	"fmt"
	"go/types"
	"strings"
	"unicode"

	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func ResolverRef(t *parser.GoType) string {
	if t.TypeName.Pkg() != nil {
		/*for _, te := range autobind {
			if te == t.TypeName.Pkg().Path() {
				return fmt.Sprintf("%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
			}
		}*/
		if t.TypeName.Exported() {
			return fmt.Sprintf("%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		}
	}
	return t.TypeName.Name()
}

func ModelRef(t *parser.GoType, isArray bool) string {
	if t.TypeName.Exported() && !t.IsDefaultPackage {
		if isArray {
			return fmt.Sprintf("[]*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		} else {
			return fmt.Sprintf("*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		}
	}
	if t.TypeName.Exported() {
		if isArray {
			return fmt.Sprintf("[]*%s", t.TypeName.Name())
		} else {
			return fmt.Sprintf("*%s", t.TypeName.Name())
		}
	}
	if isArray {
		return fmt.Sprintf("[]%s", t.TypeName.Name())
	} else {
		return t.TypeName.Name()
	}
}

func JsonRef(t *parser.GoType) string {
	switch t.TypeName.Name() {
	case "string":
		return "\"%s\""
	case "float64":
		return "%f"
	case "int64":
		return "%d"
	}
	return "%s"
}

func JsonRefVal(field *parser.Field, model *parser.GoType) string {
	if !field.TypeName.Exported() {
		return fmt.Sprintf("%s.%s", Untitle(model.TypeName.Name()), Title(field.Name))
	} else {
		// TODO: Support other standard types
		if field.TypeName.Name() == "Time" {
			return fmt.Sprintf("MarshalTime(%s.%s)", Untitle(model.TypeName.Name()), Title(field.Name))
		} else {
			return fmt.Sprintf("%s.%s.Marshal()", Untitle(model.TypeName.Name()), field.TypeName.Name())
		}
	}
}

func JsonVal(field *parser.Field) string {
	if !field.TypeName.Exported() {
		switch field.TypeName.Name() {
		case "string":
			return fmt.Sprintf("string(v.GetStringBytes(\"%s\"))", Untitle(field.Name))
		case "float64":
			return fmt.Sprintf("v.GetFloat64(\"%s\")", Untitle(field.Name))
		case "int64":
			return fmt.Sprintf("v.GetInt64(\"%s\")", Untitle(field.Name))
		}
		return "nil"
	} else {
		// TODO: Support other standard types
		if field.TypeName.Name() == "Time" {
			return fmt.Sprintf("UnmarshalTime(v.GetStringBytes(\"%s\"))", Untitle(field.Name))
		} else {
			return fmt.Sprintf("Unmarshal%s(v.Get(\"%s\"))", Title(field.Name), Untitle(field.Name))
		}
	}
}

func PkgPath(t *types.Package) string {
	return t.Path()
}

func Title(t string) string {
	return strings.Title(t)
}

func Untitle(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func TypeName(t *types.TypeName) string {
	return t.Name()
}

func Pointer(t *parser.GoType, isArray bool) string {
	if !t.TypeName.Exported() {
		return t.TypeName.Name()
	} else {
		if isArray {
			return fmt.Sprintf("[]*%s", ResolverRef(t))
		} else {
			return fmt.Sprintf("*%s", ResolverRef(t))
		}
	}
}

func Args(args []*parser.Argument) string {
	var arglist []string

	for _, arg := range args {
		arglist = append(arglist, fmt.Sprintf("%s", arg.Name))
	}
	return strings.Join(arglist, ",")
}

func ArgsW(args []*parser.Argument) string {
	var arglist []string

	for _, arg := range args {
		arglist = append(arglist, fmt.Sprintf("%s %s", arg.Name, Pointer(arg.GoType, arg.IsArray)))
	}
	return strings.Join(arglist, ",")
}

func ReturnValue(t *parser.GoType, isArray bool) string {
	defaultValue, err := graphql.GetDefaultStringValueForType(t.TypeName.Name())
	fmt.Println(t.TypeName.Name())
	if err != nil || isArray {
		return "nil"
	} else {
		return defaultValue
	}
}

func Body(t *parser.GoType, isArray bool, key string, rewriter *rewriter.Rewriter) string {
	if val, ok := rewriter.RewriteBodies[key]; ok {
		return val
	} else {
		return fmt.Sprintf(`
	return %s, nil
`, ReturnValue(t, isArray))
	}
}

func MiddlewareBody(key string, rewriter *rewriter.Rewriter) string {
	if val, ok := rewriter.RewriteBodies[key]; ok {
		return val
	} else {
		return `
	return nil
`
	}
}

func Is(key string, resolverType string) bool {
	return strings.HasPrefix(key, resolverType)
}
