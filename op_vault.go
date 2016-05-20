package spruce

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	. "github.com/geofffranks/spruce/log"
	"github.com/jhunt/tree"
	"gopkg.in/yaml.v2"
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
func (VaultOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

// Run executes the `(( vault ... ))` operator call, which entails
// interacting with the (unsealed) Vault instance to retrieve the
// given secrets.
func (VaultOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( vault ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( vault ... )) operation at $.%s\n", ev.Here)

	// syntax: (( vault "secret/path:key" ))
	// syntax: (( vault path.object "to concat with" other.object ))
	if len(args) < 1 {
		return nil, fmt.Errorf("vault operator requires at least one argument")
	}

	var l []string
	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
			DEBUG("     [%d]: error was: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			DEBUG("  arg[%d]: using string literal '%v'", i, v.Literal)
			l = append(l, fmt.Sprintf("%v", v.Literal))

		case Reference:
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
			s, err := v.Reference.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
			}

			switch s.(type) {
			case map[interface{}]interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, fmt.Errorf("tried to look up $.%s, which is not a string scalar", v.Reference)

			case []interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, fmt.Errorf("tried to look up $.%s, which is not a string scalar", v.Reference)

			default:
				l = append(l, fmt.Sprintf("%v", s))
			}

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("vault operator only accepts string literals and key reference arguments")
		}
	}
	key := strings.Join(l, "")
	DEBUG("     [0]: Using vault key '%s'\n", key)

	secret := "REDACTED"
	var err error

	if os.Getenv("REDACT") == "" {
		/*
		   user is not okay with a redacted manifest.
		   try to look up vault connection details from:
		     1. Environment Variables VAULT_ADDR and VAULT_TOKEN
		     2. ~/.svtoken file, if it exists
		     3. ~/.vault-token file, if it exists
		*/

		url := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")

		if url == "" || token == "" {
			svtoken := struct {
				Vault string `yaml:"vault"`
				Token string `yaml:"token"`
			}{}
			b, err := ioutil.ReadFile(os.ExpandEnv("${HOME}/.svtoken"))
			if err == nil {
				err = yaml.Unmarshal(b, &svtoken)
				if err == nil {
					url = svtoken.Vault
					token = svtoken.Token
				}
			}
		}

		if token == "" {
			b, err := ioutil.ReadFile(fmt.Sprintf("%s/.vault-token", os.Getenv("HOME")))
			if err == nil {
				token = strings.TrimSuffix(string(b), "\n")
			}
		}

		if url == "" || token == "" {
			return nil, fmt.Errorf("Failed to determine Vault URL / token, and the $REDACT environment variable is not set.")
		}

		os.Setenv("VAULT_ADDR", url)
		os.Setenv("VAULT_TOKEN", token)

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

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: os.Getenv("VAULT_SKIP_VERIFY") != "",
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			req.Header.Add("X-Vault-Token", os.Getenv("VAULT_TOKEN"))
			return nil
		},
	}
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
