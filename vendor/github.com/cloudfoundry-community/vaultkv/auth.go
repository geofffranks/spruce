package vaultkv

import "fmt"

//AuthOutput is the general structure as returned by AuthX functions. The
//Metadata member type is determined by the specific Auth function. Note that
//the Vault must be initialized and unsealed in order to use authentication
//endpoints.
type AuthOutput struct {
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
	Auth          struct {
		ClientToken string   `json:"client_token"`
		Accessor    string   `json:"accessor"`
		Policies    []string `json:"policies"`
	}
	//Metadata's internal structure is dependent on the auth type
	Metadata interface{} `json:"metadata"`
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
	ret = &AuthOutput{Metadata: AuthGithubMetadata{}}
	err = v.doRequest(
		"POST",
		"/auth/github/login",
		struct {
			Token string `json:"token"`
		}{Token: accessToken},
		&ret,
	)

	if err == nil {
		v.AuthToken = ret.Auth.ClientToken
	}

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
	ret = &AuthOutput{Metadata: AuthLDAPMetadata{}}
	err = v.doRequest(
		"POST",
		fmt.Sprintf("/auth/ldap/login/%s", username),
		struct {
			Password string `json:"password"`
		}{Password: password},
		&ret,
	)

	if err == nil {
		v.AuthToken = ret.Auth.ClientToken
	}

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
	ret = &AuthOutput{Metadata: AuthUserpassMetadata{}}
	err = v.doRequest(
		"POST",
		fmt.Sprintf("/auth/userpass/login/%s", username),
		struct {
			Password string `json:"password"`
		}{Password: password},
		&ret,
	)

	if err == nil {
		v.AuthToken = ret.Auth.ClientToken
	}

	return
}

//TokenRenewSelf takes the token in the Client object and attempts to renew its
// lease.
func (v *Client) TokenRenewSelf() (err error) {
	return v.doRequest("POST", "/auth/token/renew-self", nil, nil)
}

//TokenIsValid returns no error if it can look itself up. This can error
// if the token is valid but somebody has configured policies such that it can not
// look itself up. It can also error, of course, if the token is invalid.
func (v *Client) TokenIsValid() (err error) {
	return v.doRequest("GET", "/auth/token/lookup-self", nil, nil)
}
