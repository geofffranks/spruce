package botta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var client = &http.Client{}

// Prepares an HTTP request of `method` against `url`, encoding `data`
// as JSON for the request payload. Automatically sets the `Content-Type`
// and `Accept` headers to `application/json`.
func HttpRequest(method string, url string, data interface{}) (*http.Request, error) {
	var marshaled []byte
	var err error
	if data != nil {
		marshaled, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	body := bytes.NewBuffer(marshaled)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

// Prepares a GET request to the API located at `url`. Automatically
// sets the `Content-Type` and `Accept` headers to `application/json`.
func Get(url string) (*http.Request, error) {
	return HttpRequest("GET", url, nil)
}

// Prepares a POST request to the API located at `url`, using
// the provided `data`. Automatically sets the `Content-Type` and `Accept`
// headers to `application/json`, and automatically encodes `data`
// as a JSON payload to the request.
func Post(url string, data interface{}) (*http.Request, error) {
	return HttpRequest("POST", url, data)
}

// Prepares a PUT request to the API located at `url`, using
// the provided `data`. Automatically sets the `Content-Type` and `Accept`
// headers to `application/json`, and automatically encodes `data`
// as a JSON payload to the request.
func Put(url string, data interface{}) (*http.Request, error) {
	return HttpRequest("PUT", url, data)
}

// Prepares a PATCH request to the API located at `url`, using
// the provided `data`. Automatically sets the `Content-Type` and `Accept`
// headers to `application/json`, and automatically encodes `data`
// as a JSON payload to the request.
func Patch(url string, data interface{}) (*http.Request, error) {
	return HttpRequest("PATCH", url, data)
}

// Prepares a DELETE request to the API located at `url`. Automatically
// sets the `Content-Type` and `Accept` headers to `application/json`.
func Delete(url string) (*http.Request, error) {
	return HttpRequest("DELETE", url, nil)
}

// Executes an HTTP request, using botta's http client, and returns
// a validated and parsed response, ready for data inspection.
func Issue(req *http.Request) (*Response, error) {
	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return ParseResponse(r)
}

// Takes an HTTP response, and attempts to decode the payload
// as JSON. Returns an error if the payload was not valid JSON.
// If the response was a non-success HTTP status, a BadResponseCode
// error. If the JSON payload was successfully decoded, it is returned
// alongside the BadResponseCode error, so you can decode JSON error
// messages easier.
func ParseResponse(r *http.Response) (*Response, error) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Decode json first to include in Response if possible
	var o interface{}
	var jsonErr error
	if len(body) > 0 {
		jsonErr = json.Unmarshal(body, &o)
	}

	resp := &Response{
		HTTPResponse: r,
		Raw:          body,
		Data:         o,
	}

	// Return an error for failure status codes, include
	// the Response object, so end-users can decode JSON error messages
	// If the body wasn't json, we complain about the code anyway, and return
	// nil Data, but raw response + http message
	if r.StatusCode >= 400 {
		return resp, BadResponseCode{
			URL:        r.Request.URL.String(),
			StatusCode: r.StatusCode,
			Message:    string(body),
		}
	}

	// If we had a successful request, but invalid json, return an error
	// with the Response obj, so end-users can debug as they see fit
	if jsonErr != nil {
		return resp, jsonErr
	}

	// All went well, Return the Response
	return resp, nil
}

// Allows you to customize the HTTP client being used by Issue().
// This is super-useful if you need to ignore SSL certificates, use
// a proxy, or otherwise, modify the default HTTP client.
func SetClient(c *http.Client) {
	client = c
}

// Retrieves a reference to the HTTP client being used by Issue().
func Client() *http.Client {
	return client
}

// An error representing an HTTP response whose StatusCode was >= 400
type BadResponseCode struct {
	StatusCode int
	Message    string
	URL        string
}

// Formats an error message for the BadResponseCode error
func (e BadResponseCode) Error() string {
	return fmt.Sprintf("%s returned %d: %s", e.URL, e.StatusCode, e.Message)
}
