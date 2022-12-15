package vaultkv

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

//Use some heuristics to determine what is the most likely mount path if Vault won't
// tell us with its old, crusty version
func mountPathDefault(path string) string {
	path = strings.TrimLeft(path, "/")
	var prefix string
	if strings.HasPrefix(path, "auth/") {
		path = strings.TrimPrefix(path, "auth/")
		prefix = "auth/"
	}

	return fmt.Sprintf("%s%s", prefix, strings.Split(path, "/")[0])
}

//IsKVv2Mount returns true if the mount is a version 2 KV mount and false
//otherwise. This will also simply return false if no mount exists at the given
//mount point or if the Vault is too old to have the API endpoint to look for
//the mount. If a different API error occurs, it will be propagated out.
func (c *Client) IsKVv2Mount(path string) (mountPath string, isV2 bool, err error) {
	path = strings.TrimPrefix(path, "/")
	output := struct {
		Data struct {
			Secret map[string]struct {
				Type    string `json:"type"`
				Options struct {
					Version string `json:"version"`
				} `json:"options"`
			} `json:"secret"`
		} `json:"data"`
	}{}

	err = c.doRequest(
		"GET",
		fmt.Sprintf("/sys/internal/ui/mounts"),
		nil, &output)

	mountPath = strings.Trim(mountPathDefault(path), "/")
	if err != nil {
		//If we got a 404, this version of Vault is too old to possibly have a v2 backend
		if _, is404 := err.(*ErrNotFound); is404 {
			err = nil
		}

		//Then either the token is invalid (and we should err) or this could be
		// an old version of Vault that doesn't have this endpoint yet, and so its
		// interpreting this as a call to the sys/* region which this token may not have
		// access to. In this case, it would be too old of a version to have a v2 backend.
		if _, is403 := err.(*ErrForbidden); is403 {
			if c.TokenIsValid() == nil {
				err = nil
			}
		}

		return
	}

	if output.Data.Secret == nil {
		return
	}

	path = strings.Replace(path, "//", "/", -1)
	path = strings.Trim(path, "/")
	pathSplit := strings.Split(path, "/")

	for i := 1; i <= len(pathSplit); i++ {
		thisPath := strings.Join(pathSplit[:i], "/") + "/"
		if out, found := output.Data.Secret[thisPath]; found {
			mountPath = strings.TrimRight(thisPath, "/")
			isV2 = out.Options.Version == "2"
			break
		}
	}

	return
}

//V2Version is information about a version of a secret. The DeletedAt member
// will be nil to signify that a version is not deleted. Take note of the
// difference between "deleted" and "destroyed" - a deletion simply marks a
// secret as deleted, preventing it from being read. A destruction actually
// removes the data from storage irrevocably.
type V2Version struct {
	CreatedAt time.Time
	DeletedAt *time.Time
	Destroyed bool
	Version   uint
}

type v2VersionAPI struct {
	CreatedTime  string `json:"created_time"`
	DeletionTime string `json:"deletion_time"`
	Destroyed    bool   `json:"destroyed"`
	Version      uint   `json:"version"`
}

func (v v2VersionAPI) Parse() V2Version {
	ret := V2Version{
		Destroyed: v.Destroyed,
		Version:   v.Version,
	}

	//Parse those times
	ret.CreatedAt, _ = time.Parse(time.RFC3339Nano, v.CreatedTime)
	tmpDeletion, err := time.Parse(time.RFC3339Nano, v.DeletionTime)
	if err == nil {
		ret.DeletedAt = &tmpDeletion
	}

	return ret
}

//V2GetOpts are options to specify in a V2Get request.
type V2GetOpts struct {
	// Version is the version of the resource to retrieve. Setting this to zero (or
	// not setting it at all) will retrieve the latest version
	Version uint
}

//V2Get will get a secret from the given path in a KV version 2 secrets backend.
//If the secret is at "/bar" in the backend mounted at "foo", then the path
//should be "foo/bar". The response will be decoded into the item pointed to
//by output using encoding/json.Unmarshal semantics. The version to retrieve
//can be selected by setting Version in the V2GetOpts struct at opts.
func (c *Client) V2Get(mount, subpath string, output interface{}, opts *V2GetOpts) (meta V2Version, err error) {
	if output != nil &&
		reflect.ValueOf(output).Kind() != reflect.Ptr {
		err = fmt.Errorf("V2Get output target must be a pointer if non-nil")
		return
	}

	type outputData struct {
		Metadata v2VersionAPI `json:"metadata"`
		Data     interface{}  `json:"data"`
	}

	unmarshalInto := &struct {
		Data outputData `json:"data"`
	}{
		Data: outputData{
			Metadata: v2VersionAPI{},
			Data:     output,
		},
	}

	query := url.Values{}
	if opts != nil {
		query.Add("version", strconv.FormatUint(uint64(opts.Version), 10))
	}

	path := fmt.Sprintf("%s/data/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))
	err = c.doRequest("GET", path, query, unmarshalInto)
	if err != nil {
		return
	}

	meta = unmarshalInto.Data.Metadata.Parse()
	return
}

//V2List returns the list of paths nested directly under the given path. If this
//is not a "directory" for any paths, then ErrNotFound is returned. In the list
//of paths returned on success, if a path ends with a slash, then it is also a
//"directory". The Vault must be unsealed and initialized for this endpoint to
//work. No assumptions are made about the mounting point of your Key/Value
//backend.
func (c *Client) V2List(mount, subpath string) ([]string, error) {
	ret := []string{}
	path := fmt.Sprintf("%s/metadata/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))

	query := url.Values{}
	query.Add("list", "true")
	err := c.doRequest("GET", path, query, &vaultResponse{
		Data: &struct {
			Keys *[]string `json:"keys"`
		}{
			Keys: &ret,
		},
	})
	if err != nil {
		return nil, err
	}

	return ret, err
}

//V2SetOpts are options that can be specified to a V2Set call
type V2SetOpts struct {
	//CAS provides a check-and-set version number. If this is set to zero, then
	// the value will only be written if the key does not yet exist. If the CAS
	//number is non-zero, then this will only be written if the current version
	//for your this secret matches the CAS value.
	CAS *uint `json:"cas,omitempty"`
}

//WithCAS returns a pointer to a new V2SetOpts with the CAS value set to the
//given value. If i is zero, then the value will only be written if the key
//does not exist. If i is non-zero, then the value will only be written if the
//currently existing version matches i. Not calling CAS will result in no
//restriction on writing. If the mount is set up for requiring CAS, then not
//setting CAS with this function a valid number will result in a failure when
//attempting to write.
func (s V2SetOpts) WithCAS(i uint) *V2SetOpts {
	s.CAS = new(uint)
	*s.CAS = i
	return &s
}

//V2Set uses encoding/json.Marshal on the object given in values to encode
// the secret as JSON, and writes it to the path given. Populate ops to use the
// check-and-set functionality. Returns the metadata about the written secret
// if the write is successful.
func (c *Client) V2Set(mount, subpath string, values interface{}, opts *V2SetOpts) (meta V2Version, err error) {
	input := struct {
		Options *V2SetOpts  `json:"options,omitempty"`
		Data    interface{} `json:"data"`
	}{
		Options: opts,
		Data:    &values,
	}

	output := struct {
		Data v2VersionAPI `json:"data"`
	}{
		Data: v2VersionAPI{},
	}

	path := fmt.Sprintf("%s/data/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))

	err = c.doRequest("PUT", path, &input, &output)
	if err != nil {
		return
	}

	meta = output.Data.Parse()
	return
}

//V2DeleteOpts are options that can be provided to a V2Delete call.
type V2DeleteOpts struct {
	Versions []uint `json:"versions"`
}

//V2Delete marks a secret version at the given path as deleted. If opts is not
// provided or the Versions slice therein is left nil, the latest version is
// deleted. Otherwise, the specified versions are deleted. Note that the deleted
// data from this call is recoverable from a call to V2Undelete.
func (c *Client) V2Delete(mount, subpath string, opts *V2DeleteOpts) error {
	method := "DELETE"
	path := fmt.Sprintf("%s/data/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))

	if opts != nil && len(opts.Versions) > 0 {
		method = "POST"
		path = fmt.Sprintf("%s/delete/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))
	} else {
		opts = nil
	}

	return c.doRequest(method, path, opts, nil)
}

//V2Undelete marks the specified versions at the specified paths as not deleted.
func (c *Client) V2Undelete(mount, subpath string, versions []uint) error {
	path := fmt.Sprintf("%s/undelete/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))
	return c.doRequest("POST", path, struct {
		Versions []uint `json:"versions"`
	}{
		Versions: versions,
	}, nil)
}

//V2Destroy permanently deletes the specified versions at the specified path.
func (c *Client) V2Destroy(mount, subpath string, versions []uint) error {
	path := fmt.Sprintf("%s/destroy/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))
	return c.doRequest("POST", path, struct {
		Versions []uint `json:"versions"`
	}{
		Versions: versions,
	}, nil)
}

//V2DestroyMetadata permanently destroys all secret versions and all metadata
// associated with the secret at the specified path.
func (c *Client) V2DestroyMetadata(mount, subpath string) error {
	path := fmt.Sprintf("%s/metadata/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))
	return c.doRequest("DELETE", path, nil, nil)
}

//V2Metadata is the metadata associated with a secret
type V2Metadata struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	//CurrentVersion is the highest version number that has been created for this
	//secret. Deleteing or destroying the highest version does not change this
	//number.
	CurrentVersion uint
	OldestVersion  uint
	MaxVersions    uint
	Versions       []V2Version
}

type v2MetadataAPI struct {
	Data struct {
		CreatedTime    string                  `json:"created_time"`
		CurrentVersion uint                    `json:"current_version"`
		MaxVersions    uint                    `json:"max_versions"`
		OldestVersion  uint                    `json:"oldest_version"`
		UpdatedTime    string                  `json:"updated_time"`
		Versions       map[string]v2VersionAPI `json:"versions"`
	} `json:"data"`
}

//Version returns the version with the given number in the metadata as a
//V2Version object , if present. If no version with that number is present, an
//error is returned.
func (m V2Metadata) Version(number uint) (version V2Version, err error) {
	if len(m.Versions) == 0 {
		err = fmt.Errorf("That version does not exist in the metadata")
		return
	}

	firstVersion := m.Versions[0]
	index := int(number) - int(firstVersion.Version)
	if index < 0 || index > len(m.Versions) {
		err = fmt.Errorf("That version does not exist in the metadata")
		return
	}

	version = m.Versions[index]
	return
}

func (m v2MetadataAPI) Parse() V2Metadata {
	ret := V2Metadata{
		CurrentVersion: m.Data.CurrentVersion,
		MaxVersions:    m.Data.MaxVersions,
		OldestVersion:  m.Data.OldestVersion,
	}

	ret.CreatedAt, _ = time.Parse(time.RFC3339Nano, m.Data.CreatedTime)
	ret.UpdatedAt, _ = time.Parse(time.RFC3339Nano, m.Data.UpdatedTime)

	for number, metadata := range m.Data.Versions {
		toAdd := metadata.Parse()
		version64, _ := strconv.ParseUint(number, 10, 64)
		toAdd.Version = uint(version64)
		ret.Versions = append(ret.Versions, toAdd)
	}

	sort.Slice(ret.Versions,
		func(i, j int) bool { return ret.Versions[i].Version < ret.Versions[j].Version },
	)

	return ret
}

//V2GetMetadata gets the metadata associated with the secret at the specified
// path.
func (c *Client) V2GetMetadata(mount, subpath string) (meta V2Metadata, err error) {
	path := fmt.Sprintf("%s/metadata/%s", strings.Trim(mount, "/"), strings.Trim(subpath, "/"))
	output := v2MetadataAPI{}
	err = c.doRequest("GET", path, nil, &output)
	if err != nil {
		return
	}
	meta = output.Parse()
	return
}
