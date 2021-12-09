package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Interface for StandardExecuter and WASMExecuter
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
	if request.Resolver == "" {
		return errors.New("Resolver or Event missing")
	}
	if strings.HasPrefix(request.Resolver, "Query.") || strings.HasPrefix(request.Resolver, "Mutation.") {
		if request.Args == nil {
			return errors.New("Missing arguments for query/mutation")
		}
	} else if request.Resolver == "$webhook" && request.Event == nil {
		return errors.New("Webhook must have event")
	} else if request.Resolver != "$webhook" && request.Parents == nil {
		return errors.New("Missing parents for field resolver")
	}
	return nil
}

func (l *Lambda) Serve(wg *sync.WaitGroup) (*http.Server, error) {
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
	srv := &http.Server{
		Addr:    ":8686",
		Handler: r,
	}

	go func() {
		defer wg.Done()

		fmt.Println("Lambda listening on 8686")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return srv, nil
}
