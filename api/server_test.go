package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExecuterMock struct {
}

func (e ExecuterMock) Resolve(ctx context.Context, request *Request) ([]byte, *LambdaError) {
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
	em := &ExecuterMock{}
	lambda := New(em)

	var requests = []struct {
		body     *bytes.Buffer
		expected int
	}{
		{body: bytes.NewBufferString("{}"), expected: http.StatusBadRequest},
	}

	for _, request := range requests {
		req := httptest.NewRequest(http.MethodPost, "/graphql-worker", request.body)
		w := httptest.NewRecorder()

		lambda.Route(w, req)

		assert.Equal(t, request.expected, w.Result().StatusCode)
	}
}
