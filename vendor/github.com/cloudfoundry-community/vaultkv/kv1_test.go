package vaultkv_test

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudfoundry-community/vaultkv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KVv1", func() {
	//This is a hack because I don't want to refactor _everything_ in these tests
	var getAssertionPath string

	var AssertGetEquals = func(expected map[string]string) func() {
		return func() {
			output := make(map[string]string)
			err = vault.Get(getAssertionPath, &output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(expected))
		}
	}

	var AssertExists = func(exists bool) func() {
		var fn func()
		if exists {
			fn = func() {
				err = vault.Get(getAssertionPath, nil)
				Expect(err).NotTo(HaveOccurred())
			}
		} else {
			fn = func() {
				err = vault.Get(getAssertionPath, nil)
				AssertErrorOfType(&vaultkv.ErrNotFound{})
			}
		}

		return fn
	}

	BeforeEach(func() {
		InitAndUnsealVault()
	})

	Describe("Setting something in the Vault", func() {
		var testPath string
		var testValue map[string]string
		BeforeEach(func() {
			testPath = "secret/foo"
			testValue = map[string]string{
				"foo":  "bar",
				"beep": "boop",
			}
		})

		JustBeforeEach(func() {
			err = vault.Set(testPath, testValue)
		})

		When("the value is nil", func() {
			BeforeEach(func() {
				testValue = nil
				getAssertionPath = testPath
			})

			It("should err properly", func() {
				By("returning ErrBadRequest")
				AssertErrorOfType(&vaultkv.ErrBadRequest{})()

				By("having Get find nothing at this path after the call")
				AssertExists(false)()
			})
		})

		When("the path is doesn't correspond to a mounted backend", func() {
			BeforeEach(func() {
				testPath = "notabackend/foo"
			})

			It("should err properly", func() {
				By("returning ErrNotFound")
				AssertErrorOfType(&vaultkv.ErrNotFound{})()

				By("having Get find nothing at this path after the call")
				AssertExists(false)()
			})
		})

		When("the path has a leading slash", func() {
			BeforeEach(func() {
				testPath = "/secret/foo"
			})

			It("should get inserted properly", func() {
				By("not erroring")
				Expect(err).NotTo(HaveOccurred())

				By("having get find the key at the path without a slash")
				getAssertionPath = strings.TrimPrefix(testPath, "/")
				AssertExists(true)()

				By("having get find the key at the path without a slash")
				AssertGetEquals(map[string]string{"foo": "bar", "beep": "boop"})()

				By("having get find the key at the path with a slash")
				getAssertionPath = testPath
				AssertExists(true)()

				By("having get find the key at the path with a slash")
				AssertGetEquals(map[string]string{"foo": "bar", "beep": "boop"})()
			})
		})

		When("the path has a trailing slash", func() {
			BeforeEach(func() {
				testPath = "secret/foo/"
			})

			It("should get inserted properly", func() {
				By("not erroring")
				Expect(err).NotTo(HaveOccurred())

				By("having get find the key at the path without a slash")
				getAssertionPath = strings.TrimSuffix(testPath, "/")
				AssertExists(true)()

				By("having get find the key at the path without a slash")
				AssertGetEquals(map[string]string{"foo": "bar", "beep": "boop"})()

				By("having get find the key at the path with a slash")
				getAssertionPath = testPath
				AssertExists(true)()

				By("having get find the key at the path with a slash")
				AssertGetEquals(map[string]string{"foo": "bar", "beep": "boop"})()
			})
		})

		When("setting an already set key", func() {
			var secondTestValue map[string]string
			BeforeEach(func() {
				secondTestValue = map[string]string{
					"thisisanotherkey": "thisisanothervalue",
				}
				getAssertionPath = testPath
			})

			JustBeforeEach(func() {
				err = vault.Set(testPath, secondTestValue)
			})

			It("should overwrite the value", func() {
				By("not erroring")
				Expect(err).NotTo(HaveOccurred())

				By("having Get find the value that was added second")
				AssertGetEquals(map[string]string{"thisisanotherkey": "thisisanothervalue"})
			})
		})

		Describe("Get", func() {
			var getTestPath string
			var getOutputValue map[string]string

			BeforeEach(func() {
				getOutputValue = make(map[string]string)
			})

			JustBeforeEach(func() {
				err = vault.Get(getTestPath, &getOutputValue)
			})

			When("the key exists", func() {
				BeforeEach(func() {
					getTestPath = testPath
				})

				It("should retrieve the key", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("returning the value that was inserted")
					Expect(getOutputValue).To(Equal(testValue))
				})
			})

			When("the key doesn't exist", func() {
				BeforeEach(func() {
					getTestPath = fmt.Sprintf("%sabcd", testPath)
				})

				It("should return ErrNotFound", AssertErrorOfType(&vaultkv.ErrNotFound{}))
			})
		})

		Describe("Delete", func() {
			var deleteTestPath string
			JustBeforeEach(func() {
				err = vault.Delete(deleteTestPath)
			})

			When("the key exists", func() {
				BeforeEach(func() {
					deleteTestPath = testPath
					getAssertionPath = testPath
				})

				It("should delete the key", func() {
					By("not erroring")
					Expect(err).NotTo(HaveOccurred())

					By("get not finding the deleted key")
					AssertExists(false)()
				})
			})

			When("the key doesn't exist", func() {
				BeforeEach(func() {
					deleteTestPath = fmt.Sprintf("%sabcd", testPath)
				})

				It("should not return an error", func() { Expect(err).NotTo(HaveOccurred()) })
			})
		})

		Describe("Adding another key with multiple parts", func() {
			var secondTestPath string
			var secondTestValue map[string]string
			BeforeEach(func() {
				secondTestPath = "secret/beep/bar"
				secondTestValue = map[string]string{
					"werealljustlittlebabybirds": "peckingourwayoutofourshells",
				}
			})

			JustBeforeEach(func() {
				err = vault.Set(secondTestPath, secondTestValue)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("List", func() {
				var listTestPath string
				var listTestOutput []string

				JustBeforeEach(func() {
					listTestOutput, err = vault.List(listTestPath)
				})

				//Order doesn't matter
				var AssertListEquals = func(expected []string) func() {
					return func() {
						Expect(listTestOutput).ToNot(BeNil())
						Expect(expected).ToNot(BeNil())
						sort.Strings(expected)
						sort.Strings(listTestOutput)
						Expect(listTestOutput).To(Equal(expected))
					}
				}

				Context("on `secret'", func() {
					BeforeEach(func() {
						listTestPath = "secret"
					})

					It("should return the correct paths", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("returning the correct list of paths")
						AssertListEquals([]string{"foo", "beep/"})
					})
				})

				Context("on the dir of the nested key", func() {
					BeforeEach(func() {
						listTestPath = "secret/beep"
					})

					It("should return the correct paths", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("returning the correct list of paths")
						AssertListEquals([]string{"bar"})
					})
				})

				When("the path doesn't exist", func() {
					BeforeEach(func() {
						listTestPath = "secret/boo/hiss"
					})

					It("should return an ErrNotFound", AssertErrorOfType(&vaultkv.ErrNotFound{}))
				})

				When("the path is a secret, not a folder", func() {
					BeforeEach(func() {
						listTestPath = "secret/foo"
					})

					It("should return an ErrNotFound", AssertErrorOfType(&vaultkv.ErrNotFound{}))
				})
			})

			Describe("Get", func() {
				var getTestPath string
				var getOutputValue map[string]string

				BeforeEach(func() {
					getOutputValue = make(map[string]string)
				})

				JustBeforeEach(func() {
					err = vault.Get(getTestPath, &getOutputValue)
				})
				Context("on the nested key", func() {
					BeforeEach(func() {
						getTestPath = secondTestPath
					})

					It("should retrieve the value", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("returning the value that was inserted")
						Expect(getOutputValue).To(Equal(secondTestValue))
					})
				})
			})

			Describe("Delete", func() {
				var deleteTestPath string
				JustBeforeEach(func() {
					err = vault.Delete(deleteTestPath)
				})
				Context("on the nested key", func() {
					BeforeEach(func() {
						deleteTestPath = secondTestPath
						getAssertionPath = deleteTestPath
					})

					It("should delete the key", func() {
						By("not erroring")
						Expect(err).NotTo(HaveOccurred())

						By("having Get be unable to find the key")
						AssertExists(false)()
					})
				})
			})
		})
	})
})
