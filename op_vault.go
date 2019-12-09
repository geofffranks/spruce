package spruce

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cloudfoundry-community/vaultkv"
	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/yaml"
)

var kv *vaultkv.KV = nil

var vaultSecretCache = map[string]map[string]interface{}{}

//VaultRefs maps secret path to paths in YAML structure which call for it
var VaultRefs = map[string][]string{}

//SkipVault toggles whether calls to the Vault operator actually cause the
// Vault to be contacted and the keys substituted in.
var SkipVault bool

// The VaultOperator provides a means of injecting credentials and
// other secrets from a Vault (vaultproject.io) Secure Key Storage
// instance.
type VaultOperator struct{}

// Setup ...
func (VaultOperator) Setup() error {
	return nil
}

// Phase identifies what phase of document management the vault
// operator should be evaluated in.  Vault lives in the Eval phase
func (VaultOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies collects implicit dependencies that a given `(( vault ... ))`
// call has. There are no dependencies other that those given as args to the
// command.
func (VaultOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

func initializeVaultClient() error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	namespace := os.Getenv("VAULT_NAMESPACE")
	skip := false

	if addr == "" || token == "" {
		svtoken := struct {
			Vault      string `yaml:"vault"`
			Token      string `yaml:"token"`
			Namespace  string `yaml:"namespace"`
			SkipVerify bool   `yaml:"skip_verify"`
		}{}
		b, err := ioutil.ReadFile(os.ExpandEnv("${HOME}/.svtoken"))
		if err == nil {
			err = yaml.Unmarshal(b, &svtoken)
			if err == nil {
				addr = svtoken.Vault
				token = svtoken.Token
				namespace = svtoken.Namespace
				skip = svtoken.SkipVerify
			}
		}
	}

	if skipVaultVerify(os.Getenv("VAULT_SKIP_VERIFY")) {
		skip = true
	}

	if token == "" {
		b, err := ioutil.ReadFile(fmt.Sprintf("%s/.vault-token", os.Getenv("HOME")))
		if err == nil {
			token = strings.TrimSuffix(string(b), "\n")
		}
	}

	if addr == "" || token == "" {
		return fmt.Errorf("Failed to determine Vault URL / token, and the $REDACT environment variable is not set.")
	}

	roots, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("unable to retrieve system root certificate authorities: %s", err)
	}

	parsedURL, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("Could not parse Vault URL `%s': %s", addr, err)
	}

	if parsedURL.Port() == "" {
		if parsedURL.Scheme == "http" {
			parsedURL.Host = parsedURL.Host + ":80"
		} else {
			parsedURL.Host = parsedURL.Host + ":443"
		}
	}

	client := &vaultkv.Client{
		AuthToken: token,
		VaultURL:  parsedURL,
		Namespace: namespace,
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					RootCAs:            roots,
					InsecureSkipVerify: skip,
				},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				req.Header.Add("X-Vault-Token", token)
				req.Header.Add("X-Vault-Namespace", token)
				return nil
			},
		},
	}
	if DebugOn {
		client.Trace = os.Stderr
	}

	if err != nil {
		return fmt.Errorf("Error setting up Vault client: %s", err)
	}

	kv = client.NewKV()

	return nil
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
				return nil, ansi.Errorf("@R{tried to look up} @c{$.%s}@R{, which is not a string scalar}", v.Reference)

			case []interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, ansi.Errorf("@R{tried to look up} @c{$.%s}@R{, which is not a string scalar}", v.Reference)

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

	//Append the location from which this operator was called to the list of
	// places from which this key was referenced
	if refs, found := VaultRefs[key]; !found {
		VaultRefs[key] = []string{ev.Here.String()}
	} else {
		VaultRefs[key] = append(refs, ev.Here.String())
	}

	secret := "REDACTED"
	var err error

	if !SkipVault {
		/*
		   user is not okay with a redacted manifest.
		   try to look up vault connection details from:
		     1. Environment Variables VAULT_ADDR and VAULT_TOKEN
		     2. ~/.svtoken file, if it exists
		     3. ~/.vault-token file, if it exists
		*/

		if kv == nil {
			err := initializeVaultClient()
			if err != nil {
				return nil, fmt.Errorf("Error during Vault client initialization: %s", err)
			}
		}

		leftPart, rightPart := parsePath(key)
		if leftPart == "" || rightPart == "" {
			return nil, ansi.Errorf("@R{invalid argument} @c{%s}@R{; must be in the form} @m{path/to/secret:key}", key)
		}
		var fullSecret map[string]interface{}
		var found bool
		if fullSecret, found = vaultSecretCache[leftPart]; found {
			DEBUG("vault: Cache hit for `%s`", leftPart)
		} else {
			DEBUG("vault: Cache MISS for `%s`", leftPart)
			// Secret isn't cached. Grab it from the vault.
			fullSecret, err = getVaultSecret(leftPart)
			if err != nil {
				//Normalize the error messages
				switch err.(type) {
				case *vaultkv.ErrNotFound:
					err = fmt.Errorf("secret %s not found", key)
				}

				return nil, err
			}
			vaultSecretCache[leftPart] = fullSecret
		}

		secret, err = extractSubkey(fullSecret, leftPart, rightPart)
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

func getVaultSecret(secret string) (map[string]interface{}, error) {
	ret := map[string]interface{}{}

	DEBUG("Fetching Vault secret at `%s'", secret)
	_, err := kv.Get(secret, &ret, nil)
	if err != nil {
		DEBUG(" failure.")
		return nil, err
	}

	DEBUG("  success.")
	return ret, nil
}

func extractSubkey(secretMap map[string]interface{}, secret, subkey string) (string, error) {
	DEBUG("  extracting the [%s] subkey from the secret", subkey)

	secretSubkeyPath := fmt.Sprintf("%s:%s", secret, subkey)
	v, ok := secretMap[subkey]
	if !ok {
		DEBUG("    !! %s not found!\n", secretSubkeyPath)
		return "", ansi.Errorf("@R{secret} @c{%s} @R{not found}", secretSubkeyPath)
	}
	if _, ok := v.(string); !ok {
		DEBUG("    !! %s is not a string!\n", secretSubkeyPath)
		return "", ansi.Errorf("@R{secret} @c{%s} @R{is not a string}", secretSubkeyPath)
	}
	DEBUG(" success.")
	return v.(string), nil
}

func parsePath(path string) (secret, key string) {
	secret = path
	if idx := strings.LastIndex(path, ":"); idx >= 0 {
		secret = path[:idx]
		key = path[idx+1:]
	}
	return
}

func skipVaultVerify(env string) bool {
	env = strings.ToLower(env)
	if env == "" || env == "no" || env == "false" || env == "0" || env == "off" {
		return false
	}
	return true
}
