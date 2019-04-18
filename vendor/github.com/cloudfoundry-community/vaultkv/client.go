//Package vaultkv provides a client with functions that make API calls that a user of
// Vault may commonly want.
package vaultkv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//Client provides functions that access and abstract the Vault API.
// VaultURL must be set to the for the client to work. Only Vault versions
// 0.6.5 and above are tested to work with this client.
type Client struct {
	AuthToken string
	VaultURL  *url.URL
	//If Client is nil, http.DefaultClient will be used
	Client *http.Client
	//If Trace is non-nil, information about HTTP requests will be given into the
	//Writer.
	Trace io.Writer
}

type vaultResponse struct {
	Data interface{} `json:"data"`
	//There's totally more to the response, but this is all I care about atm.
}

//URL encoded values can be given as a *url.Values as "input" when performing
// a GET call
func (v *Client) doRequest(
	method, path string,
	input interface{},
	output interface{}) error {

	u := *v.VaultURL
	u.Path = fmt.Sprintf("/v1/%s", strings.Trim(path, "/"))
	if u.Port() == "" {
		u.Host = fmt.Sprintf("%s:8200", u.Host)
	}

	var query url.Values
	var body io.Reader
	if input != nil {
		if strings.ToUpper(method) == "GET" {
			//Input has to be a url.Values
			query = input.(url.Values)
		} else {
			body = &bytes.Buffer{}
			err := json.NewEncoder(body.(*bytes.Buffer)).Encode(input)
			if err != nil {
				return err
			}
		}
	}

	resp, err := v.Curl(method, path, query, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return v.parseError(resp)
	}

	if output != nil && resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(output)
	}

	return err
}

//Curl takes the given path, prepends <VaultURL>/v1/ to it, and makes the request
// with the remainder of the given parameters. Errors returned only reflect
// transport errors, not HTTP semantic errors
func (v *Client) Curl(method string, path string, urlQuery url.Values, body io.Reader) (*http.Response, error) {
	//Setup URL
	u := *v.VaultURL
	u.Path = fmt.Sprintf("/v1/%s", strings.Trim(path, "/"))
	if u.Port() == "" {
		u.Host = fmt.Sprintf("%s:8200", u.Host)
	}
	u.RawQuery = urlQuery.Encode()

	//Do the request
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if v.Trace != nil {
		dump, _ := httputil.DumpRequest(req, true)
		v.Trace.Write([]byte(fmt.Sprintf("Request:\n%s\n", dump)))
	}

	token := v.AuthToken
	if token == "" {
		token = "01234567-89ab-cdef-0123-456789abcdef"
	}
	req.Header.Set("X-Vault-Token", token)

	client := v.Client
	if client == nil {
		client = http.DefaultClient
	}

	if client.CheckRedirect == nil {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return fmt.Errorf("Stopped after 10 redirects")
			}
			req.Header.Set("X-Vault-Token", token)
			return nil
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &ErrTransport{message: err.Error()}
	}

	if v.Trace != nil {
		dump, _ := httputil.DumpResponse(resp, true)
		v.Trace.Write([]byte(fmt.Sprintf("Response:\n%s\n", dump)))
	}

	return resp, nil
}
