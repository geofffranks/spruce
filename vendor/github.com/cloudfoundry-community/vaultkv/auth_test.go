package vaultkv_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth", func() {
	BeforeEach(func() {
		InitAndUnsealVault()
	})

	Describe("TokenIsValid", func() {
		JustBeforeEach(func() {
			err = vault.TokenIsValid()
		})

		Context("When the token is valid", func() {
			It("should not err", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When the token is invalid", func() {
			BeforeEach(func() {
				if vault.AuthToken[0] == 'a' {
					vault.AuthToken = "b" + vault.AuthToken[1:]
				} else {
					vault.AuthToken = "a" + vault.AuthToken[1:]
				}
			})
			It("should err", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When the token not properly formatted", func() {
			BeforeEach(func() {
				vault.AuthToken = vault.AuthToken[1:]
			})
			It("should err", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When there is no token", func() {
			BeforeEach(func() {
				vault.AuthToken = ""
			})
			It("should err", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
