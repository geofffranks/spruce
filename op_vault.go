package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// The VaultOperator provides a means of injecting credentials and
// other secrets from a Vault (vaultproject.io) Secure Key Storage
// instance.
type VaultOperator struct{}

// Setup ...
func (VaultOperator) Setup() error {
	return nil
}

// Phase identifies what phase of document management the vault
// operator should be evaulated in.  Vault lives in the Eval phase
func (VaultOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies collects implicit dependencies that a given
// `(( vault ... ))` call has.  There are none.
func (VaultOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

// Run executes the `(( vault ... ))` operator call, which entails
// interacting with the (unsealed) Vault instance to retrieve the
// given secrets.
func (VaultOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( vault ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( vault ... )) operation at $.%s\n", ev.Here)

	// syntax: (( vault "secret/path:key" ))
	if len(args) != 1 {
		return nil, fmt.Errorf("vault operator requires exactly one argument")
	}

	v, err := args[0].Resolve(ev.Tree)
	if err != nil {
		DEBUG("  arg[0]: failed to resolve expression to a concrete value")
		DEBUG("     [0]: error was: %s", err)
		return nil, err
	}

	var key string
	switch v.Type {
	case Literal:
		DEBUG("  arg[0]: using string literal '%v'", v.Literal)
		key = fmt.Sprintf("%v", v.Literal)

	case Reference:
		DEBUG("  arg[0]: trying to resolve reference $.%s", v.Reference)
		s, err := v.Reference.Resolve(ev.Tree)
		if err != nil {
			DEBUG("     [0]: resolution failed\n    error: %s", err)
			return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
		}

		switch s.(type) {
		case map[interface{}]interface{}:
			DEBUG("  arg[0]: %v is not a string scalar", s)
			return nil, fmt.Errorf("tried to look up $.%s, which is not a string scalar", v.Reference)

		case []interface{}:
			DEBUG("  arg[0]: %v is not a string scalar", s)
			return nil, fmt.Errorf("tried to look up $.%s, which is not a string scalar", v.Reference)

		default:
			key = fmt.Sprintf("%v", s)
		}

	default:
		DEBUG("  arg[0]: I don't know what to do with '%v'", args[0])
		return nil, fmt.Errorf("vault operator only accepts string literals and key reference arguments")
	}
	DEBUG("     [0]: Using vault key '%s'\n", key)

	secret := "REDACTED"
	if os.Getenv("VAULT_ADDR") != "" && os.Getenv("VAULT_TOKEN") != "" {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid argument %s; must be in the form path/to/secret:key", key)
		}
		secret, err = getVaultSecret(parts[0], parts[1])
		if err != nil {
			return nil, err
		}
	}

	return &Response{
		Type:  Replace,
		Value: secret,
	}, nil
}

func init() {
	RegisterOp("vault", VaultOperator{})
}

/****** VAULT INTEGRATION ***********************************/

func getVaultSecret(secret string, subkey string) (string, error) {
	vault := os.Getenv("VAULT_ADDR")
	DEBUG("  accessing the vault at %s", vault)

	url := fmt.Sprintf("%s/v1/%s", vault, secret)
	DEBUG("  crafting GET %s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		DEBUG("    !! failed to craft API request:\n    !! %s\n", err)
		return "", fmt.Errorf("failed to retrieve %s:%s from Vault (%s): %s",
			secret, subkey, vault, err)
	}
	req.Header.Add("X-Vault-Token", os.Getenv("VAULT_TOKEN"))

	DEBUG("  issuing GET %s", url)
	res, err := client.Do(req)
	if err != nil {
		DEBUG("    !! failed to issue API request:\n    !! %s\n", err)
		return "", fmt.Errorf("failed to retrieve %s:%s from Vault (%s): %s",
			secret, subkey, vault, err)
	}
	defer res.Body.Close()

	TRACE("    reading response body")
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		DEBUG("    !! failed to read JSON:\n    !! %s\n", err)
		return "", fmt.Errorf("failed to retrieve %s:%s from Vault (%s): %s",
			secret, subkey, vault, err)
	}

	TRACE("    decoding raw JSON:\n%s\n", string(b))
	var raw struct {
		Data   map[string]interface{}
		Errors []string
	}
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&raw)
	if err != nil {
		DEBUG("    !! failed to decode JSON:\n    !! %s\n", err)
		return "", fmt.Errorf("bad JSON response received from Vault: \"%s\"", string(b))
	}
	if len(raw.Errors) > 0 {
		DEBUG("    !! error: %s", raw.Errors[0])
		return "", fmt.Errorf("failed to retrieve %s:%s from Vault (%s): %s",
			secret, subkey, vault, raw.Errors[0])
	}

	DEBUG("  extracting the [%s] subkey from the secret", subkey)
	v, ok := raw.Data[subkey]
	if !ok {
		DEBUG("    !! %s:%s not found!\n", secret, subkey)
		return "", fmt.Errorf("secret %s:%s not found", secret, subkey)
	}
	if _, ok := v.(string); !ok {
		DEBUG("    !! %s:%s is not a string!\n", secret, subkey)
		return "", fmt.Errorf("secret %s:%s is not a string", secret, subkey)
	}

	DEBUG("  success.")
	return v.(string), nil
}
