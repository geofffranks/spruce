package vaultkv

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/hashicorp/cap/util"
	"github.com/hashicorp/go-secure-stdlib/base62"
)

// AuthOIDCMetadata is the metadata member set by AuthOIDC
type AuthOIDCMetadata struct {
	AuthURL string `json:"auth_url"`
}

// authHalts are the signals we want to interrupt our auth callback on.
// SIGTSTP is omitted for Windows.
var authHalts = []os.Signal{os.Interrupt, os.Kill}

// AuthOIDC is a shorthand for AuthOIDCMount against the default OIDC mountpoint,
// 'OIDC'.
func (v *Client) AuthOIDC(username, password string) (ret *AuthOutput, err error) {
	return v.AuthOIDCMount("OIDC")
}

type loginResponse struct {
	authOutput *AuthOutput
	err        error
}

// AuthOIDCMount submits the given username and password to the OIDC auth endpoint
// mounted at the given mountpoint, checking it against existing OIDC auth
// configurations. If auth is successful, then the AuthOutput object is returned,
// and this client's AuthToken is set to the returned token. Given mountpoint is
// relative to /v1/auth.
func (v *Client) AuthOIDCMount(mount string) (ret *AuthOutput, err error) {
	// handle ctrl-c while waiting for the callback
	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, authHalts...)
	defer signal.Stop(sigintCh)
	raw := &authOutputRaw{}

	authURL, clientNonce, err := fetchAuthURL(v, mount)
	if err != nil {
		return nil, err
	}
	doneCh := make(chan loginResponse)
	http.HandleFunc("/oidc/callback", callbackHandler(v, mount, clientNonce, doneCh))

	port := "8250"
	listenAddress := "localhost"
	listener, err := net.Listen("tcp", listenAddress+":"+port)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	fmt.Fprintf(os.Stderr, "Complete the login via your OIDC provider. Launching browser to:\n\n    %s\n\n\n", authURL)
	if err := util.OpenURL(authURL); err != nil {
		return nil, fmt.Errorf("failed to launch the browser , err=%w", err)
	}
	fmt.Fprintf(os.Stderr, "Waiting for OIDC authentication to complete...\n")

	// Start local server
	go func() {
		err := http.Serve(listener, nil)
		if err != nil && err != http.ErrServerClosed {
			doneCh <- loginResponse{nil, err}
		}
	}()
	// Wait for either the callback to finish, or a halt signal (e.g., SIGKILL, SIGINT, SIGTSTP) to be received or up to 2 minutes
	select {
	case s := <-doneCh:
		return s.authOutput, s.err
	case <-sigintCh:
		return nil, errors.New("Interrupted")
	case <-time.After(2 * time.Minute):
		return nil, errors.New("Timed out waiting for response from provider")
	}

	ret = raw.toFinal(AuthOIDCMetadata{})
	v.AuthToken = ret.ClientToken
	return
}
func fetchAuthURL(v *Client, mount string) (string, string, error) {
	//var authURL string

	clientNonce, err := base62.Random(20)
	if err != nil {
		return "", "", err
	}

	callbackPort := "8250"
	callbackMethod := "http"
	callbackHost := "localhost"
	redirectURI := fmt.Sprintf("%s://%s:%s/oidc/callback", callbackMethod, callbackHost, callbackPort)
	data := map[string]interface{}{
		// only default role is supported
		//"role":         role,
		"redirect_uri": redirectURI,
		"client_nonce": clientNonce,
	}
	raw := &authOutputRaw{}

	err = v.doRequest(
		"POST",
		fmt.Sprintf("auth/%s/oidc/auth_url", mount),
		data,
		&raw,
	)
	if err != nil {
		return "", "", err
	}

	authUrl := raw.Data["auth_url"].(string)
	return authUrl, clientNonce, err
}
func callbackHandler(v *Client, mount string, clientNonce string, doneCh chan<- loginResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var response string
		var authOutput *AuthOutput
		var err error
		defer func() {
			w.Write([]byte(response))
			doneCh <- loginResponse{authOutput, err}
		}()

		// TODO: consider checking for method for post for additional auth step if required
		raw := &authOutputRaw{}
		query := url.Values{}
		query.Add("state", req.FormValue("state"))
		query.Add("code", req.FormValue("code"))
		query.Add("id_token", req.FormValue("id_token"))
		query.Add("client_nonce", clientNonce)
		err = v.doRequest(
			"GET",
			fmt.Sprintf("auth/%s/oidc/callback", mount),
			query,
			&raw,
		)
		authOutput = &AuthOutput{}
		authOutput.ClientToken = raw.Auth.ClientToken

		successHtml := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<style>
				body {
					font-family: Arial, sans-serif;
					text-align: center;
					padding: 50px;
				}
				h1 {
					color: #4CAF50;
				}
				p {
					color: #333;
				}
			</style>
		</head>
		<body>
			<h1>Success!</h1>
			<p>Your request was successful.</p>
		</body>
		</html>`
		errorHtml := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
		  <title>500 Internal Server Error</title>
		</head>
		<body>
		  <h1>500 Internal Server Error</h1>
		  <p>Something went wrong on our end. We're working on fixing it, and we'll be back as soon as possible.</p>
		  <p>In the meantime, please try again later.</p>
		</body>
		</html>`
		if err != nil {
			fmt.Println("Error calling back to vault", err.Error())
			response = errorHtml
		} else {
			response = successHtml
		}
	}
}
