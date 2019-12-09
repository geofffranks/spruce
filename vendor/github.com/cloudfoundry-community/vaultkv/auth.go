package vaultkv

import (
	"encoding/json"
	"fmt"
	"time"
)

//SetAuthToken provides a thread-safe way to set the auth token for the client.
//Setting AuthToken directly is still valid, but may race if a coroutine can
//possibly make a request with the client while the AuthToken is being written
//to. This function handles a mutex which avoids that.
func (v *Client) SetAuthToken(token string) {
	v.tokenLock.Lock()
	v.AuthToken = token
	v.tokenLock.Unlock()
}

//AuthOutput is the general structure as returned by AuthX functions. The
//Metadata member type is determined by the specific Auth function. Note that
//the Vault must be initialized and unsealed in order to use authentication
//endpoints.
type AuthOutput struct {
	Renewable     bool
	LeaseDuration time.Duration
	ClientToken   string
	Accessor      string
	Policies      []string
	//Metadata's internal structure is dependent on the auth type
	Metadata interface{}
}

type authOutputRaw struct {
	Renewable     bool `json:"renewable"`
	LeaseDuration int  `json:"lease_duration"`
	Auth          struct {
		ClientToken   string                 `json:"client_token"`
		Accessor      string                 `json:"accessor"`
		Policies      []string               `json:"policies"`
		Renewable     bool                   `json:"renewable"`
		LeaseDuration int                    `json:"lease_duration"`
		Metadata      map[string]interface{} `json:"metadata"`
	} `json:"auth"`
	//Metadata's internal structure is dependent on the auth type
	Metadata map[string]interface{} `json:"metadata"`
}

func (a authOutputRaw) toFinal(m interface{}) *AuthOutput {
	ret := &AuthOutput{
		ClientToken:   a.Auth.ClientToken,
		Accessor:      a.Auth.Accessor,
		Policies:      a.Auth.Policies,
		Renewable:     a.Auth.Renewable || a.Renewable,
		LeaseDuration: time.Duration(a.LeaseDuration) * time.Second,
	}

	metadata := a.Metadata

	if len(metadata) == 0 {
		metadata = a.Auth.Metadata
	}

	if len(metadata) != 0 && m != nil {
		b, err := json.Marshal(&ret.Metadata)
		if err != nil {
			panic("could not marshal map that we created")
		}

		err = json.Unmarshal(b, &m)
		if err != nil {
			panic("could not unmarshal json that we created")
		}

		ret.Metadata = m
	}

	if ret.LeaseDuration == 0 {
		ret.LeaseDuration = time.Duration(a.Auth.LeaseDuration) * time.Second
	}

	return ret
}

//AuthGithubMetadata is the metadata member set by AuthGithub.
type AuthGithubMetadata struct {
	Username     string `json:"username"`
	Organization string `json:"org"`
}

//AuthGithub submits the given accessToken to the github auth endpoint, checking
// it against configurations for Github organizations. If the accessToken
// belongs to an authorized account, then the AuthOutput object is returned, and
// this client's AuthToken is set to the returned token.
func (v *Client) AuthGithub(accessToken string) (ret *AuthOutput, err error) {
	raw := &authOutputRaw{}
	err = v.doRequest(
		"POST",
		"/auth/github/login",
		struct {
			Token string `json:"token"`
		}{Token: accessToken},
		&raw,
	)
	if err != nil {
		return
	}
	ret = raw.toFinal(AuthGithubMetadata{})
	v.AuthToken = ret.ClientToken

	return
}

//AuthLDAPMetadata is the metadata member set by AuthLDAP
type AuthLDAPMetadata struct {
	Username string `json:"username"`
}

//AuthLDAP submits the given username and password to the LDAP auth endpoint,
//checking it against existing LDAP auth configurations. If auth is successful,
//then the AuthOutput object is returned, and this client's AuthToken is set to
//the returned token.
func (v *Client) AuthLDAP(username, password string) (ret *AuthOutput, err error) {
	raw := &authOutputRaw{}
	err = v.doRequest(
		"POST",
		fmt.Sprintf("/auth/ldap/login/%s", username),
		struct {
			Password string `json:"password"`
		}{Password: password},
		&raw,
	)
	if err != nil {
		return
	}

	ret = raw.toFinal(AuthLDAPMetadata{})
	v.AuthToken = ret.ClientToken

	return
}

//AuthUserpassMetadata is the metadata member set by AuthUserpass
type AuthUserpassMetadata struct {
	Username string `json:"username"`
}

//AuthUserpass submits the given username and password to the userpass auth
//endpoint. If a username with that password exists, then the AuthOutput object
//is returned, and this client's AuthToken is set to the returned token.
func (v *Client) AuthUserpass(username, password string) (ret *AuthOutput, err error) {
	raw := &authOutputRaw{}
	err = v.doRequest(
		"POST",
		fmt.Sprintf("/auth/userpass/login/%s", username),
		struct {
			Password string `json:"password"`
		}{Password: password},
		&raw,
	)
	if err != nil {
		return
	}

	ret = raw.toFinal(AuthUserpassMetadata{})
	v.AuthToken = ret.ClientToken

	return
}

func (v *Client) AuthApprole(roleID, secretID string) (ret *AuthOutput, err error) {
	raw := &authOutputRaw{}
	err = v.doRequest(
		"POST",
		"/auth/approle/login",
		struct {
			RoleID   string `json:"role_id"`
			SecretID string `json:"secret_id"`
		}{
			RoleID:   roleID,
			SecretID: secretID,
		},
		&raw,
	)
	if err != nil {
		return
	}

	ret = raw.toFinal(nil)
	v.AuthToken = ret.ClientToken

	return
}

//TokenRenewSelf takes the token in the Client object and attempts to renew its
// lease.
func (v *Client) TokenRenewSelf() (err error) {
	return v.doRequest("POST", "/auth/token/renew-self", nil, nil)
}

type TokenInfo struct {
	Accessor       string
	CreationTime   time.Time
	CreationTTL    time.Duration
	DisplayName    string
	EntityID       string
	ExpireTime     time.Time
	ExplicitMaxTTL time.Duration
	ID             string
	IssueTime      time.Time
	NumUses        int64
	Orphan         bool
	Path           string
	Policies       []string
	Renewable      bool
	TTL            time.Duration
}

type tokenInfoRaw struct {
	Data struct {
		Accessor       string   `json:"accessor"`
		CreationTime   int64    `json:"creation_time"`
		CreationTTL    int64    `json:"creation_ttl"`
		DisplayName    string   `json:"display_name"`
		EntityID       string   `json:"entity_id"`
		ExpireTime     string   `json:"expire_time"`
		ExplicitMaxTTL int64    `json:"explicit_max_ttl"`
		ID             string   `json:"id"`
		IssueTime      string   `json:"issue_time"`
		NumUses        int64    `json:"num_uses"`
		Orphan         bool     `json:"orphan"`
		Path           string   `json:"path"`
		Policies       []string `json:"policies"`
		Renewable      bool     `json:"renewable"`
		TTL            int64    `json:"ttl"`
	} `json:"data"`
}

//TokenInfoSelf returns the contents of the token self info endpoint of the vault
func (v *Client) TokenInfoSelf() (ret *TokenInfo, err error) {
	raw := tokenInfoRaw{}
	err = v.doRequest("GET", "/auth/token/lookup-self", nil, &raw)
	if err != nil {
		return
	}

	var expTime, issTime time.Time
	if raw.Data.ExpireTime != "" {
		expTime, err = time.Parse(time.RFC3339Nano, raw.Data.ExpireTime)
		if err != nil {
			return
		}
	}

	if raw.Data.IssueTime != "" {
		issTime, err = time.Parse(time.RFC3339Nano, raw.Data.IssueTime)
		if err != nil {
			return
		}
	}

	ret = &TokenInfo{
		Accessor:       raw.Data.Accessor,
		CreationTime:   time.Unix(raw.Data.CreationTime, 0),
		CreationTTL:    time.Duration(raw.Data.CreationTTL) * time.Second,
		DisplayName:    raw.Data.DisplayName,
		EntityID:       raw.Data.EntityID,
		ExpireTime:     expTime,
		ExplicitMaxTTL: time.Duration(raw.Data.ExplicitMaxTTL) * time.Second,
		ID:             raw.Data.ID,
		IssueTime:      issTime,
		NumUses:        raw.Data.NumUses,
		Orphan:         raw.Data.Orphan,
		Path:           raw.Data.Path,
		Policies:       raw.Data.Policies,
		Renewable:      raw.Data.Renewable,
		TTL:            time.Duration(raw.Data.TTL) * time.Second,
	}

	return
}

//TokenIsValid returns no error if it can look itself up. This can error
// if the token is valid but somebody has configured policies such that it can not
// look itself up. It can also error, of course, if the token is invalid.
func (v *Client) TokenIsValid() (err error) {
	return v.doRequest("GET", "/auth/token/lookup-self", nil, nil)
}
