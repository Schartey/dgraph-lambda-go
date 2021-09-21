package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type ExecuterInterface interface {
	Resolve(ctx context.Context, request *Request) ([]byte, *LambdaError)
}

type Lambda struct {
	Executor ExecuterInterface
}

func New(executer ExecuterInterface) *Lambda {
	return &Lambda{Executor: executer}
}

func (l *Lambda) Route(w http.ResponseWriter, r *http.Request) {
	res, err := l.resolve(w, r)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(int(err.Status))
		w.Write([]byte(err.Error()))
	}
	w.Write(res)
}

func (l *Lambda) resolve(w http.ResponseWriter, r *http.Request) ([]byte, *LambdaError) {
	decoder := json.NewDecoder(r.Body)

	var request *Request
	err := decoder.Decode(&request)
	if err != nil {
		return nil, &LambdaError{Underlying: err, Status: http.StatusBadRequest}
	}
	if request == nil {
		return nil, &LambdaError{Underlying: errors.New("body cannot be nil"), Status: http.StatusBadRequest}
	}
	err = l.validate(request)
	if err != nil {
		return nil, &LambdaError{Underlying: errors.Wrap(err, "Invalid request"), Status: http.StatusBadRequest}
	}

	return l.Executor.Resolve(r.Context(), request)
}

func (l *Lambda) validate(request *Request) error {
	if request.Resolver == "" && request.Event.Operation == "" {
		return errors.New("Resolver or Operation missing")
	}
	if request.Resolver != "" {
		if strings.HasPrefix(request.Resolver, "Query.") || strings.HasPrefix(request.Resolver, "Mutation.") {
			if request.Args == nil {
				return errors.New("Missing arguments for query/mutation")
			}
		} else {
			if request.Parents == nil {
				return errors.New("Missing parents for field resolver")
			}
		}
	}
	return nil
}

func (l *Lambda) Serve() error {
	/* Setup Router */
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/graphql-worker", func(w http.ResponseWriter, r *http.Request) {
		res, err := l.resolve(w, r)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(int(err.Status))
			w.Write([]byte(err.Error()))
		}
		w.Write(res)
	})
	fmt.Println("Lambda listening on 8686")
	fmt.Println(http.ListenAndServe(":8686", r))

	return nil
}
