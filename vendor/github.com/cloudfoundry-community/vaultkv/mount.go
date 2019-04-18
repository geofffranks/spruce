package vaultkv

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	//MountTypeGeneric is what the key value backend was called prior to 0.8.0
	MountTypeGeneric = "generic"
	//MountTypeKV is the type string to get a Key Value backend
	MountTypeKV = "kv"
)

//Mount represents a backend mounted at a point in Vault.
type Mount struct {
	//The type of mount at this point
	Type        string
	Description string
	Config      *MountConfig
	Options     map[string]interface{}
}

//MountConfig specifies configuration options given when initializing a backend.
type MountConfig struct {
	DefaultLeaseTTL time.Duration
	MaxLeaseTTL     time.Duration
	PluginName      string
	ForceNoCache    bool
}

type mountListAPI struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Config      mountConfigListAPI     `json:"config"`
	Options     map[string]interface{} `json:"options"`
}

func (m mountListAPI) Parse() Mount {
	return Mount{
		Type:        m.Type,
		Description: m.Description,
		Config:      m.Config.Parse(),
		Options:     m.Options,
	}
}

type mountConfigListAPI struct {
	//time in seconds
	DefaultLeaseTTL int    `json:"default_lease_ttl"`
	MaxLeaseTTL     int    `json:"max_lease_ttl"`
	PluginName      string `json:"plugin_name"`
	ForceNoCache    bool   `json:"force_no_cache"`
}

func (m mountConfigListAPI) Parse() *MountConfig {
	return &MountConfig{
		DefaultLeaseTTL: time.Duration(m.DefaultLeaseTTL) * time.Second,
		MaxLeaseTTL:     time.Duration(m.MaxLeaseTTL) * time.Second,
		PluginName:      m.PluginName,
		ForceNoCache:    m.ForceNoCache,
	}
}

type mountConfigEnableAPI struct {
	DefaultLeaseTTL string `json:"default_lease_ttl,omitempty"`
	MaxLeaseTTL     string `json:"max_lease_ttl,omitempty"`
	PluginName      string `json:"plugin_name,omitempty"`
	ForceNoCache    bool   `json:"force_no_cache,omitempty"`
}

func newMountConfigEnableAPI(conf *MountConfig) *mountConfigEnableAPI {
	if conf == nil {
		return nil
	}

	return &mountConfigEnableAPI{
		DefaultLeaseTTL: func() string {
			if conf.DefaultLeaseTTL == 0 {
				return ""
			}
			return conf.DefaultLeaseTTL.String()
		}(),
		MaxLeaseTTL: func() string {
			if conf.MaxLeaseTTL == 0 {
				return ""
			}
			return conf.DefaultLeaseTTL.String()
		}(),
		PluginName:   conf.PluginName,
		ForceNoCache: conf.ForceNoCache,
	}
}

//ListMounts queries the Vault backend for a list of active mounts that can
// be seen with the current authentication token. It is returned as a map
// of mount points to mount information.
func (c *Client) ListMounts() (map[string]Mount, error) {
	output := map[string]interface{}{}
	//Prior to 1.10, the mount names were top level keys. Then, they duplicated the
	// information into "data" with other metadata in the top level keys. So we need
	// to check if the data key is there (and isn't just a mount name)
	err := c.doRequest("GET", "/sys/mounts", nil, &output)
	if err != nil {
		return nil, err
	}

	var mounts map[string]mountListAPI
	if dataKey, ok := output["data"]; ok {
		mounts = getMountList(dataKey)
	}

	if mounts == nil {
		mounts := getMountList(output)
		if mounts == nil {
			return nil, fmt.Errorf("Could not parse mount list")
		}
	}

	ret := map[string]Mount{}
	for k, v := range mounts {
		ret[strings.TrimRight(k, "/")] = v.Parse()
	}

	return ret, err
}

func getMountList(candidate interface{}) map[string]mountListAPI {
	//check if data key is not a mount name
	b, err := json.Marshal(&candidate)
	if err != nil {
		return nil
	}

	tmpOutput := map[string]mountListAPI{}
	err = json.Unmarshal(b, &tmpOutput)
	if err != nil {
		return nil
	}

	return tmpOutput
}

//KVMountOptions is a map[string]interface{} that can be given as the options
//when mounting a backend. It has Version manipulation functions to make life
//easier.
type KVMountOptions map[string]interface{}

//GetVersion retruns the version held in the KVMountOptions object
func (o KVMountOptions) GetVersion() int {
	if o == nil {
		return 1
	}

	version, hasExplicitVersion := o["version"]
	if !hasExplicitVersion {
		return 1
	}

	vStr := version.(string)
	ret, _ := strconv.Atoi(vStr)
	return ret
}

//WithVersion returns a new KVMountOptions object with the given version
func (o KVMountOptions) WithVersion(version int) KVMountOptions {
	if o == nil {
		o = make(map[string]interface{}, 1)
	}

	o["version"] = strconv.Itoa(version)
	return o
}

//EnableSecretsMount mounts a secrets backend at the given path, configured with
// the given Mount configuration.
func (c *Client) EnableSecretsMount(path string, config Mount) error {
	input := struct {
		Type        string                `json:"type"`
		Description string                `json:"description"`
		Config      *mountConfigEnableAPI `json:"config,omitempty"`
		Options     interface{}           `json:"options,omitempty"`
	}{
		Type:        config.Type,
		Description: config.Description,
		Config:      newMountConfigEnableAPI(config.Config),
		Options:     config.Options,
	}

	return c.doRequest("POST", fmt.Sprintf("/sys/mounts/%s", path), &input, nil)
}

//DisableSecretsMount deletes the mount at the given path.
func (c *Client) DisableSecretsMount(path string) error {
	return c.doRequest("DELETE", fmt.Sprintf("/sys/mounts/%s", path), nil, nil)
}
