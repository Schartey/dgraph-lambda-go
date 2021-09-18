package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type ExecuterInterface interface {
	Resolve(ctx context.Context, dbody DBody) ([]byte, *LambdaError)
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

	var dbody *DBody
	err := decoder.Decode(&dbody)
	if err != nil {
		fmt.Println(err.Error())
	}
	if dbody == nil {
		return nil, &LambdaError{Underlying: errors.New("body cannot be nil"), Status: http.StatusBadRequest}
	}

	return l.Executor.Resolve(r.Context(), *dbody)
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
