package vaultkv_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry-community/vaultkv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sys", func() {

	//Uses SealStatus to get seal state
	var AssertStatusSealed = func(expected bool) func() {
		return func() {
			state, err := vault.SealStatus()
			Expect(err).NotTo(HaveOccurred())
			Expect(state).ToNot(BeNil())
			Expect(state.Sealed).To(Equal(expected))
		}
	}

	Describe("Initialization", func() {
		var output *vaultkv.InitVaultOutput
		var input vaultkv.InitConfig
		JustBeforeEach(func() {
			output, err = vault.InitVault(input)
		})

		var AssertHasRootToken = func() func() {
			return func() {
				Expect(output).ToNot(BeNil())
				Expect(output.RootToken).ToNot(BeEmpty())
			}
		}

		var AssertHasUnsealKeys = func(numKeys int) func() {
			return func() {
				Expect(output).ToNot(BeNil())
				Expect(output.Keys).To(HaveLen(numKeys))
				Expect(output.KeysBase64).To(HaveLen(numKeys))
				for i, key := range output.Keys {
					//Decode the base64
					buf := strings.NewReader(output.KeysBase64[i])
					b64decoder := base64.NewDecoder(base64.StdEncoding, buf)
					b64decoded, err := ioutil.ReadAll(b64decoder)
					Expect(err).NotTo(HaveOccurred(), "should not have erred on decoding base64")
					//Encode into hex
					hexEncoded := hex.EncodeToString(b64decoded)
					Expect(string(hexEncoded)).To(Equal(key),
						fmt.Sprintf("base64 string `%s' does not decode to the same string as the hex string `%s' decodes", output.KeysBase64[i], key))
				}
			}
		}

		var AssertInitializationStatus = func(expected bool) func() {
			return func() {
				actual, err := vault.IsInitialized()
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			}
		}

		When("the Vault is not initialized", func() {
			When("there's only one secret share", func() {
				BeforeEach(func() {
					input = vaultkv.InitConfig{
						Shares:    1,
						Threshold: 1,
					}
				})

				It("should initialize the vault", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("returning a root token")
					AssertHasRootToken()()

					By("returning one unseal key")
					AssertHasUnsealKeys(1)()

					By("having the vault say it is initialized")
					AssertInitializationStatus(true)()
				})

				Describe("Unseal with an InitVaultOutput", func() {
					JustBeforeEach(func() {
						err = output.Unseal()
					})

					It("should unseal the vault properly", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("the vault saying that it is unsealed")
						sealState, err := vault.SealStatus()
						Expect(err).NotTo(HaveOccurred())
						Expect(sealState).NotTo(BeNil())
						Expect(sealState.Sealed).To(BeFalse())
					})
				})
			})

			When("there are multiple secret shares", func() {
				BeforeEach(func() {
					input = vaultkv.InitConfig{
						Shares:    3,
						Threshold: 2,
					}
				})

				It("should initialize the vault", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("returning a root token")
					AssertHasRootToken()()

					By("returning the correct amount of unseal keys")
					AssertHasUnsealKeys(3)()

					By("having the vault say it is initialized")
					AssertInitializationStatus(true)()
				})
			})

			When("0 secret shares are requested", func() {
				BeforeEach(func() {
					input = vaultkv.InitConfig{
						Shares:    0,
						Threshold: 0,
					}

					It("should err properly", func() {
						By("returning ErrBadRequest")
						AssertErrorOfType(&vaultkv.ErrBadRequest{})()

						By("having the vault say that it is not yet initialized")
						AssertInitializationStatus(false)()
					})
				})
			})

			When("the threshold is larger than the number of shares", func() {
				BeforeEach(func() {
					input = vaultkv.InitConfig{
						Shares:    3,
						Threshold: 4,
					}

					It("should err properly", func() {
						By("returning ErrBadRequest")
						AssertErrorOfType(&vaultkv.ErrBadRequest{})()

						By("having the vault say that it is not yet initialized")
						AssertInitializationStatus(false)()
					})
				})
			})
		})

		When("the Vault has already been initialized", func() {
			BeforeEach(func() {
				input = vaultkv.InitConfig{
					Shares:    1,
					Threshold: 1,
				}
				_, err = vault.InitVault(input)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should err properly", func() {
				By("returning ErrBadRequest")
				AssertErrorOfType(&vaultkv.ErrBadRequest{})()

				By("having the vault say that it is still initialized")
				AssertInitializationStatus(true)()
			})
		})
	})

	Describe("Unseal", func() {
		var output *vaultkv.SealState
		var unsealKey string

		BeforeEach(func() {
			unsealKey = "pLacEhoLdeR="
		})
		JustBeforeEach(func() {
			output, err = vault.Unseal(unsealKey)
		})

		var AssertSealed = func(expected bool) func() {
			return func() {
				Expect(output).ToNot(BeNil())
				Expect(output.Sealed).To(Equal(expected))
			}
		}

		var AssertProgressIs = func(expected int) func() {
			return func() {
				state, err := vault.SealStatus()
				Expect(err).NotTo(HaveOccurred())
				Expect(state.Progress).To(Equal(expected))
			}
		}

		When("the vault is initialized", func() {
			var initOut *vaultkv.InitVaultOutput
			Context("with one share", func() {
				BeforeEach(func() {
					initOut, err = vault.InitVault(vaultkv.InitConfig{
						Shares:    1,
						Threshold: 1,
					})
				})

				When("unseal key is correct", func() {
					BeforeEach(func() {
						unsealKey = initOut.Keys[0]
					})

					It("should unseal the vault", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("returning that the unseal is finished")
						AssertSealed(false)()

						By("having the vault say that it is now unsealed")
						AssertStatusSealed(false)()
					})

					When("an unseal is attempted after the vault is unsealed", func() {
						BeforeEach(func() {
							_, err = vault.Unseal(unsealKey)
							Expect(err).NotTo(HaveOccurred())
						})

						It("should idempotently state that the vault is unsealed", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("returning that the unseal is finished")
							AssertSealed(false)()

							By("having the vault say that it is now unsealed")
							AssertStatusSealed(false)()
						})
					})

					When("an unseal reset is requested after the vault is unsealed", func() {
						JustBeforeEach(func() {
							err = vault.ResetUnseal()
						})

						It("should err properly", func() {
							By("returning an ErrBadRequest")
							AssertErrorOfType(&vaultkv.ErrBadRequest{})()
						})
					})
				})

				When("the unseal key is wrong", func() {
					BeforeEach(func() {
						// make the unseal key wrong
						unsealKey = initOut.Keys[0]
						replacementChar := "a"
						if unsealKey[0] == 'a' {
							replacementChar = "b"
						}

						unsealKey = fmt.Sprintf("%s%s", replacementChar, unsealKey[1:])
					})

					It("should err properly", func() {
						By("returning an ErrBadRequest")
						AssertErrorOfType(&vaultkv.ErrBadRequest{})()

						By("having the vault say that it is still sealed")
						AssertStatusSealed(true)()
					})
				})

				When("the unseal key is improperly formatted", func() {
					BeforeEach(func() {
						unsealKey = "bettergocatchitlol.png"
					})

					It("should err properly", func() {
						By("returning an ErrBadRequest")
						AssertErrorOfType(&vaultkv.ErrBadRequest{})()

						By("having the vault say that it is still sealed")
						AssertStatusSealed(true)()
					})
				})
			})

			Context("with a threshold greater than one", func() {
				var testShares, testThreshold int
				BeforeEach(func() {
					testShares = 3
					testThreshold = 3
					initOut, err = vault.InitVault(vaultkv.InitConfig{
						Shares:    testShares,
						Threshold: testThreshold,
					})
				})

				When("the unseal key is improperly formatted", func() {
					It("should return an ErrBadRequest", AssertErrorOfType(&vaultkv.ErrBadRequest{}))
					It("should not have increased the progress count", AssertProgressIs(0))
					Specify("SealStatus should return that the Vault is still sealed", AssertStatusSealed(true))
				})

				When("the unseal keys are correct", func() {
					BeforeEach(func() {
						unsealKey = initOut.Keys[0]
					})

					It("should unseal the vault", func() {
						By("not returning an error after the first key is given")
						Expect(err).NotTo(HaveOccurred())

						By("increasing the progress count to one")
						AssertProgressIs(1)()

						/*
							Test that values are getting populated, but it should be redundant
							after the first key
						*/
						By("returning that the progress is 1")
						Expect(output.Progress).To(Equal(1))

						By("returning the correct threshold required")
						Expect(output.Threshold).To(Equal(testThreshold))

						By("returning the correct number of shares")
						Expect(output.NumShares).To(Equal(testShares))

						By("having a nonce")
						Expect(output.Nonce).ToNot(BeEmpty())

						By("returning the correct version")
						Expect(output.Version).To(Equal(currentVaultVersion))

						By("returning that the Vault is still sealed")
						AssertSealed(true)()

						By("the vault saying that the vault is still sealed")
						AssertStatusSealed(true)()

						By("not returning an error after the second key is given")
						output, err = vault.Unseal(initOut.Keys[1])
						Expect(err).NotTo(HaveOccurred())

						By("increasing the progress count to two")
						AssertProgressIs(2)()

						By("returning that the Vault is still sealed")
						AssertSealed(true)()

						By("the vault saying that the vault is still sealed")
						AssertStatusSealed(true)()

						By("not returning an error after the final key is given")
						output, err = vault.Unseal(initOut.Keys[2])
						Expect(err).NotTo(HaveOccurred())

						By("returning that the Vault is now unsealed")
						AssertSealed(false)()

						By("the vault saying that the vault is now unsealed")
						AssertStatusSealed(false)()
					})

					Describe("ResetUnseal", func() {
						JustBeforeEach(func() {
							output, err = vault.Unseal(initOut.Keys[0])
							Expect(err).NotTo(HaveOccurred())
							AssertProgressIs(1)()
							err = vault.ResetUnseal()
						})

						It("should reset the current unseal attempt", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("having SealState claim that the progress is 0")
							AssertProgressIs(0)()
						})
					})
				})
			})
		})
	})

	Describe("Seal", func() {
		JustBeforeEach(func() {
			err = vault.Seal()
		})

		When("the vault is not initialized", func() {
			It("should not return an error", func() { Expect(err).NotTo(HaveOccurred()) })
		})

		When("the vault is initialized", func() {
			var initOut *vaultkv.InitVaultOutput
			BeforeEach(func() {
				initOut, err = vault.InitVault(vaultkv.InitConfig{
					Shares:    1,
					Threshold: 1,
				})
			})
			When("the vault is already sealed", func() {

				It("should idempotently return that the vault is sealed", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("the vault claiming that it is still sealed")
					AssertStatusSealed(true)()
				})
			})

			When("the vault is unsealed", func() {
				BeforeEach(func() {
					sealState, err := vault.Unseal(initOut.Keys[0])
					Expect(err).NotTo(HaveOccurred())
					Expect(sealState).NotTo(BeNil())
					Expect(sealState.Sealed).To(BeFalse())
				})

				It("should seal the vault", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("the vault saying that it is now sealed")
					AssertStatusSealed(true)()
				})
			})
		})
	})

	Describe("Mount", func() {
		var testBackendName string
		var testBackendConfig vaultkv.Mount

		JustBeforeEach(func() {
			InitAndUnsealVault()
			err = vault.EnableSecretsMount(
				testBackendName,
				testBackendConfig,
			)
		})

		BeforeEach(func() {
			testBackendName = "beepboop"
			testBackendConfig = vaultkv.Mount{
				Description: "a test mount",
			}
		})

		Describe("Mounting a non-existent backend type", func() {
			BeforeEach(func() {
				testBackendConfig.Type = "dcgeduceohdursaoceh"
			})

			It("should return ErrBadRequest", AssertErrorOfType(&vaultkv.ErrBadRequest{}))
		})

		Describe("Mounting a KVv1 backend", func() {
			BeforeEach(func() {
				testBackendConfig.Type = vaultkv.MountTypeKV
				if parseSemver(currentVaultVersion).LessThan(semver{0, 8, 0}) {
					testBackendConfig.Type = vaultkv.MountTypeGeneric
				}
			})

			It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })

			Describe("Listing the backends", func() {
				var backendList map[string]vaultkv.Mount
				JustBeforeEach(func() {
					backendList, err = vault.ListMounts()
				})

				It("should show the new backend in the list", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("having a backend with the correct name")
					backend, ok := backendList[testBackendName]
					Expect(ok).To(BeTrue())

					By("having that backend display the correct type")
					Expect(backend.Type).To(Equal(testBackendConfig.Type))

					By("having that backend display the correct description")
					Expect(backend.Description).To(Equal(testBackendConfig.Description))
				})
			})

			Describe("Unmounting the backend", func() {
				JustBeforeEach(func() {
					err = vault.DisableSecretsMount(testBackendName)
				})
				It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })

				Describe("Listing the backends", func() {
					var backendList map[string]vaultkv.Mount
					JustBeforeEach(func() {
						backendList, err = vault.ListMounts()
					})

					It("should provide a list that has the backend gone", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("having the mount not be present")
						_, ok := backendList[testBackendName]
						Expect(ok).To(BeFalse())
					})
				})
			})

			Describe("Unmounting a backend that doesn't exist", func() {
				JustBeforeEach(func() {
					err = vault.DisableSecretsMount("hsaetdieogudsoearu")
				})

				It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })
			})
		})

		Describe("Mounting a KVv2 backend", func() {
			BeforeEach(func() {
				if parseSemver(currentVaultVersion).LessThan(semver{0, 10, 0}) {
					Skip("KV version 2 did not exist before 0.10.0")
				}

				testBackendConfig = vaultkv.Mount{
					Type:        vaultkv.MountTypeKV,
					Description: "A test v2 backend",
					Options:     vaultkv.KVMountOptions{}.WithVersion(2),
				}
			})

			It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })

			Describe("Listing the backends", func() {
				var backendList map[string]vaultkv.Mount
				JustBeforeEach(func() {
					backendList, err = vault.ListMounts()
				})

				It("should have a backend which is properly configured", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())
					By("being in the returned list")
					backend, ok := backendList[testBackendName]
					Expect(ok).To(BeTrue(), "Expected the created backend to appear in the mount list")

					By("having a options entry")
					Expect(backend.Options).NotTo(BeNil())

					By("being version 2")
					Expect(vaultkv.KVMountOptions(backend.Options).GetVersion()).To(Equal(2))
				})
			})
		})
	})

	Describe("Health", func() {
		JustBeforeEach(func() {
			err = vault.Health(true)
		})

		When("the vault is initialized", func() {
			var initOut *vaultkv.InitVaultOutput
			BeforeEach(func() {
				initOut, err = vault.InitVault(vaultkv.InitConfig{
					Shares:    1,
					Threshold: 1,
				})
			})

			When("the vault is unsealed", func() {
				BeforeEach(func() {
					sealState, err := vault.Unseal(initOut.Keys[0])
					Expect(err).NotTo(HaveOccurred())
					Expect(sealState).NotTo(BeNil())
					Expect(sealState.Sealed).To(BeFalse())
				})

				It("should not return an error", func() { Expect(err).NotTo(HaveOccurred()) })

				When("the auth token is wrong", func() {
					BeforeEach(func() {
						vault.AuthToken = ""
					})

					It("should not return an error", func() { Expect(err).NotTo(HaveOccurred()) })
				})
			})
		})
	})
})
