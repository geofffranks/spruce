package vaultkv

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"strings"
)

func (v *Client) doSysRequest(
	method, path string,
	input interface{},
	output interface{}) error {
	err := v.doRequest(method, path, input, output)
	//In sys contexts, 400 can mean that the Vault is uninitialized.
	if _, is400 := err.(*ErrBadRequest); is400 {
		initialized, initErr := v.IsInitialized()
		if initErr != nil {
			return initErr
		}

		if !initialized {
			return &ErrUninitialized{message: "Your Vault is not initialized"}
		}
	}

	return err
}

//IsInitialized returns true if the targeted Vault is initialized
func (v *Client) IsInitialized() (is bool, err error) {
	//Don't call doSysRequest from here because it calls IsInitialized
	// and that could get ugly
	err = v.doRequest(
		"GET",
		"/sys/init",
		nil,
		&struct {
			Initialized *bool `json:"initialized"`
		}{
			Initialized: &is,
		})

	return
}

//SealState is the return value from Unseal and SealStatus. Type is only
//populated by SealStatus. ClusterName and ClusterID are only populated is
//Vault is unsealed.
type SealState struct {
	//Type is the type of unseal key. It is not returned from Unseal
	Type   string `json:"type,omitempty"`
	Sealed bool   `json:"sealed"`
	//Threshold is the number of keys required to reconstruct the master key
	Threshold int `json:"t"`
	//NumShares is the number of keys the master key has been split into
	NumShares int `json:"n"`
	//Progress is the number of keys that have been provided in the current unseal attempt
	Progress int    `json:"progress"`
	Nonce    string `json:"nonce"`
	Version  string `json:"version"`
	//ClusterName is only returned from an unsealed Vault.
	ClusterName string `json:"cluster_name,omitempty"`
	//ClusterID is only returned from an unsealed Vault.
	ClusterID string `json:"cluster_id,omitempty"`
}

//SealStatus calls the /sys/seal-status endpoint and returns the info therein
func (v *Client) SealStatus() (ret *SealState, err error) {
	err = v.doSysRequest(
		"GET",
		"/sys/seal-status",
		nil,
		&ret)

	return
}

//InitConfig is the information passed to InitVault to configure the Vault.
//Shares and Threshold are required.
type InitConfig struct {
	//Split the master key into this many shares
	Shares int `json:"secret_shares"`
	//This many shares are required to reconstruct the master key
	Threshold       int      `json:"secret_threshold"`
	RootTokenPGPKey string   `json:"root_token_pgp_key"`
	PGPKeys         []string `json:"pgp_keys"`
}

//InitVaultOutput is the return value of InitVault, and contains the generated
//Keys and RootToken.
type InitVaultOutput struct {
	client     *Client
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

//Unseal takes the keys in the InitVaultOutput object and sends each one to the
//unseal endpoint. If any of the unseal calls are unsuccessful, an error is
//returned.
func (i *InitVaultOutput) Unseal() error {
	for _, key := range i.Keys {
		sealState, err := i.client.Unseal(key)
		if err != nil {
			return err
		}

		if !sealState.Sealed {
			break
		}
	}

	return nil
}

//InitVault puts to the /sys/init endpoint to initialize the Vault, and returns
// the root token and unseal keys that were generated. The token of the client
// object is automatically set to the root token if the init is successful.
//If the vault has already been initialized, this returns *ErrBadRequest
func (v *Client) InitVault(in InitConfig) (out *InitVaultOutput, err error) {
	out = &InitVaultOutput{}
	err = v.doSysRequest(
		"PUT",
		"/sys/init",
		&in,
		&out,
	)

	if err == nil {
		v.AuthToken = out.RootToken
	}

	out.client = v

	return
}

//Seal puts to the /sys/seal endpoint to seal the Vault.
// If the Vault is already sealed, this doesn't return an error.
// If the Vault is unsealed and an incorrect token is provided, then this
// returns *ErrForbidden. Newer versions of Vault (0.11.2+) APIs return errors
// if the Vault is uninitialized or already sealed. This function squelches
// these errors for consistency with earlier versions of Vault
func (v *Client) Seal() error {
	err := v.doSysRequest("PUT", "/sys/seal", nil, nil)
	if err != nil && (IsUninitialized(err) || IsSealed(err)) {
		err = nil
	}

	return err
}

//Unseal puts to the /sys/unseal endpoint with a single key to progress the
//unseal attempt. If the unseal was successful, then the Sealed member of the
//returned struct will be false. If the given unseal key is improperly
//formatted, an *ErrBadRequest is returned. If the vault is already unsealed,
//no error is returned
func (v *Client) Unseal(key string) (out *SealState, err error) {
	out = &SealState{}
	err = v.doSysRequest(
		"PUT",
		"/sys/unseal",
		&struct {
			Key string `json:"key"`
		}{
			Key: key,
		},
		&out,
	)

	if IsInternalServer(err) {
		if strings.Contains(err.Error(), "message authentication failed") {
			err = &ErrBadRequest{message: err.Error()}
		}
	}

	return
}

//ResetUnseal resets the current unseal attempt, such that the progress towards
//an unseal becomes 0. If the vault is unsealed, nothing happens and no error
//is returned.
func (v *Client) ResetUnseal() (err error) {
	err = v.doSysRequest(
		"PUT",
		"/sys/unseal",
		&struct {
			Reset bool `json:"reset"`
		}{
			Reset: true,
		},
		nil,
	)

	return
}

//Health gives information about the current state of the Vault. If standbyok
//is set to true, no error will be returned in the case that the targeted vault
//is a standby node. If the targeted node is a standby and standbyok is false,
//then ErrStandby will be returned. If the Vault is not yet initialized,
//ErrUninitialized will be returned. If the Vault is initialized but sealed,
//then ErrSealed will be returned. If none of these are the case, no error is
//returned.
func (v *Client) Health(standbyok bool) error {
	//Don't call doRequest from Health because ParseError calls Health
	query := url.Values{}
	if standbyok {
		query.Add("standbyok", "true")
	}

	resp, err := v.Curl("GET", "/sys/health", query, nil)
	if err != nil {
		return err
	}

	errorsStruct := apiError{}
	err = json.NewDecoder(resp.Body).Decode(&errorsStruct)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	errorMessage := strings.Join(errorsStruct.Errors, "\n")

	switch resp.StatusCode {
	case 200:
		err = nil
	case 429:
		err = &ErrStandby{message: errorMessage}
	case 501:
		err = &ErrUninitialized{message: errorMessage}
	case 503:
		err = &ErrSealed{message: errorMessage}
	default:
		err = errors.New(errorMessage)
	}

	return err
}
