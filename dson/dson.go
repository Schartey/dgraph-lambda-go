package dson

import jsoniter "github.com/json-iterator/go"

var DqlJson = jsoniter.Config{
	TagKey: "dql",
}.Froze()

func Marshal(v interface{}) ([]byte, error) {
	return DqlJson.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return DqlJson.Unmarshal(data, v)
}
