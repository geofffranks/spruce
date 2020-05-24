package vaultkv

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

//ErrBadRequest represents 400 status codes that are returned from the API.
//See: your fault.
type ErrBadRequest struct {
	message string
}

func (e *ErrBadRequest) Error() string {
	return fmt.Sprintf("400 Bad Request: %s", e.message)
}

//IsBadRequest returns true if the error is an ErrBadRequest
func IsBadRequest(err error) bool {
	_, is := err.(*ErrBadRequest)
	return is
}

//ErrForbidden represents 403 status codes returned from the API. This could be
// if your auth is wrong or expired, or you simply don't have access to do the
// particular thing you're trying to do. Check your privilege.
type ErrForbidden struct {
	message string
}

func (e *ErrForbidden) Error() string {
	return fmt.Sprintf("403 Forbidden: %s", e.message)
}

//IsForbidden returns true if the error is an ErrForbidden
func IsForbidden(err error) bool {
	_, is := err.(*ErrForbidden)
	return is
}

//ErrNotFound represents 404 status codes returned from the API. This could be
// either that the thing you're looking for doesn't exist, or in some cases
// that you don't have access to the thing you're looking for and that Vault is
// hiding it from you.
type ErrNotFound struct {
	message string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("404 Not Found: %s", e.message)
}

//IsNotFound returns true if the error is an ErrNotFound
func IsNotFound(err error) bool {
	_, is := err.(*ErrNotFound)
	return is
}

//ErrStandby is only returned from Health() if standbyok is set to false and the
// node you're querying is a standby.
type ErrStandby struct {
	message string
}

func (e *ErrStandby) Error() string {
	return fmt.Sprintf("503 Standby: %s", e.message)
}

//IsErrStandby returns true if the error is an ErrStandby
func IsErrStandby(err error) bool {
	_, is := err.(*ErrStandby)
	return is
}

//ErrInternalServer represents 500 status codes that are returned from the API.
//See: their fault.
type ErrInternalServer struct {
	message string
}

func (e *ErrInternalServer) Error() string {
	return fmt.Sprintf("500 Internal Server Error: %s", e.message)
}

//IsInternalServer returns true if the error is an ErrInternalServer
func IsInternalServer(err error) bool {
	_, is := err.(*ErrInternalServer)
	return is
}

//ErrSealed represents the 503 status code that is returned by Vault most
// commonly if the Vault is currently sealed, but could also represent the Vault
// being in a maintenance state.
type ErrSealed struct {
	message string
}

func (e *ErrSealed) Error() string {
	return fmt.Sprintf("503 Sealed: %s", e.message)
}

//IsSealed returns true if the error is an ErrSealed
func IsSealed(err error) bool {
	_, is := err.(*ErrSealed)
	return is
}

//ErrUninitialized represents a 503 status code being returned and the Vault
//being uninitialized.
type ErrUninitialized struct {
	message string
}

func (e *ErrUninitialized) Error() string {
	return fmt.Sprintf("503 Uninitialized: %s", e.message)
}

//IsUninitialized returns true if the error is an ErrUninitialized
func IsUninitialized(err error) bool {
	_, is := err.(*ErrUninitialized)
	return is
}

//ErrTransport is returned if an error was encountered trying to reach the API,
// as opposed to an error from the API, is returned
type ErrTransport struct {
	message string
}

func (e *ErrTransport) Error() string {
	return fmt.Sprintf("Transport Error: %s", e.message)
}

//IsTransport returns true if the error is an ErrTransport
func IsTransport(err error) bool {
	_, is := err.(*ErrTransport)
	return is
}

//ErrKVUnsupported is returned by the KV object when the user requests an
// operation that cannot be performed by the actual version of the KV backend
// that the KV object is abstracting
type ErrKVUnsupported struct {
	message string
}

func (e *ErrKVUnsupported) Error() string {
	return fmt.Sprintf("KV Version Support: %s", e.message)
	return e.message
}

//IsErrKVUnsupported returns true if the error is an ErrKVUnsupported
func IsErrKVUnsupported(err error) bool {
	_, is := err.(*ErrKVUnsupported)
	return is
}

type apiError struct {
	Errors []string `json:"errors"`
}

func (v *Client) parseError(r *http.Response) (err error) {
	errorsStruct := apiError{}
	err = json.NewDecoder(r.Body).Decode(&errorsStruct)
	if err != nil {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			return fmt.Errorf("Could not parse response body as JSON, and returned Content-Type is `%s'. Client may not be reaching Vault", contentType)
		}
		return err
	}
	errorMessage := strings.Join(errorsStruct.Errors, "\n")

	switch r.StatusCode {
	case 400:
		err = &ErrBadRequest{message: errorMessage}
	case 403:
		err = &ErrForbidden{message: errorMessage}
	case 404:
		err = &ErrNotFound{message: errorMessage}
	case 500:
		err = &ErrInternalServer{message: errorMessage}
	case 503:
		err = v.parse503(errorMessage)
	default:
		err = errors.New(errorMessage)
	}

	return
}

func (v *Client) parse503(message string) (err error) {
	err = v.Health(true)
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *ErrStandby:
		e.message = message
		err = e
	case *ErrUninitialized:
		e.message = message
		err = e
	case *ErrSealed:
		e.message = message
		err = e
	}

	return err
}
