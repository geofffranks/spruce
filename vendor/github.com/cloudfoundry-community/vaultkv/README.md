# vaultkv

## How to Use

This is a GoDoc: https://godoc.org/github.com/cloudfoundry-community/vaultkv

If you want to do anything with this library, then you'll need to make a
Client object. The Client object will need, at the very least, its `VaultURI`
member populated. AuthToken should be set to your bearer token for Vault. If
you need a bearer token created from some other auth method, you can call one
of the AuthX functions (currently, we support Github, LDAP, and Userpass). An
http client can be optionally provided (if not, then `http.DefaultClient`
will be used). If you would like to see information about the requests and
responses, then you can optionally provide an io.Writer for trace logs to be
streamed to.

```go
func main() {
  vault := &vaultkv.Client{
  AuthToken: "01234567-89ab-cdef-0123-456789abcdef",
    VaultURL: vaultURI,
    Client: &http.Client{
      Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
          InsecureSkipVerify: true,
        },
      },
    },
    Trace: os.Stdout,
  }

  output := struct{
    Bar string `json:"bar"`
  }{}
  err := vault.Get("secret/foo", &output)
  if err != nil {
    os.Exit(1)
  }

  fmt.Printf("output.Bar is `%s'\n", output.Bar)
}
```

## Testing

Run `./test` in the base directory to test all supported Vault versions. Run `./test latest` to test only the latest supported version of Vault.