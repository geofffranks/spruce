// Package vaultkv provides a client with functions that make API calls that a user of
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
	"sync"
)

//Client provides functions that access and abstract the Vault API.
// VaultURL must be set to the for the client to work. Only Vault versions
// 0.6.5 and above are tested to work with this client.
type Client struct {
	AuthToken string
	VaultURL  *url.URL
	//If Client is nil, http.DefaultClient will be used.
	//
	//Note that vaultkv writes a CheckRedirect into whichever http.Client it
	//ends up using, on the first request, so that a Vault redirect carries
	//the current auth token. A CheckRedirect already set on the client is
	//left alone. Callers that share this http.Client with other code should
	//know that: the field is written once, and a caller calling Do on the
	//same client concurrently with vaultkv's first request reads it while it
	//is being written.
	Client *http.Client
	//If Trace is non-nil, information about HTTP requests will be given into the
	//Writer.
	Trace io.Writer
	//Namespace, if non-empty, will send a X-Vault-Namespace header on requests with
	// the given value.
	Namespace    string
	tokenLock    sync.RWMutex
	redirectOnce sync.Once
}

// devFallbackToken is used as the X-Vault-Token when no AuthToken has been
// set, matching Vault's dev-mode root token.
const devFallbackToken = "01234567-89ab-cdef-0123-456789abcdef"

// redirectMu orders the installation itself. Two vaultkv Clients can share one
// http.Client, and every Client with a nil Client shares http.DefaultClient, so
// the write to CheckRedirect needs ordering that no single Client can provide.
// It is taken once per Client rather than once per request: redirectOnce's fast
// path is an atomic load, which is what a request pays.
var redirectMu sync.Mutex

// installRedirectPolicy sets the Vault redirect policy exactly once per
// vaultkv.Client, without racing concurrent first requests. The closure
// reads the token at redirect time, so SetAuthToken takes effect on
// later redirects. A caller-supplied CheckRedirect is left untouched.
// Known remainder: if two vaultkv.Clients share one http.Client, the
// closure pins the first vaultkv.Client's token source; that sharing
// already misroutes tokens on redirects today and is out of scope here.
func (v *Client) installRedirectPolicy(client *http.Client) {
	v.redirectOnce.Do(func() {
		redirectMu.Lock()
		defer redirectMu.Unlock()
		if client.CheckRedirect != nil {
			return
		}
		v.setRedirectPolicy(client)
	})
}

func (v *Client) setRedirectPolicy(client *http.Client) {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 10 {
			return fmt.Errorf("Stopped after 10 redirects")
		}
		// Stamp the token only on a hop to the configured Vault. When
		// Client.Client is nil this policy lands on http.DefaultClient,
		// which the whole process shares, and an unrelated redirect
		// through it must not carry a Vault token off to a third party.
		// A Vault standby redirecting to the active node is a different
		// host and so is not stamped here; the token still reaches it,
		// because Go copies X-Vault-Token onto the redirected request
		// itself, treating only Authorization and Cookie as sensitive.
		if !v.isVaultHost(req.URL.Host) {
			return nil
		}
		v.tokenLock.RLock()
		tok := v.AuthToken
		v.tokenLock.RUnlock()
		if tok == "" {
			tok = devFallbackToken
		}
		req.Header.Set("X-Vault-Token", tok)
		return nil
	}
}

// isVaultHost reports whether host addresses the configured Vault. It accepts
// the host as written and, when no port was configured, the port 8200 that
// Curl fills in when it builds a request URL.
func (v *Client) isVaultHost(host string) bool {
	if v.VaultURL == nil {
		return false
	}
	if host == v.VaultURL.Host {
		return true
	}

	return v.VaultURL.Port() == "" && host == fmt.Sprintf("%s:8200", v.VaultURL.Host)
}

type vaultResponse struct {
	Data interface{} `json:"data"`
	//There's totally more to the response, but this is all I care about atm.
}

// drainLimit is how much of a response body is worth discarding to keep the
// connection. On the success path the JSON decoder has already consumed the
// value, so the drain returns io.EOF at once; on an error path the body is a
// short JSON error object well under this bound.
const drainLimit = 4 << 10

// drainBody discards what is left of resp's body and closes it, so the
// connection returns to the keep-alive pool: an unread body otherwise costs
// the whole TCP+TLS connection on every Go toolchain before 1.27.
//
// Only a body whose length the server declared, and declared small, is
// drained. Reading is blocking and happens on the caller's goroutine, and Curl
// builds its requests with no context while the client sets no timeout, so a
// chunked or truncated body would otherwise pin that goroutine for as long as
// the server chose to stay silent. Giving up the connection is the cheaper
// loss.
//
// A drain error is surfaced through err only when nothing earlier failed,
// preserving the truncation signal the previous ReadAll gave.
func drainBody(resp *http.Response, err *error) {
	if resp.ContentLength >= 0 && resp.ContentLength <= drainLimit {
		if _, derr := io.CopyN(io.Discard, resp.Body, drainLimit); derr != nil && derr != io.EOF {
			if *err == nil {
				*err = derr
			}
		}
	}

	resp.Body.Close()
}

//URL encoded values can be given as a *url.Values as "input" when performing
// a GET call
func (v *Client) doRequest(
	method, path string,
	input interface{},
	output interface{}) (err error) {

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
	defer func() {
		drainBody(resp, &err)
	}()

	if resp.StatusCode/100 != 2 {
		return v.parseError(resp)
	}

	if output != nil && resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(output)
		if err != nil {
			if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
				return fmt.Errorf("Could not parse response body as JSON, and returned Content-Type is `%s'. Client may not be reaching Vault", contentType)
			}
			return err
		}
	}

	return nil
}

//Curl takes the given path, prepends <VaultURL>/v1/ to it, and makes the request
// with the remainder of the given parameters. Errors returned only reflect
// transport errors, not HTTP semantic errors
func (v *Client) Curl(method string, path string, urlQuery url.Values, body io.Reader) (*http.Response, error) {
	//Setup URL
	u := *v.VaultURL
	pathPrefix := strings.Trim(u.Path, "/")
	if pathPrefix != "" {
		pathPrefix = pathPrefix + "/"
	}
	u.Path = fmt.Sprintf("/%sv1/%s", pathPrefix, strings.Trim(path, "/"))
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
		_, _ = v.Trace.Write([]byte(fmt.Sprintf("Request:\n%s\n", dump)))
	}

	v.tokenLock.RLock()
	token := v.AuthToken
	v.tokenLock.RUnlock()
	if token == "" {
		token = devFallbackToken
	}
	req.Header.Set("X-Vault-Token", token)

	if v.Namespace != "" && !pathNamespaceBlacklisted(path) {
		req.Header.Set("X-Vault-Namespace", strings.Trim(v.Namespace, "/")+"/")
	}

	client := v.Client
	if client == nil {
		client = http.DefaultClient
	}

	v.installRedirectPolicy(client)

	resp, err := client.Do(req)
	if err != nil {
		if v.Trace != nil {
			_, _ = v.Trace.Write([]byte(fmt.Sprintf("transport err %s\n", err.Error())))
		}
		return nil, &ErrTransport{message: err.Error()}
	}

	if v.Trace != nil {
		dump, _ := httputil.DumpResponse(resp, true)
		_, _ = v.Trace.Write([]byte(fmt.Sprintf("Response:\n%s\n", dump)))
	}

	return resp, nil
}

var namespaceBlacklisted []string = []string{
	"sys/health",
	"sys/seal-status",
}

func pathNamespaceBlacklisted(path string) bool {
	path = strings.Trim(path, "/")
	for _, blacklisted := range namespaceBlacklisted {
		if path == blacklisted {
			return true
		}
	}

	return false
}
