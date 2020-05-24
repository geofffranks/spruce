package vaultkv_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/vaultkv"
)

var _ = Describe("KV", func() {
	const testMountName = "zip/zop/zoobity/bop"
	var testkv *vaultkv.KV
	BeforeEach(func() {
		InitAndUnsealVault()
		testkv = vault.NewKV()
		Expect(testkv).NotTo(BeNil())
	})

	unityTests := func() {
		Describe("MountPath", func() {
			var mountOutput string
			JustBeforeEach(func() {
				mountOutput, err = testkv.MountPath(fmt.Sprintf("%s/boop", testMountName))
			})

			It("should return the proper mount name", func() {
				By("not returning an error")
				Expect(err).NotTo(HaveOccurred())

				By("having the returned mount name be the same as the created mount's name")
				Expect(mountOutput).To(BeEquivalentTo(testMountName))
			})
		})

		Describe("Set", func() {
			var testSetPath string
			var testSetValues map[string]string
			var testSetOptions *vaultkv.KVSetOpts
			var testVersionOutput vaultkv.KVVersion
			BeforeEach(func() {
				testSetPath = fmt.Sprintf("%s/boop", testMountName)
			})

			JustBeforeEach(func() {
				testVersionOutput, err = testkv.Set(testSetPath, testSetValues, testSetOptions)
			})

			AfterEach(func() {
				testSetValues = nil
				testSetOptions = nil
				testVersionOutput = vaultkv.KVVersion{}
			})

			Context("With a non-empty map input", func() {
				BeforeEach(func() {
					testSetValues = map[string]string{"foo": "bar"}
				})

				It("should write the proper values to the key", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("returning the proper version output")
					Expect(testVersionOutput.Version).To(BeEquivalentTo(uint(1)))
				})

				Describe("List", func() {
					var testListPath string
					var testListOutput []string
					JustBeforeEach(func() {
						testListOutput, err = testkv.List(testListPath)
					})

					When("the path exists", func() {
						BeforeEach(func() {
							_, err = testkv.Set(fmt.Sprintf("%s/foo/bar", testMountName), testSetValues, nil)
							Expect(err).NotTo(HaveOccurred())
							testListPath = testMountName
						})

						It("should list the keys", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("returning the expected keys")
							Expect(testListOutput).To(Equal([]string{"boop", "foo/"}))
						})
					})

					When("the path does not exist", func() {
						BeforeEach(func() {
							testListPath = fmt.Sprintf("%s/this/shouldnt/exist", testMountName)
						})

						It("should return ErrNotFound", AssertErrorOfType(&vaultkv.ErrNotFound{}))
					})
				})

				Describe("Get", func() {
					var testGetOutput map[string]string
					var testGetVersionOutput vaultkv.KVVersion
					JustBeforeEach(func() {
						testGetOutput = map[string]string{}
						testGetVersionOutput, err = testkv.Get(testSetPath, &testGetOutput, nil)
					})

					It("should get the key", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("returning the same version info as the Set")
						Expect(testGetVersionOutput).To(Equal(testVersionOutput))

						By("returning the same values that were set")
						Expect(testGetOutput).To(Equal(testSetValues))
					})
				})

				Describe("Delete", func() {
					var testDeleteVersions []uint
					JustBeforeEach(func() {
						err = testkv.Delete(testSetPath, &vaultkv.KVDeleteOpts{
							Versions:  testDeleteVersions,
							V1Destroy: true,
						})
					})
					AfterEach(func() {
						testDeleteVersions = nil
					})
					Context("Not specifying a version to delete", func() {
						It("should delete the only (newest) version", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("Get being unable to find it")
							_, err = testkv.Get(testSetPath, nil, nil)
							AssertErrorOfType(&vaultkv.ErrNotFound{})()
						})
					})

					Context("Specifying a version to delete", func() {
						When("the version exists", func() {
							BeforeEach(func() {
								testDeleteVersions = []uint{1}
							})

							It("should delete the specified version", func() {
								By("not erroring")
								Expect(err).NotTo(HaveOccurred())

								By("Get being unable to find it")
								_, err = testkv.Get(testSetPath, nil, nil)
								AssertErrorOfType(&vaultkv.ErrNotFound{})()
							})

							Context("and then deleting it again", func() {
								JustBeforeEach(func() {
									err = testkv.Delete(testSetPath, &vaultkv.KVDeleteOpts{
										Versions:  testDeleteVersions,
										V1Destroy: true,
									})
								})

								It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })
							})
						})

						When("the version does not exist", func() {
							BeforeEach(func() {
								testDeleteVersions = []uint{12}
							})

							It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })
						})
					})
				})

				Describe("Destroy", func() {
					When("the version exists and it is the only version", func() {
						JustBeforeEach(func() {
							err = testkv.Destroy(testSetPath, []uint{1})
						})

						It("should delete the metadata", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("Get being unable to find the key")
							_, err = testkv.Get(testSetPath, nil, nil)
							AssertErrorOfType(&vaultkv.ErrNotFound{})

							By("Versions being unable to find the key")
							_, err = testkv.Versions(testSetPath)
							AssertErrorOfType(&vaultkv.ErrNotFound{})
						})
					})

					When("the version does not exist", func() {
						JustBeforeEach(func() {
							err = testkv.Destroy(testSetPath, []uint{12})
						})

						It("should not delete anything", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())

							By("Get being able to find the key")
							_, err = testkv.Get(testSetPath, nil, nil)
							Expect(err).NotTo(HaveOccurred())

							By("Versions being able to find the key")
							var meta []vaultkv.KVVersion
							meta, err = testkv.Versions(testSetPath)
							Expect(err).NotTo(HaveOccurred())

							By("Versions reporting that version 1 still exists")
							Expect(meta).To(HaveLen(1))
						})
					})

					When("the path does not exist", func() {
						JustBeforeEach(func() {
							err = testkv.Destroy(testSetPath+"abcd", []uint{12})
						})

						It("should not err", func() { Expect(err).NotTo(HaveOccurred()) })
					})
				})

				Describe("DestroyAll", func() {
					JustBeforeEach(func() {
						err = testkv.DestroyAll(testSetPath)
					})

					It("should delete the metadata", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("Get being unable to find the key")
						_, err = testkv.Get(testSetPath, nil, nil)
						AssertErrorOfType(&vaultkv.ErrNotFound{})

						By("Versions being unable to find the key")
						_, err = testkv.Versions(testSetPath)
						AssertErrorOfType(&vaultkv.ErrNotFound{})
					})
				})
			})
		})
	}

	Context("With a KV v1 mount", func() {
		BeforeEach(func() {
			mountType := vaultkv.MountTypeKV
			if parseSemver(currentVaultVersion).LessThan(semver{0, 8, 0}) {
				mountType = vaultkv.MountTypeGeneric
			}
			err = vault.EnableSecretsMount(testMountName, vaultkv.Mount{
				Type:    mountType,
				Options: vaultkv.KVMountOptions{}.WithVersion(1),
			})

			Expect(err).NotTo(HaveOccurred())
		})

		unityTests()

		//There are some things that we cannot make exactly the same between kv v1 and v2. We test those things here.
		Describe("v1 specific", func() {
			Describe("isKVv2Mount", func() {
				var mountName string
				var isV2 bool
				JustBeforeEach(func() {
					mountName, isV2, err = vault.IsKVv2Mount(testMountName)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return the mount name and that it is not a v2 mount", func() {
					Expect(mountName).To(BeEquivalentTo(testMountName))
					Expect(isV2).To(BeFalse())
				})
			})

			Describe("Version", func() {
				var testOutputVersion uint
				JustBeforeEach(func() {
					testOutputVersion, err = testkv.MountVersion(testMountName)
				})

				It("should return 1", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())
					By("returning the correct version number")
					Expect(testOutputVersion).To(BeEquivalentTo(1))
				})
			})

			Describe("Set", func() {
				var testSetPath string
				var testSetValues map[string]string
				var testVersionOutput vaultkv.KVVersion
				BeforeEach(func() {
					testSetPath = fmt.Sprintf("%s/boop", testMountName)
				})

				JustBeforeEach(func() {
					testSetValues = map[string]string{"foo": "bar"}
					testVersionOutput, err = testkv.Set(testSetPath, testSetValues, nil)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Getting a version >1", func() {
					JustBeforeEach(func() {
						_, err = testkv.Get(testSetPath, nil, &vaultkv.KVGetOpts{Version: 2})
					})

					It("should return ErrNotFound", func() {
						AssertErrorOfType(&vaultkv.ErrNotFound{})
					})
				})

				When("Overwriting the previously Set key", func() {
					JustBeforeEach(func() {
						testSetValues = map[string]string{"beep": "boop"}
						testVersionOutput, err = testkv.Set(testSetPath, testSetValues, nil)
					})

					It("should overwrite the key without issue", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("returning that the version is one")
						Expect(testVersionOutput.Version).To(BeEquivalentTo(1))
					})
				})

				Describe("Delete with V1Destroy set to false", func() {
					JustBeforeEach(func() {
						err = testkv.Delete(testSetPath, &vaultkv.KVDeleteOpts{
							Versions:  []uint{1},
							V1Destroy: false,
						})
					})

					It("should return ErrKVUnsupported", func() {
						AssertErrorOfType(&vaultkv.ErrKVUnsupported{})
					})
				})

				Describe("Undelete", func() {
					JustBeforeEach(func() {
						err = testkv.Undelete(testSetPath, []uint{1})
					})
					It("should return ErrKVUnsupported", func() {
						AssertErrorOfType(&vaultkv.ErrKVUnsupported{})
					})
				})
			})
		})
	})

	Context("With a KV v2 mount", func() {
		BeforeEach(func() {
			if parseSemver(currentVaultVersion).LessThan(semver{0, 10, 0}) {
				Skip("This version of Vault does not support KVv2")
			} else {
				err = vault.EnableSecretsMount(testMountName, vaultkv.Mount{
					Type:    vaultkv.MountTypeKV,
					Options: vaultkv.KVMountOptions{}.WithVersion(2),
				})
			}

			Expect(err).NotTo(HaveOccurred())
		})

		unityTests()

		Describe("KV v2 specific", func() {
			Describe("isKVv2Mount", func() {
				var mountName string
				var isV2 bool
				JustBeforeEach(func() {
					mountName, isV2, err = vault.IsKVv2Mount(testMountName)
					Expect(err).NotTo(HaveOccurred())
				})

				Specify("should return the mount name and that it is a v2 mount", func() {
					Expect(mountName).To(BeEquivalentTo(testMountName))
					Expect(isV2).To(BeTrue())
				})
			})

			Describe("Version", func() {
				var testOutputVersion uint
				JustBeforeEach(func() {
					testOutputVersion, err = testkv.MountVersion(testMountName)
				})
				It("should return 2", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())
					By("returning the correct version number")
					Expect(testOutputVersion).To(BeEquivalentTo(2))
				})
			})

			Describe("Set", func() {
				var testVersionOutput vaultkv.KVVersion
				var testSetPath string
				var testSetValue map[string]string
				JustBeforeEach(func() {
					testSetPath = fmt.Sprintf("%s/testfield", testMountName)
					testSetValue = map[string]string{"foo": "bar"}
					testVersionOutput, err = testkv.Set(testSetPath, testSetValue, nil)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("setting another version", func() {
					var testVersionOutput2 vaultkv.KVVersion
					var testSetValue2 map[string]string
					JustBeforeEach(func() {
						testSetValue2 = map[string]string{"beep": "boop"}
						testVersionOutput2, err = testkv.Set(testSetPath, testSetValue2, nil)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should have version 2", func() {
						Expect(testVersionOutput2.Version).To(BeEquivalentTo(2))
					})

					Describe("Get", func() {
						var testGetVersionOutput vaultkv.KVVersion
						var testGetValue map[string]string
						var testGetVersion uint

						JustBeforeEach(func() {
							testGetVersionOutput, err = testkv.Get(testSetPath, &testGetValue, &vaultkv.KVGetOpts{Version: testGetVersion})
							Expect(err).NotTo(HaveOccurred())
						})

						Context("getting version 1", func() {
							BeforeEach(func() { testGetVersion = 1 })
							It("should get the first version", func() {
								Expect(testGetVersionOutput.Version).To(Equal(testGetVersion))
								Expect(testGetValue).To(Equal(testSetValue))
							})
						})

						Context("getting version 2", func() {
							BeforeEach(func() { testGetVersion = 2 })
							It("should get the second version", func() {
								Expect(testGetVersionOutput.Version).To(Equal(testGetVersion))
								Expect(testGetValue).To(Equal(testSetValue2))
							})
						})
					})
				})

				Describe("Delete", func() {
					JustBeforeEach(func() {
						err = testkv.Delete(testSetPath, &vaultkv.KVDeleteOpts{Versions: []uint{testVersionOutput.Version}})
						Expect(err).NotTo(HaveOccurred())
					})

					It("should have the version marked as deleted", func() {
						var versionData []vaultkv.KVVersion
						versionData, err = testkv.Versions(testSetPath)
						Expect(versionData).To(HaveLen(1))
						Expect(versionData[0].Deleted).To(BeTrue())
					})
					Describe("Undelete", func() {
						JustBeforeEach(func() {
							err = testkv.Undelete(testSetPath, []uint{testVersionOutput.Version})
						})

						It("should undelete the secret", func() {
							By("not erroring")
							Expect(err).NotTo(HaveOccurred())
							By("having Get fetch the secret that was initially inserted")
							value := map[string]string{}
							version, err := testkv.Get(testSetPath, &value, nil)
							Expect(err).NotTo(HaveOccurred())
							Expect(version.Version).To(Equal(testVersionOutput.Version))
							Expect(value).To(Equal(testSetValue))
						})
					})

				})
			})
		})
	})
})
