package generator

import (
	"fmt"
	"go/types"
	"strings"
	"unicode"

	"github.com/miko/dgraph-lambda-go/codegen/config"
	"github.com/miko/dgraph-lambda-go/codegen/graphql"
	"github.com/miko/dgraph-lambda-go/codegen/parser"
	"github.com/miko/dgraph-lambda-go/codegen/rewriter"
	"github.com/pkg/errors"
)

func Generate(c *config.Config, p *parser.Tree, r *rewriter.Rewriter) error {

	if err := generateModel(c, p); err != nil {
		return errors.Wrap(err, "Could not generate model")
	}
	if err := generateFieldResolvers(c, p, r); err != nil {
		return errors.Wrap(err, "Could not generate field resolvers")
	}
	if err := generateQueryResolvers(c, p, r); err != nil {
		return errors.Wrap(err, "Could not generate query resolvers")
	}
	if err := generateMutationResolvers(c, p, r); err != nil {
		return errors.Wrap(err, "Could not generate mutation resolvers")
	}
	if err := generateMiddleware(c, p, r); err != nil {
		return errors.Wrap(err, "Could not generate middleware resolvers")
	}
	if err := generateWebhook(c, p, r); err != nil {
		return errors.Wrap(err, "Could not generate webhook resolvers")
	}
	if err := generateExecuter(c, p, r); err != nil {
		return errors.Wrap(err, "Could not generate executer")
	}
	return nil
}

func resolverRef(t *parser.GoType) string {
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

func pkgPath(t *types.Package) string {
	return t.Path()
}

func title(t string) string {
	return strings.Title(t)
}

func untitle(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func typeName(t *types.TypeName) string {
	return t.Name()
}

func pointer(t *parser.GoType, isArray bool) string {
	if !t.TypeName.Exported() {
		return t.TypeName.Name()
	} else {
		if isArray {
			return fmt.Sprintf("[]*%s", resolverRef(t))
		} else {
			return fmt.Sprintf("*%s", resolverRef(t))
		}
	}
}

func args(args []*parser.Argument) string {
	var arglist []string

	for _, arg := range args {
		arglist = append(arglist, fmt.Sprintf("%s", arg.Name))
	}
	return strings.Join(arglist, ",")
}

//Replace reserved words which could cause go compiler to fail
var ReservedMap map[string]string = map[string]string{
	"type": "xtype",
}

func argsW(args []*parser.Argument) string {
	var arglist []string

	for _, arg := range args {
		if replaced, exists := ReservedMap[arg.Name]; exists {
			arg.Name = replaced
		}
		arglist = append(arglist, fmt.Sprintf("%s %s", arg.Name, pointer(arg.GoType, arg.IsArray)))
	}
	return strings.Join(arglist, ",")
}

func returnValue(t *parser.GoType, isArray bool) string {
	defaultValue, err := graphql.GetDefaultStringValueForType(t.TypeName.Name())
	fmt.Println(t.TypeName.Name())
	if err != nil || isArray {
		return "nil"
	} else {
		return defaultValue
	}
}

func body(t *parser.GoType, isArray bool, key string, rewriter *rewriter.Rewriter) string {
	if val, ok := rewriter.RewriteBodies[key]; ok {
		return val
	} else {
		return fmt.Sprintf(`
	return %s, nil
`, returnValue(t, isArray))
	}
}

func middlewareBody(key string, rewriter *rewriter.Rewriter) string {
	if val, ok := rewriter.RewriteBodies[key]; ok {
		return val
	} else {
		return `
	return nil
`
	}
}

func is(key string, resolverType string) bool {
	return strings.HasPrefix(key, resolverType)
}
