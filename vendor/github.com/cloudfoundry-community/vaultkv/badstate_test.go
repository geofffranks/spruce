package vaultkv_test

import (
	"fmt"

	"github.com/cloudfoundry-community/vaultkv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = When("the vault is uninitialized", func() {
	type spec struct {
		Name       string
		Setup      func()
		MinVersion *semver
	}
	Specify("Most commands should return ErrUninitialized", func() {
		for _, s := range []spec{
			spec{"Health", func() { err = vault.Health(true) }, nil},
			spec{"EnableSecretsMount", func() { err = vault.EnableSecretsMount("beep", vaultkv.Mount{}) }, nil},
			spec{"Unseal", func() { _, err = vault.Unseal("pLacEhoLdeR=") }, nil},
			spec{"Get", func() { err = vault.Get("secret/sure/whatever", nil) }, nil},
			spec{"Set", func() { err = vault.Set("secret/sure/whatever", map[string]string{"foo": "bar"}) }, nil},
			spec{"Delete", func() { err = vault.Delete("secret/sure/whatever") }, nil},
			spec{"List", func() { _, err = vault.List("secret/sure/whatever") }, nil},
			spec{"V2Get", func() { _, err = vault.V2Get("secret", "foo", nil, nil) }, &semver{0, 10, 0}},
			spec{"V2Set", func() { _, err = vault.V2Set("secret", "foo", map[string]string{"beep": "boop"}, nil) }, &semver{0, 10, 0}},
			spec{"V2Delete", func() { err = vault.V2Delete("secret", "foo", nil) }, &semver{0, 10, 0}},
			spec{"V2Undelete", func() { err = vault.V2Undelete("secret", "foo", []uint{1}) }, &semver{0, 10, 0}},
			spec{"V2Destroy", func() { err = vault.V2Destroy("secret", "foo", []uint{1}) }, &semver{0, 10, 0}},
			spec{"V2DestroyMetadata", func() { err = vault.V2DestroyMetadata("secret", "foo") }, &semver{0, 10, 0}},
			spec{"V2GetMetadata", func() { _, err = vault.V2GetMetadata("secret", "foo") }, &semver{0, 10, 0}},
			spec{"KVGet", func() { _, err = vault.NewKV().Get("secret/foo", nil, nil) }, &semver{0, 10, 0}},
			spec{"KVSet", func() { _, err = vault.NewKV().Set("secret/foo", map[string]string{"beep": "boop"}, nil) }, &semver{0, 10, 10}},
			spec{"KVDelete", func() { err = vault.NewKV().Delete("secret/foo", nil) }, &semver{0, 10, 0}},
			spec{"KVUndelete", func() { err = vault.NewKV().Undelete("secret/foo", []uint{1}) }, &semver{0, 10, 0}},
			spec{"KVDestroy", func() { err = vault.NewKV().Destroy("secret/foo", []uint{1}) }, &semver{0, 10, 0}},
			spec{"KVDestroyAll", func() { err = vault.NewKV().DestroyAll("secret/foo") }, &semver{0, 10, 0}},
		} {
			if s.MinVersion != nil && parseSemver(currentVaultVersion).LessThan(*s.MinVersion) {
				continue
			}
			(s.Setup)()
			Expect(err).To(HaveOccurred(),
				fmt.Sprintf("`%s' did not produce an error", s.Name))
			Expect(vaultkv.IsUninitialized(err)).To(BeTrue())
		}
	})
})

var _ = When("the vault is initialized", func() {
	type spec struct {
		Name       string
		Setup      func()
		MinVersion *semver
	}

	var initOut *vaultkv.InitVaultOutput

	BeforeEach(func() {
		initOut, err = vault.InitVault(vaultkv.InitConfig{
			Shares:    1,
			Threshold: 1,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	When("the vault is sealed", func() {
		Specify("Most commands should return ErrSealed", func() {
			for _, s := range []spec{
				spec{"Health", func() { err = vault.Health(true) }, nil},
				spec{"EnableSecretsMount", func() { err = vault.EnableSecretsMount("beep", vaultkv.Mount{}) }, nil},
				spec{"DisableSecretsMount", func() { err = vault.DisableSecretsMount("beep") }, nil},
				spec{"Get", func() { err = vault.Get("secret/sure/whatever", nil) }, nil},
				spec{"Set", func() { err = vault.Set("secret/sure/whatever", map[string]string{"foo": "bar"}) }, nil},
				spec{"Delete", func() { err = vault.Delete("secret/sure/whatever") }, nil},
				spec{"List", func() { _, err = vault.List("secret/sure/whatever") }, nil},
				spec{"V2Get", func() { _, err = vault.V2Get("secret", "foo", nil, nil) }, &semver{0, 10, 0}},
				spec{"V2Set", func() { _, err = vault.V2Set("secret", "foo", map[string]string{"beep": "boop"}, nil) }, &semver{0, 10, 0}},
				spec{"V2Delete", func() { err = vault.V2Delete("secret", "foo", nil) }, &semver{0, 10, 0}},
				spec{"V2Undelete", func() { err = vault.V2Undelete("secret", "foo", []uint{1}) }, &semver{0, 10, 0}},
				spec{"V2Destroy", func() { err = vault.V2Destroy("secret", "foo", []uint{1}) }, &semver{0, 10, 0}},
				spec{"V2DestroyMetadata", func() { err = vault.V2DestroyMetadata("secret", "foo") }, &semver{0, 10, 0}},
				spec{"V2GetMetadata", func() { _, err = vault.V2GetMetadata("secret", "foo") }, &semver{0, 10, 0}},
				spec{"KVGet", func() { _, err = vault.NewKV().Get("secret/foo", nil, nil) }, &semver{0, 10, 0}},
				spec{"KVSet", func() { _, err = vault.NewKV().Set("secret/foo", map[string]string{"beep": "boop"}, nil) }, &semver{0, 10, 10}},
				spec{"KVDelete", func() { err = vault.NewKV().Delete("secret/foo", nil) }, &semver{0, 10, 0}},
				spec{"KVUndelete", func() { err = vault.NewKV().Undelete("secret/foo", []uint{1}) }, &semver{0, 10, 0}},
				spec{"KVDestroy", func() { err = vault.NewKV().Destroy("secret/foo", []uint{1}) }, &semver{0, 10, 0}},
				spec{"KVDestroyAll", func() { err = vault.NewKV().DestroyAll("secret/foo") }, &semver{0, 10, 0}},
			} {
				if s.MinVersion != nil && parseSemver(currentVaultVersion).LessThan(*s.MinVersion) {
					continue
				}
				(s.Setup)()
				Expect(err).To(HaveOccurred(),
					fmt.Sprintf("`%s' did not produce an error", s.Name))
				Expect(vaultkv.IsSealed(err)).To(BeTrue(),
					fmt.Sprintf("`%s' did not make error of type *ErrSealed", s.Name))
			}
		})

		When("the vault is unsealed", func() {
			BeforeEach(func() {
				sealState, err := vault.Unseal(initOut.Keys[0])
				Expect(err).NotTo(HaveOccurred())
				Expect(sealState).NotTo(BeNil())
				Expect(sealState.Sealed).To(BeFalse())
			})
			Specify("KV commands targeted at non-existent things should 404", func() {
				for _, s := range []spec{
					spec{"Get", func() { err = vault.Get("secret/sure/whatever", nil) }, nil},
					spec{"List", func() { _, err = vault.List("secret/sure/whatever") }, nil},
					spec{"V2Get", func() { _, err = vault.V2Get("secret", "foo", nil, nil) }, &semver{0, 10, 0}},
					spec{"V2GetMetadata", func() { _, err = vault.V2GetMetadata("secret", "foo") }, &semver{0, 10, 0}},
				} {
					if s.MinVersion != nil && parseSemver(currentVaultVersion).LessThan(*s.MinVersion) {
						continue
					}
					(s.Setup)()
					Expect(err).To(HaveOccurred(),
						fmt.Sprintf("`%s' did not produce an error", s.Name))
					Expect(vaultkv.IsNotFound(err)).To(BeTrue(),
						fmt.Sprintf("`%s' did not make error of type *ErrNotFound", s.Name))
				}
			})

			When("the auth token is wrong", func() {
				BeforeEach(func() {
					//If this is your token, I'm sorry
					vault.AuthToken = "01234567-89ab-cdef-0123-456789abcdef"
				})
				Specify("Most commands should give a 403", func() {
					for _, s := range []spec{
						spec{"EnableSecretsMount", func() { err = vault.EnableSecretsMount("beep", vaultkv.Mount{}) }, nil},
						spec{"DisableSecretsMount", func() { err = vault.DisableSecretsMount("beep") }, nil},
						spec{"Get", func() { err = vault.Get("secret/sure/whatever", nil) }, nil},
						spec{"Set", func() { err = vault.Set("secret/sure/whatever", map[string]string{"foo": "bar"}) }, nil},
						spec{"Delete", func() { err = vault.Delete("secret/sure/whatever") }, nil},
						spec{"List", func() { _, err = vault.List("secret/sure/whatever") }, nil},
						spec{"V2Get", func() { _, err = vault.V2Get("secret", "foo", nil, nil) }, &semver{0, 10, 0}},
						spec{"V2Set", func() { _, err = vault.V2Set("secret", "foo", map[string]string{"beep": "boop"}, nil) }, &semver{0, 10, 0}},
						spec{"V2Delete", func() { err = vault.V2Delete("secret", "foo", nil) }, &semver{0, 10, 0}},
						spec{"V2Undelete", func() { err = vault.V2Undelete("secret", "foo", []uint{1}) }, &semver{0, 10, 0}},
						spec{"V2Destroy", func() { err = vault.V2Destroy("secret", "foo", []uint{1}) }, &semver{0, 10, 0}},
						spec{"V2DestroyMetadata", func() { err = vault.V2DestroyMetadata("secret", "foo") }, &semver{0, 10, 0}},
						spec{"V2GetMetadata", func() { _, err = vault.V2GetMetadata("secret", "foo") }, &semver{0, 10, 0}},
						spec{"KVGet", func() { _, err = vault.NewKV().Get("secret/foo", nil, nil) }, &semver{0, 10, 0}},
						spec{"KVSet", func() { _, err = vault.NewKV().Set("secret/foo", map[string]string{"beep": "boop"}, nil) }, &semver{0, 10, 10}},
						spec{"KVDelete", func() { err = vault.NewKV().Delete("secret/foo", &vaultkv.KVDeleteOpts{V1Destroy: true}) }, &semver{0, 10, 0}},
						spec{"KVUndelete", func() { err = vault.NewKV().Undelete("secret/foo", []uint{1}) }, &semver{0, 10, 0}},
						spec{"KVDestroy", func() { err = vault.NewKV().Destroy("secret/foo", []uint{1}) }, &semver{0, 10, 0}},
						spec{"KVDestroyAll", func() { err = vault.NewKV().DestroyAll("secret/foo") }, &semver{0, 10, 0}},
					} {
						(s.Setup)()
						Expect(err).To(HaveOccurred(),
							fmt.Sprintf("`%s' did not produce an error", s.Name))
						Expect(vaultkv.IsForbidden(err)).To(BeTrue(),
							fmt.Sprintf("`%s' did not give a 403 - gave %s", s.Name, err))
					}
				})
			})
		})
	})
})
