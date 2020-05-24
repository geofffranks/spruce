package vaultkv_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/vaultkv"
)

var _ = Describe("Generate Root", func() {
	When("the vault is not initialized", func() {
		Describe("Starting a new generate root operation", func() {
			JustBeforeEach(func() {
				_, err = vault.NewGenerateRoot()
			})

			It("should return ErrUninitialized", AssertErrorOfType(&vaultkv.ErrUninitialized{}))
		})
	})

	When("the vault is initialized", func() {
		var initShares, initThreshold int
		var initOutput *vaultkv.InitVaultOutput
		BeforeEach(func() {
			initShares = 1
			initThreshold = 1
		})

		JustBeforeEach(func() {
			initOutput, err = vault.InitVault(vaultkv.InitConfig{
				Shares:    initShares,
				Threshold: initThreshold,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		When("Vault is sealed", func() {
			Describe("Starting a new generate root operation", func() {
				JustBeforeEach(func() {
					_, err = vault.NewGenerateRoot()
				})

				It("should return ErrSealed", AssertErrorOfType(&vaultkv.ErrSealed{}))
			})
		})

		When("Vault is unsealed", func() {
			JustBeforeEach(func() {
				err = initOutput.Unseal()
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("Starting a new generate root operation", func() {
				var genRoot *vaultkv.GenerateRoot

				var AssertRemaining = func(rem int) func() {
					return func() {
						Expect(genRoot.Remaining()).To(Equal(rem))
					}
				}

				JustBeforeEach(func() {
					genRoot, err = vault.NewGenerateRoot()
				})

				Context("With one key in the previous initialization", func() {
					It("should generate a new root token properly", func() {
						By("initializing the generate root operation without erroring")
						Expect(err).NotTo(HaveOccurred())

						By("having remaining report one")
						AssertRemaining(1)()

						By("having State not return nil")
						state := genRoot.State()

						//State with zero keys submitted
						By("having the state say Required is one")
						Expect(state.Required).To(Equal(1))

						By("having the state say Progress is zero")
						Expect(state.Progress).To(Equal(0))

						var genRootDone bool
						genRootDone, err = genRoot.Submit(initOutput.Keys[0])
						By("having the first key submission not err")
						Expect(err).NotTo(HaveOccurred())

						By("having the first key submission finish the generate root operation")
						Expect(genRootDone).To(BeTrue())

						By("having Remaining return zero")
						AssertRemaining(0)()

						var newToken string
						newToken, err = genRoot.RootToken()
						By("having RootToken not err")
						Expect(err).NotTo(HaveOccurred())

						By("having RootToken give back the root token")
						Expect(newToken).NotTo(BeEmpty())

						By("Being able to use the returned token to authenticate")
						vault.AuthToken = newToken
						mountType := vaultkv.MountTypeKV
						if parseSemver(currentVaultVersion).LessThan(semver{0, 8, 0}) {
							mountType = vaultkv.MountTypeGeneric
						}

						err = vault.EnableSecretsMount("beep", vaultkv.Mount{Type: mountType})
						Expect(err).NotTo(HaveOccurred())
					})

					Describe("Submitting too many keys all at once", func() {
						var genRootDone bool
						JustBeforeEach(func() {
							genRootDone, err = genRoot.Submit(initOutput.Keys[0], "a", "b", "c")
						})

						It("should properly generate a new root token (as long as the first keys are correct)", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("saying that the generate root operation is done")
							Expect(genRootDone).To(BeTrue())
						})
					})

					Describe("Submitting an incorrect key", func() {
						var genRootDone bool
						JustBeforeEach(func() {
							//If this is somehow your unseal key, then I'm sorry
							genRootDone, err = genRoot.Submit("k8vk0IdoDeNAJl5JDJ282eehqIbRLv5WWoBy6ppBK9c=")
						})

						It("should err properly", func() {
							By("returning an ErrBadRequest")
							AssertErrorOfType(&vaultkv.ErrBadRequest{})()

							By("saying that it's not done")
							Expect(genRootDone).To(BeFalse())
						})
					})
				})

				Context("With multiple keys in the previous initialization", func() {
					BeforeEach(func() {
						initShares = 3
						initThreshold = 3
					})

					It("should allow new root token generation attempt to be created", func() {
						By("not erroring from the creation of the generate root operation")
						Expect(err).NotTo(HaveOccurred())

						By("having Remaining return three")
						AssertRemaining(3)()

						By("having the first key submission not err")
						var genRootDone bool
						genRootDone, err = genRoot.Submit(initOutput.Keys[0])
						Expect(err).NotTo(HaveOccurred())

						By("not claiming to be done with the generate root operation")
						Expect(genRootDone).To(BeFalse())

						By("having Remaining return two")
						AssertRemaining(2)()

						By("cancelling the generate root operation not returning an error")
						err = genRoot.Cancel()
						Expect(err).NotTo(HaveOccurred())

						By("submitting after the generate root operation was cancelled returning an ErrBadRequest")
						genRootDone, err = genRoot.Submit(initOutput.Keys[0])
						AssertErrorOfType(&vaultkv.ErrBadRequest{})()

						By("the submission after the generate root operation was cancelled returning that the operation is done")
						Expect(genRootDone).To(BeTrue())

					})

					Describe("Submitting all necessary keys", func() {
						var genRootDone bool
						Context("All at once", func() {
							JustBeforeEach(func() {
								genRootDone, err = genRoot.Submit(initOutput.Keys...)
							})

							It("should generate a new root token successfully", func() {
								By("not erroring")
								Expect(err).NotTo(HaveOccurred())

								By("claiming that the generate root operation is done")
								Expect(genRootDone).To(BeTrue())

								By("having Remaining return 0")
							})
						})

						Context("One Submit call at a time", func() {
							var genRootDone bool
							JustBeforeEach(func() {
								for _, key := range initOutput.Keys {
									genRootDone, err = genRoot.Submit(key)
									Expect(err).NotTo(HaveOccurred())
								}
							})

							It("should generate a new root token successfully", func() {
								By("returning that the generate root operation is done")
								Expect(genRootDone).To(BeTrue())

								By("having Remaining return zero")
								AssertRemaining(0)()
							})
						})
					})
				})
			})
		})
	})
})
