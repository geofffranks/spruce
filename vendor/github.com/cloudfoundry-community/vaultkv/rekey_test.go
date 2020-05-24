package vaultkv_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/vaultkv"
)

var _ = Describe("Rekey", func() {
	When("the vault is not initialized", func() {
		Describe("Starting a new rekey operation", func() {
			JustBeforeEach(func() {
				_, err = vault.NewRekey(vaultkv.RekeyConfig{
					Shares:    1,
					Threshold: 1,
				})
			})

			It("should return ErrUninitialized", AssertErrorOfType(&vaultkv.ErrUninitialized{}))
		})

		Describe("Getting the current rekey operation", func() {
			JustBeforeEach(func() {
				_, err = vault.CurrentRekey()
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
			Describe("Starting a new rekey operation", func() {
				JustBeforeEach(func() {
					_, err = vault.NewRekey(vaultkv.RekeyConfig{
						Shares:    1,
						Threshold: 1,
					})
				})

				It("should return ErrSealed", AssertErrorOfType(&vaultkv.ErrSealed{}))
			})

			Describe("Getting the current rekey operation", func() {
				JustBeforeEach(func() {
					_, err = vault.CurrentRekey()
				})

				It("should return ErrSealed", AssertErrorOfType(&vaultkv.ErrSealed{}))
			})
		})

		When("Vault is unsealed", func() {
			JustBeforeEach(func() {
				err = initOutput.Unseal()
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("CurrentRekey with no rekey in progress", func() {
				JustBeforeEach(func() {
					_, err = vault.CurrentRekey()
				})

				It("should return ErrNotFound", AssertErrorOfType(&vaultkv.ErrNotFound{}))
			})

			Describe("Starting a new rekey operation", func() {
				var rekeyConf vaultkv.RekeyConfig
				var rekey *vaultkv.Rekey

				var AssertRemaining = func(rem int) func() {
					return func() {
						Expect(rekey.Remaining()).To(Equal(rem))
					}
				}

				var AssertHasKeys = func(numKeys int) func() {
					return func() {
						Expect(rekey.Keys()).To(HaveLen(numKeys))
					}
				}

				JustBeforeEach(func() {
					rekey, err = vault.NewRekey(rekeyConf)
				})

				Context("With one key in the previous initialization", func() {
					Context("With one share and threshold of one requested", func() {
						BeforeEach(func() {
							rekeyConf.Shares = 1
							rekeyConf.Threshold = 1
						})

						It("should rekey properly", func() {
							By("initializing the rekey without erroring")
							Expect(err).NotTo(HaveOccurred())

							By("having remaining report one")
							AssertRemaining(1)()

							By("having State not return nil")
							state := rekey.State()

							//State with zero keys submitted
							By("having the state say PendingShares is one")
							Expect(state.PendingShares).To(Equal(1))

							By("having the state say PendingThreshold is one")
							Expect(state.PendingThreshold).To(Equal(1))

							By("having the state say Required is one")
							Expect(state.Required).To(Equal(1))

							By("having the state say Progress is zero")
							Expect(state.Progress).To(Equal(0))

							var rekeyDone bool
							rekeyDone, err = rekey.Submit(initOutput.Keys[0])
							By("having the first key submission not err")
							Expect(err).NotTo(HaveOccurred())

							By("having the first key submission finish the rekey")
							Expect(rekeyDone).To(BeTrue())

							By("having Keys have one new key")
							Expect(rekey.Keys()).To(HaveLen(1))

							By("having Remaining return zero")
							AssertRemaining(0)()
						})

						Describe("Submitting too many keys all at once", func() {
							var rekeyDone bool
							JustBeforeEach(func() {
								rekeyDone, err = rekey.Submit(initOutput.Keys[0], "a", "b", "c")
							})

							It("should properly unseal the vault (as long as the first keys are correct)", func() {
								By("not erroring")
								Expect(err).NotTo(HaveOccurred())

								By("saying that the rekey is done")
								Expect(rekeyDone).To(BeTrue())
							})
						})

						Describe("Submitting an incorrect key", func() {
							var rekeyDone bool
							JustBeforeEach(func() {
								//If this is somehow your unseal key, then I'm sorry
								rekeyDone, err = rekey.Submit("k8vk0IdoDeNAJl5JDJ282eehqIbRLv5WWoBy6ppBK9c=")
							})

							It("should err properly", func() {
								By("returning an ErrBadRequest")
								AssertErrorOfType(&vaultkv.ErrBadRequest{})()

								By("saying that it's not done")
								Expect(rekeyDone).To(BeFalse())
							})
						})
					})

					Context("with improper rekey parameters", func() {
						BeforeEach(func() {
							rekeyConf.Shares = 1
							rekeyConf.Threshold = 2
						})
						It("should return ErrBadRequest", AssertErrorOfType(&vaultkv.ErrBadRequest{}))
					})

				})

				Context("With multiple keys in the previous initialization", func() {
					BeforeEach(func() {
						initShares = 3
						initThreshold = 3
					})

					Context("With one share and threshold of one requested", func() {
						BeforeEach(func() {
							rekeyConf.Shares = 1
							rekeyConf.Threshold = 1
						})

						It("should allow rekey operations", func() {
							By("not erroring from the creation of the rekey")
							Expect(err).NotTo(HaveOccurred())

							By("having Remaining return three")
							AssertRemaining(3)()

							By("having the first key submission not err")
							var rekeyDone bool
							rekeyDone, err = rekey.Submit(initOutput.Keys[0])
							Expect(err).NotTo(HaveOccurred())

							By("not claiming to be done with the rekey")
							Expect(rekeyDone).To(BeFalse())

							By("having Remaining return two")
							AssertRemaining(2)()

							By("getting the current rekey operation not erroring")
							rekey, err = vault.CurrentRekey()
							Expect(err).NotTo(HaveOccurred())

							By("the CurrentRekey operation not returning nil")
							Expect(rekey).NotTo(BeNil())

							By("the CurrentRekey return value's Remaining should return two")
							AssertRemaining(2)

							By("cancelling the rekey not returning an error")
							err = rekey.Cancel()
							Expect(err).NotTo(HaveOccurred())

							By("submitting after the rekey was cancelled returning an ErrBadRequest")
							rekeyDone, err = rekey.Submit(initOutput.Keys[0])
							AssertErrorOfType(&vaultkv.ErrBadRequest{})()

							By("the submission after the rekey was cancelled returning that the rekey is done")
							Expect(rekeyDone).To(BeTrue())

						})

						Describe("Submitting all necessary keys", func() {
							var rekeyDone bool
							Context("All at once", func() {
								JustBeforeEach(func() {
									rekeyDone, err = rekey.Submit(initOutput.Keys...)
								})

								It("should rekey the vault successfully", func() {
									By("not erroring")
									Expect(err).NotTo(HaveOccurred())

									By("claiming that the rekey is done")
									Expect(rekeyDone).To(BeTrue())

									By("having Remaining return 0")

									By("Keys returning the 1 new key")
									AssertHasKeys(1)()
								})
							})

							Context("One Submit call at a time", func() {
								var rekeyDone bool
								JustBeforeEach(func() {
									for _, key := range initOutput.Keys {
										rekeyDone, err = rekey.Submit(key)
										Expect(err).NotTo(HaveOccurred())
									}
								})

								It("should rekey successfully", func() {
									By("returning that the rekey is done")
									Expect(rekeyDone).To(BeTrue())

									By("having Remaining return zero")
									AssertRemaining(0)()

									By("having Keys return the one new key")
									AssertHasKeys(1)()
								})
							})
						})
					})
				})
			})
		})
	})
})
