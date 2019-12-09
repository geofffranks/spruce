package vaultkv

import (
	"fmt"
	"strings"
	"sync"
)

//KV provides an abstraction to the Vault tree which makes dealing with
// the potential of both KV v1 and KV v2 backends easier to work with.
// KV v1 backends are exposed through this interface much like KV v2
// backends with only one version. There are limitations around Delete
// and Undelete calls because of the lack of versioning in KV v1 backends.
// See the documentation around those functions for more details.
// An empty KV struct is not request-ready. Please call Client.NewKV instead.
type KV struct {
	Client *Client
	//Map from mount name to [true if version 2. False otherwise]
	mounts map[string]kvMount
	lock   sync.RWMutex
}

type kvMount interface {
	Get(mount, subpath string, output interface{}, opts *KVGetOpts) (meta KVVersion, err error)
	Set(mount, subpath string, values map[string]string, opts *KVSetOpts) (meta KVVersion, err error)
	List(mount, subpath string) (paths []string, err error)
	Delete(mount, subpath string, opts *KVDeleteOpts) (err error)
	Undelete(mount, subpath string, versions []uint) (err error)
	Destroy(mount, subpath string, versions []uint) (err error)
	DestroyAll(mount, subpath string) (err error)
	Versions(mount, subpath string) (ret []KVVersion, err error)
	MountVersion() (version uint)
}

/*====================
        KV V1
====================*/
type kvv1Mount struct {
	client *Client
}

func v1ConstructPath(mount, subpath string) string {
	mount = strings.Trim(mount, "/")
	subpath = strings.Trim(subpath, "/")
	return strings.Trim(fmt.Sprintf("%s/%s", mount, subpath), "/")
}

func (k kvv1Mount) Get(mount, subpath string, output interface{}, opts *KVGetOpts) (meta KVVersion, err error) {
	if opts != nil && opts.Version > 1 {
		err = &ErrNotFound{"No versions greater than one in KV v1 backend"}
		return
	}

	path := v1ConstructPath(mount, subpath)
	err = k.client.Get(path, output)
	if err == nil {
		meta.Version = 1
	}
	return
}

func (k kvv1Mount) List(mount, subpath string) (paths []string, err error) {
	path := v1ConstructPath(mount, subpath)
	return k.client.List(path)
}

func (k kvv1Mount) Set(mount, subpath string, values map[string]string, opts *KVSetOpts) (meta KVVersion, err error) {
	path := v1ConstructPath(mount, subpath)
	err = k.client.Set(path, values)
	if err == nil {
		meta.Version = 1
	}
	return
}

func (k kvv1Mount) Delete(mount, subpath string, opts *KVDeleteOpts) (err error) {
	if opts == nil || !opts.V1Destroy {
		return &ErrKVUnsupported{"Refusing to destroy KV v1 value from delete call"}
	}

	//opts should be non-nil here because of the check earlier in the function
	return k.Destroy(mount, subpath, opts.Versions)
}

func (k kvv1Mount) Undelete(mount, subpath string, versions []uint) (err error) {
	return &ErrKVUnsupported{"Cannot undelete secret in KV v1 backend"}
}

func (k kvv1Mount) Destroy(mount, subpath string, versions []uint) (err error) {
	shouldDelete := len(versions) == 0
	for _, v := range versions {
		if v <= 1 {
			shouldDelete = true
		}
	}

	if shouldDelete {
		path := v1ConstructPath(mount, subpath)
		err = k.client.Delete(path)
	}
	return err
}

func (k kvv1Mount) DestroyAll(mount, subpath string) (err error) {
	path := v1ConstructPath(mount, subpath)
	return k.client.Delete(path)
}

func (k kvv1Mount) Versions(mount, subpath string) (ret []KVVersion, err error) {
	path := v1ConstructPath(mount, subpath)
	err = k.client.Get(path, nil)
	if err != nil {
		return nil, err
	}
	ret = []KVVersion{{Version: 1}}
	return
}

func (k kvv1Mount) MountVersion() (version uint) {
	return 1
}

/*====================
        KV V2
====================*/
type kvv2Mount struct {
	client *Client
}

func (k kvv2Mount) Get(mount, subpath string, output interface{}, opts *KVGetOpts) (meta KVVersion, err error) {
	var o *V2GetOpts
	if opts != nil {
		o = &V2GetOpts{
			Version: opts.Version,
		}
	}

	var m V2Version
	m, err = k.client.V2Get(mount, subpath, output, o)
	if err == nil {
		meta.Deleted = m.DeletedAt != nil
		meta.Destroyed = m.Destroyed
		meta.Version = m.Version
	}
	return
}

func (k kvv2Mount) List(mount, subpath string) (paths []string, err error) {
	return k.client.V2List(mount, subpath)
}

func (k kvv2Mount) Set(mount, subpath string, values map[string]string, opts *KVSetOpts) (meta KVVersion, err error) {
	var m V2Version
	m, err = k.client.V2Set(mount, subpath, values, nil)
	if err == nil {
		meta.Version = m.Version
	}
	return
}

func (k kvv2Mount) Delete(mount, subpath string, opts *KVDeleteOpts) (err error) {
	versions := []uint{}
	if opts != nil {
		versions = opts.Versions
	}
	return k.client.V2Delete(mount, subpath, &V2DeleteOpts{Versions: versions})
}

func (k kvv2Mount) Undelete(mount, subpath string, versions []uint) (err error) {
	return k.client.V2Undelete(mount, subpath, versions)
}

func (k kvv2Mount) Destroy(mount, subpath string, versions []uint) (err error) {
	return k.client.V2Destroy(mount, subpath, versions)
}

func (k kvv2Mount) DestroyAll(mount, subpath string) (err error) {
	return k.client.V2DestroyMetadata(mount, subpath)
}

func (k kvv2Mount) Versions(mount, subpath string) (ret []KVVersion, err error) {
	var meta V2Metadata
	meta, err = k.client.V2GetMetadata(mount, subpath)
	if err != nil {
		return nil, err
	}

	ret = make([]KVVersion, len(meta.Versions))
	for i := range meta.Versions {
		ret[i].Deleted = meta.Versions[i].DeletedAt != nil
		ret[i].Destroyed = meta.Versions[i].Destroyed
		ret[i].Version = meta.Versions[i].Version
	}
	return
}

func (k kvv2Mount) MountVersion() (version uint) {
	return 2
}

/*==========================
       KV Abstraction
==========================*/

//NewKV returns an initialized KV object.
func (v *Client) NewKV() *KV {
	return &KV{Client: v, mounts: map[string]kvMount{}}
}

func (k *KV) mountForPath(path string) (mountPath string, ret kvMount, err error) {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	var found bool
	k.lock.RLock()
	for i := 1; i <= len(pathParts); i++ {
		mountPath = strings.Join(pathParts[:i], "/")
		ret, found = k.mounts[mountPath]
		if found {
			break
		}
	}
	k.lock.RUnlock()
	if found {
		return
	}

	k.lock.Lock()
	defer k.lock.Unlock()
	for i := 1; i <= len(pathParts); i++ {
		mountPath = strings.Join(pathParts[:i], "/")
		ret, found = k.mounts[mountPath]
		if found {
			break
		}
	}
	if found {
		return
	}

	mountPath, isV2, err := k.Client.IsKVv2Mount(path)
	if err != nil {
		return
	}

	ret = kvv1Mount{k.Client}
	if isV2 {
		ret = kvv2Mount{k.Client}
	}

	k.mounts[mountPath] = ret

	return
}

func subtractMount(mount string, path string) string {
	mount = strings.Trim(mount, "/")
	path = strings.Trim(path, "/")
	var ret string
	if mount != path {
		ret = strings.Trim(strings.TrimPrefix(path, mount), "/")
	}
	return ret
}

//KVGetOpts are options applicable to KV.Get
type KVGetOpts struct {
	// Version is the version of the resource to retrieve. Setting this to zero (or
	// not setting it at all) will retrieve the latest version
	Version uint
}

//KVVersion contains information about a version of a secret.
type KVVersion struct {
	Version   uint
	Deleted   bool
	Destroyed bool
}

//Get retrieves the value at the given path in the tree. This follows the
//semantics of Client.Get or Client.V2Get, chosen based on the backend mounted
//at the path given.
func (k *KV) Get(path string, output interface{}, opts *KVGetOpts) (meta KVVersion, err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.Get(mountPath, path, output, opts)
}

//List retrieves the paths under the given path. If the path does not exist or
//it is not a folder, ErrNotFound is thrown. Results ending with a slash are
//folders.
func (k *KV) List(path string) (paths []string, err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.List(mountPath, path)
}

//KVSetOpts are the options for a set call to the KV.Set() call. Currently there
// are none, but it exists in case the API adds support in the future for things
// that we can put here.
type KVSetOpts struct{}

//Set puts the values given at the path given. If KV v1, the previous value, if
//any, is overwritten.  If KV v2, a new version is created.
func (k *KV) Set(path string, values map[string]string, opts *KVSetOpts) (meta KVVersion, err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.Set(mountPath, path, values, opts)
}

//KVDeleteOpts are options applicable to KV.Delete
type KVDeleteOpts struct {
	//Versions are the versions of the secret to delete. If left nil,
	// the latest version is deleted.
	Versions []uint
	//V1Destroy, if true, will call Client.Delete if the given path
	// to delete is a V1 backend (thus permanently destroying the secret).
	// If it is false and the backend is V1, an ErrKVUnsupported error will
	// be thrown. This has no effect on KV v2 backends.
	V1Destroy bool
}

//Delete attempts to mark the secret at the given path (and version) as deleted.
// For KV v1, temporarily deleting a secret is not possible. Use the V1Destroy
// option as a way to safeguard against unwanted destruction of secrets.
func (k *KV) Delete(path string, opts *KVDeleteOpts) (err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.Delete(mountPath, path, opts)
}

//Undelete attempts to unmark deletion on a previously deleted version.
// KV v1 backends cannot do this, and so if the backend is KV v1, this
// returns an ErrKVUnsupported.
func (k *KV) Undelete(path string, versions []uint) (err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.Undelete(mountPath, path, versions)
}

//Destroy attempts to irrevocably delete the given versions at the given
// path. For KV v1 backends, this is a call to Client.Delete. for KV v2
// backends, this is a call to Client.V2Destroy
func (k *KV) Destroy(path string, versions []uint) (err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.Destroy(mountPath, path, versions)
}

//DestroyAll attempts to irrevocably delete all versions of the secret
// at the given path. For KV v1 backends, this is a call to Client.Delete.
// For v2 backends, this is a call to Client.V2DestroyMetadata
func (k *KV) DestroyAll(path string) (err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.DestroyAll(mountPath, path)
}

//Versions returns the versions of the secret available. If no secret
// exists at this path, ErrNotFound is returned. If the secret exists
// and this is a KV v1 backend, one version is returned.
func (k *KV) Versions(path string) (ret []KVVersion, err error) {
	mountPath, mount, err := k.mountForPath(path)
	if err != nil {
		return
	}

	path = subtractMount(mountPath, path)
	return mount.Versions(mountPath, path)
}

//MountVersion returns the KV version of the mount for the given path.
// v1 mounts return 1; v2 mounts return 2.
func (k *KV) MountVersion(mount string) (version uint, err error) {
	_, m, err := k.mountForPath(mount)
	if err != nil {
		return
	}

	return m.MountVersion(), nil
}

//MountPath returns the path of the mount on which the given path is mounted.
// If no such mount can be found, an error is returned.
func (k *KV) MountPath(path string) (mount string, err error) {
	mount, _, err = k.mountForPath(path)
	return
}
