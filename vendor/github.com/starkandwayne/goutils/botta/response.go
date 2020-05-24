package botta

import (
	"net/http"

	"github.com/starkandwayne/goutils/tree"
)

// Represents the HTTP response from your API
type Response struct {
	HTTPResponse *http.Response
	Raw          []byte
	Data         interface{}
}

// Returns a string-typed value found in the JSON data
// returned in the response, at `path`
func (r *Response) StringVal(path string) (string, error) {
	return tree.FindString(r.Data, path)
}

// Returns a tree.Number-typed value found in the JSON data
// returned in the response, at `path`
func (r *Response) NumVal(path string) (tree.Number, error) {
	return tree.FindNum(r.Data, path)
}

// Returns a bool-typed value found in the JSON data
// returned in the response, at `path`
func (r *Response) BoolVal(path string) (bool, error) {
	return tree.FindBool(r.Data, path)
}

// Returns a map[string]interface{}-typed value found in the JSON data
// returned in the response, at `path`
func (r *Response) MapVal(path string) (map[string]interface{}, error) {
	return tree.FindMap(r.Data, path)
}

// Returns a []interface{}-typed value found in the JSON data
// returned in the response, at `path`
func (r *Response) ArrayVal(path string) ([]interface{}, error) {
	return tree.FindArray(r.Data, path)
}

// Returns an interface{} value found in the JSON data
// returned in the response, at `path`
func (r *Response) Val(path string) (interface{}, error) {
	return tree.Find(r.Data, path)
}
