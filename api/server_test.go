package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func Test_Route_Invalid_Body(t *testing.T) {
	em := &ExecuterMock{}
	lambda := New(em)

	var requests = []struct {
		body     *bytes.Buffer
		expected int
	}{
		{body: nil, expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(""), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString("[]"), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString("invalid"), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString("{"), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString("{}"), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "event": { "operation":""} }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"", "event": { "operation":""} }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"User.test" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"User.test", "event": { "operation":""} }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"Query.test" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"Query.test", "parents": "" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"Query.test", "args": "" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"Mutation.test" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"Mutation.test", "parents": "" }`), expected: http.StatusBadRequest},
		{body: bytes.NewBufferString(`{ "resolver":"Mutation.test", "args": "" }`), expected: http.StatusBadRequest},
	}

	for _, request := range requests {
		var req *http.Request
		if request.body == nil {
			req = httptest.NewRequest(http.MethodPost, "/graphql-worker", nil)
		} else {
			req = httptest.NewRequest(http.MethodPost, "/graphql-worker", request.body)
		}
		w := httptest.NewRecorder()

		lambda.Route(w, req)

		assert.Equal(t, request.expected, w.Result().StatusCode)
	}
}

func Test_Route_Valid_Body(t *testing.T) {
	em := ExecuterMock{}
	em.On("Resolve", mock.Anything, mock.Anything).Return(nil, nil)

	lambda := New(em)

	var requests = []struct {
		body     *bytes.Buffer
		expected int
	}{
		{body: bytes.NewBufferString(`{ "resolver":"User.test", "parents": "" }`), expected: http.StatusOK},
		{body: bytes.NewBufferString(`{ "resolver":"Query.test", "args": {} }`), expected: http.StatusOK},
		{body: bytes.NewBufferString(`{ "resolver":"Mutation.test", "args": {} }`), expected: http.StatusOK},
	}

	for _, request := range requests {
		req := httptest.NewRequest(http.MethodPost, "/graphql-worker", request.body)
		w := httptest.NewRecorder()

		lambda.Route(w, req)

		assert.Equal(t, request.expected, w.Result().StatusCode)
		em.AssertExpectations(t)
	}
}
