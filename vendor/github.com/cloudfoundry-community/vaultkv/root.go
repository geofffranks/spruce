package vaultkv

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
)

//GenerateRoot has functions for generating a new root token. Create this
//object with NewGenerateRoot(). That function performs the necessary
//initialization for the process
type GenerateRoot struct {
	client *Client
	otp    []byte
	//Versions of Vault before 11.2 encode their 16 byte root tokens as base16 uuids
	// After that... they are encoded as base64 strings
	shouldNotUUIDify bool
	state            GenerateRootState
}

//GenerateRootState contains state information about the GenerateRoot operation
type GenerateRootState struct {
	Started  bool   `json:"started"`
	Nonce    string `json:"nonce"`
	Progress int    `json:"progress"`
	Required int    `json:"required"`
	//Vault versions >= 0.9.x return the value as encoded_token
	EncodedToken string `json:"encoded_token"`
	//Vault versions before 0.9.x returned the value as encoded_root_token
	EncodedRootToken string `json:"encoded_root_token"`
	//Vault versions before 0.11.2 had the user generate their own 16-byte root token
	// Versions after that have the API generate it for you (at a length that is
	// decided by the API).
	OTP       string `json:"otp"`
	OTPLength int    `json:"otp_length"`
	Complete  bool   `json:"complete"`
}

//NewGenerateRoot initializes and returns a new generate root object.
func (v *Client) NewGenerateRoot() (*GenerateRoot, error) {
	ret := GenerateRoot{
		client: v,
		otp:    make([]byte, 16),
	}
	_, err := rand.Read(ret.otp)
	if err != nil {
		return nil, fmt.Errorf("Could not generate random values")
	}

	base64OTP := make([]byte, base64.StdEncoding.EncodedLen(len(ret.otp)))
	base64.StdEncoding.Encode(base64OTP, ret.otp)

	err = v.doRequest("PUT", "/sys/generate-root/attempt",
		map[string]string{"otp": string(base64OTP)}, &ret.state)
	if err != nil && !IsBadRequest(err) {
		return nil, err
	}

	if ret.state.OTPLength != 0 || IsBadRequest(err) {
		//Then we need to let the API generate the root token
		err = ret.Cancel()
		if err != nil {
			return nil, err
		}

		//In 0.11.2 and 0.11.3, you can't provide an empty body or else Vault EOFs.
		// So you have to give an empty string otp to prompt Vault to make an otp
		// of the proper length for you. This was fixed in 0.11.4.
		err = v.doRequest("PUT", "/sys/generate-root/attempt",
			map[string]string{"otp": ""}, &ret.state)
		if err != nil {
			return nil, err
		}
		ret.otp = []byte(ret.state.OTP)
		ret.shouldNotUUIDify = true
	}

	return &ret, nil
}

var genRootRegexp = regexp.MustCompile("no root generation in progress")

//Submit gives keys to the generate root token operation specified by this
//*GenerateRoot object. Any keys beyond the current required amount are
//ignored. If the Rekey is successful after all keys have been sent, then done
//will be returned as true. If the threshold is reached and any of the keys
//were incorrect, an *ErrBadRequest is returned and done is false. In this
//case, the generate root is not cancelled, but is instead reset. No error is
//given for an incorrect key before the threshold is reached. An *ErrBadRequest
//may also be returned if there is no longer any generate root token operation
//in progress, but in this case, done will be returned as true. To retrieve the
//new keys after submitting enough existing keys, call RootToken() on the
//GenerateRoot object.
func (g *GenerateRoot) Submit(keys ...string) (done bool, err error) {
	for _, key := range keys {
		g.state, err = g.client.genRootSubmit(key, g.state.Nonce)
		if err != nil {
			if ebr, is400 := err.(*ErrBadRequest); is400 {
				g.state.Progress = 0
				//I really hate error string checking, but there's no good way that doesn't
				//require another API call (which could, in turn, err, and leave us in a
				//wrong state). This checks if the generate root token is no longer in
				// progress
				if genRootRegexp.MatchString(ebr.message) {
					done = true
				}
			}

			return
		}

		if g.state.Complete {
			break
		}
	}

	return g.state.Complete, nil
}

//Cancel cancels the current generate root operation
func (g *GenerateRoot) Cancel() error {
	return g.client.GenerateRootCancel()
}

//GenerateRootCancel cancels the current generate root operation
func (v *Client) GenerateRootCancel() error {
	return v.doSysRequest("DELETE", "/sys/generate-root/attempt", nil, nil)
}

func (v *Client) genRootSubmit(key string, nonce string) (ret GenerateRootState, err error) {
	err = v.doSysRequest(
		"PUT",
		"/sys/generate-root/update",
		&struct {
			Key   string `json:"key"`
			Nonce string `json:"nonce"`
		}{
			Key:   key,
			Nonce: nonce,
		},
		&ret,
	)

	return
}

//Remaining returns the number of keys yet required by this generate root token
//operation. This does not refresh state, and only reflects the last action of
//this GenerateRoot object.
func (g *GenerateRoot) Remaining() int {
	return g.state.Required - g.state.Progress
}

//State returns the current state of the generate root operation. This does not
//refresh state, and only reflects the last action of this GenerateRoot object.
func (g *GenerateRoot) State() GenerateRootState {
	return g.state
}

//RootToken returns the new root token from this operation if the operation has
//been successful. The return value is undefined if the operation is not yet
//successful.
func (g *GenerateRoot) RootToken() (string, error) {
	rawTok := g.state.EncodedToken
	if rawTok == "" {
		rawTok = g.state.EncodedRootToken
	}

	for len(rawTok)%4 != 0 {
		rawTok += "="
	}

	tok, err := base64.StdEncoding.DecodeString(rawTok)
	if err != nil {
		return "", fmt.Errorf("Could not decode base64 token: %s", err)
	}

	if len(tok) != len(g.otp) {
		return "", fmt.Errorf("token length / one-time password length mismatch (%d/%d)", len(tok), len(g.otp))
	}
	for i := 0; i < len(g.otp); i++ {
		tok[i] ^= g.otp[i]
	}

	ret := string(tok)
	if !g.shouldNotUUIDify {
		tokHex := make([]byte, hex.EncodedLen(len(tok)))
		hex.Encode(tokHex, tok)
		ret = fmt.Sprintf("%s-%s-%s-%s-%s",
			tokHex[0:8], tokHex[8:12], tokHex[12:16], tokHex[16:20], tokHex[20:])
	}

	return ret, nil
}
