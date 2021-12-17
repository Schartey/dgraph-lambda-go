package wasm

import (
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fastjson"
)

// We have to check if this works...the client should always import this package

// This function is imported from JavaScript, as it doesn't define a body.
// You should define a function named 'main.add' in the WebAssembly 'env'
// module from JavaScript.
//export Log
func Log(string)

func UnmarshalRequest(data []byte) (*Request, error) {
	t := &Request{}
	var p fastjson.Parser
	v, err := p.Parse(string(data))
	if err != nil {
		return nil, err
	}
	t.AccessToken = string(v.GetStringBytes("X-Dgraph-AccessToken"))
	t.Args = v.GetArray("args")
	t.AuthHeader = AuthHeader{
		Key:   string(v.Get("authHeader").GetStringBytes("key")),
		Value: string(v.Get("authHeader").GetStringBytes("value")),
	}

	// TODO: Webhooks
	//t.Event = &Event{}

	// TODO: Support InfoField
	/*t.Info = InfoField{}
	t.Info.Field.Alias = string(v.Get("info").Get("field").GetStringBytes("alias"))
	t.Info.Field.Name = string(v.Get("info").Get("field").GetStringBytes("name"))
	t.Info.Field.Arguments = v.Get("info").Get("field").GetStringBytes("arguments")
	t.Info.Field.Directives = []Directive{}
	t.Info.Field.SelectionSet = []SelectionField{}*/

	t.Parents = v.GetArray("parents")

	t.Resolver = string(v.GetStringBytes("resolver"))
	Log(t.Resolver)

	return t, nil
}
func MarshalStringArray(strs []string) []byte {
	if strs == nil {
		return []byte("null")
	}
	var escStrs []string
	for _, str := range strs {
		escStrs = append(escStrs, fmt.Sprintf("\"%s\"", str))
	}
	return []byte(fmt.Sprintf("[%s]", strings.Join(escStrs, ",")))
}

func UnmarshalStringArray(v *fastjson.Value) []string {
	if v == nil {
		return nil
	}
	var data []string
	values, err := v.Array()
	if err != nil {
		return []string{}
	}
	for _, s := range values {
		data = append(data, s.String())
	}
	return data
}

func MarshalTime(t *time.Time) string {
	if t == nil {
		return "null"
	}
	if v, err := t.MarshalJSON(); err != nil {
		fmt.Println(err)
		return "null"
	} else {
		return string(v)
	}
}

func UnmarshalTime(data []byte) *time.Time {
	if len(data) == 0 {
		return nil
	}
	var t time.Time
	if err := t.UnmarshalText(data); err != nil {
		fmt.Println(err)
		return nil
	} else {
		return &t
	}
}
