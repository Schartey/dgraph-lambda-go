package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/dgraph-io/dgo/v210"
	"gitlab.com/trendsnap/trendgraph/dgraph-lambda-go/request"
	"google.golang.org/grpc"
)

type Resolver struct {
	plugins    []*plugin.Plugin
	middleware []MiddlewareFunc
	resolvers  map[string]HandlerFunc
	// graphql
	conn *grpc.ClientConn
	dql  *dgo.Dgraph
}

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type HandlerFunc func(context.Context, []byte, request.AuthHeader, *dgo.Dgraph) error

func NewResolver(dql *dgo.Dgraph) *Resolver {
	return &Resolver{resolvers: make(map[string]HandlerFunc), dql: dql}
}

func (r *Resolver) LoadPlugins(pluginPath string) error {

	err := r.registerPlugins(pluginPath)
	if err != nil {
		return err
	}
	return nil
}

// Use adds middleware to the chain which is run after router.
func (r *Resolver) ResolveFunc(resolver string, handlerFunc HandlerFunc) {
	r.resolvers[resolver] = handlerFunc
}

// Use adds middleware to the chain which is run after router.
func (r *Resolver) Use(middleware ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *Resolver) Resolve(ctx context.Context, dbody *request.DBody) error {
	args, err := json.Marshal(dbody.Args)
	if err != nil {
		fmt.Println("Could not marshal arguments")
	}

	h := r.resolvers[dbody.Resolver]

	h = applyMiddleware(h, r.middleware...)

	err = h(ctx, args, dbody.AuthHeader, r.dql)
	if err != nil {
		return err
	}
	return nil
}

func (r *Resolver) close() {
	r.conn.Close()
}

func (r *Resolver) registerPlugins(path string) error {
	r.plugins = nil

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".so") {
				fmt.Printf("Loading plugin: %s\n", info.Name())
				p, err := plugin.Open(path)
				if err != nil {
					fmt.Printf("Could not load plugin %s: %s\n", info.Name(), err.Error())
				} else {
					r.plugins = append(r.plugins, p)
					register, err := p.Lookup("Register")
					if err != nil {
						fmt.Printf("Could not call Register on plugin %s: %s\n", info.Name(), err.Error())
					} else {
						register.(func(*Resolver))(r)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("%d Plugins found\n", len(r.plugins))

	return nil
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
