package api

import (
	"context"
	"encoding/json"
)

type AuthHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Directive struct {
	Name      string                     `json:"name"`
	Arguments map[string]json.RawMessage `json:"arguments"`
}

type SelectionField struct {
	Alias        string                     `json:"alias"`
	Name         string                     `json:"name"`
	Arguments    map[string]json.RawMessage `json:"arguments"`
	Directives   []Directive                `json:"directives"`
	SelectionSet []SelectionField           `json:"slectionSet"`
}

type InfoField struct {
	Field SelectionField `json:"field"`
}

type Request struct {
	AccessToken string                     `json:"X-Dgraph-AccessToken"`
	Args        map[string]json.RawMessage `json:"args"`
	Field       InfoField                  `json:"info"`
	AuthHeader  AuthHeader                 `json:"authHeader"`
	Resolver    string                     `json:"resolver"`
	Parents     json.RawMessage            `json:"parents"`
	Event       *Event                     `json:"event"`
}

type Event struct {
	TypeName  string           `json:"__typename"`
	CommitTs  uint64           `json:"commitTs"`
	Operation string           `json:"operation"`
	Add       *AddEventInfo    `json:"add"`
	Update    *UpdateEventInfo `json:"update"`
	Delete    *DeleteEventInfo `json:"delete"`
}

type AddEventInfo struct {
	RootUIDs []string                 `json:"rootUIDs"`
	Input    []map[string]interface{} `json:"input"`
}

type UpdateEventInfo struct {
	RootUIDs    []string               `json:"rootUIDs"`
	SetPatch    map[string]interface{} `json:"setPatch"`
	RemovePatch map[string]interface{} `json:"removePatch"`
}

type DeleteEventInfo struct {
	RootUIDs []string `json:"rootUIDs"`
}

type MiddlewareFunc func(MiddlewareContext) MiddlewareContext

type MiddlewareContext struct {
	Ctx     context.Context
	Request *Request
}

type HandlerFunc func(ctx context.Context, input []byte, parents []byte, authHeader AuthHeader) (interface{}, error)
