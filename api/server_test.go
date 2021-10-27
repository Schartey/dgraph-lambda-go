package api

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ExecuterMock struct {
	mock.Mock
}

func (e ExecuterMock) Resolve(ctx context.Context, request *Request) ([]byte, *LambdaError) {
	e.Called(ctx, request)
	return nil, nil
}

var invalidRequests = []struct {
	body     string
	expected int
}{
	{body: "", expected: http.StatusBadRequest},
	{body: "[]", expected: http.StatusBadRequest},
	{body: "invalid", expected: http.StatusBadRequest},
	{body: "{", expected: http.StatusBadRequest},
	{body: "{}", expected: http.StatusBadRequest},
	{body: `{ "resolver":"" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"$webhook" }`, expected: http.StatusBadRequest},
	{body: `{ "event": { "operation":""} }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"", "event": { "operation":""} }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"User.test" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"User.test", "event": { "operation":""} }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"Query.test" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"Query.test", "parents": "" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"Query.test", "args": "" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"Mutation.test" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"Mutation.test", "parents": "" }`, expected: http.StatusBadRequest},
	{body: `{ "resolver":"Mutation.test", "args": "" }`, expected: http.StatusBadRequest},
}

var validRequests = []struct {
	body     string
	expected int
}{
	{body: `{ "resolver":"User.test", "parents": "" }`, expected: http.StatusOK},
	{body: `{ "resolver":"Query.test", "args": {} }`, expected: http.StatusOK},
	{body: `{ "resolver":"Mutation.test", "args": {} }`, expected: http.StatusOK},
	{body: `{ "resolver":"$webhook", "event": {} }`, expected: http.StatusOK},
}

func Test_Route_Invalid_Body(t *testing.T) {
	em := &ExecuterMock{}
	lambda := New(em)

	for _, request := range invalidRequests {
		req := httptest.NewRequest(http.MethodPost, "/graphql-worker", bytes.NewBufferString(request.body))
		w := httptest.NewRecorder()

		lambda.Route(w, req)

		assert.Equal(t, request.expected, w.Result().StatusCode)
	}
}

func Test_Route_Valid_Body(t *testing.T) {
	em := ExecuterMock{}
	em.On("Resolve", mock.Anything, mock.Anything).Return(nil, nil)

	lambda := New(em)

	for _, request := range validRequests {
		req := httptest.NewRequest(http.MethodPost, "/graphql-worker", bytes.NewBufferString(request.body))
		w := httptest.NewRecorder()

		lambda.Route(w, req)

		assert.Equal(t, request.expected, w.Result().StatusCode)
		em.AssertExpectations(t)
	}
}

func Test_Serve_Invalid_Body(t *testing.T) {
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	em := &ExecuterMock{}
	lambda := New(em)

	srv, err := lambda.Serve(httpServerExitDone)
	assert.NoError(t, err)

	for _, request := range invalidRequests {
		res, err := http.Post("http://localhost:8686/graphql-worker", "application/json", bytes.NewBufferString(request.body))
		assert.NoError(t, err)
		assert.Equal(t, request.expected, res.StatusCode)
		time.Sleep(1 * time.Second)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		assert.NoError(t, err)
	}
	httpServerExitDone.Wait()
}

func Test_Serve_Valid_Body(t *testing.T) {
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)

	em := &ExecuterMock{}
	em.On("Resolve", mock.Anything, mock.Anything).Return(nil, nil)
	lambda := New(em)

	srv, err := lambda.Serve(httpServerExitDone)
	assert.NoError(t, err)

	for _, request := range validRequests {
		res, err := http.Post("http://localhost:8686/graphql-worker", "application/json", bytes.NewBufferString(request.body))
		fmt.Println(res.Status)
		assert.NoError(t, err)
		assert.Equal(t, request.expected, res.StatusCode)
		em.AssertExpectations(t)
		time.Sleep(1 * time.Second)
	}
	if err := srv.Shutdown(context.TODO()); err != nil {
		assert.NoError(t, err)
	}
	httpServerExitDone.Wait()
}
