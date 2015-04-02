package main_test

import (
	"os"
	"strings"

	. "github.com/ibmjstart/cf-download"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// unit tests of individual functions
var _ = Describe("CfDownload", func() {
	args := make([]string, 7)

	Describe("Test Flag functionality", func() {

		Context("Check if overWrite flag works", func() {
			It("Should set the overwrite_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--overwrite"

				flagVals := ParseFlags(args)
				Expect(flagVals.MaxRoutines_flag).To(Equal(200))
				Expect(flagVals.OverWrite_flag).To(BeTrue())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if verbose flag works", func() {
			It("Should set the verbose_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--verbose"

				flagVals := ParseFlags(args)
				Expect(flagVals.MaxRoutines_flag).To(Equal(200))
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeTrue())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if Routines flag works", func() {
			It("Should set the maxRoutines_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--routines"
				args[3] = "555"

				flagVals := ParseFlags(args)
				Expect(flagVals.MaxRoutines_flag).To(Equal(555))
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if instance (i) flag works", func() {
			It("Should set the instance_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--i"
				args[3] = "3"

				flagVals := ParseFlags(args)
				Expect(flagVals.MaxRoutines_flag).To(Equal(200))
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("3"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal(""))
			})
		})

		Context("Check if omit flag works", func() {
			It("Should set the omit_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--omit"
				args[3] = "app/node_modules"

				flagVals := ParseFlags(args)
				Expect(flagVals.MaxRoutines_flag).To(Equal(200))
				Expect(flagVals.OverWrite_flag).To(BeFalse())
				Expect(flagVals.Instance_flag).To(Equal("0"))
				Expect(flagVals.Verbose_flag).To(BeFalse())
				Expect(flagVals.Omit_flag).To(Equal("app/node_modules"))
			})
		})

	})

	Describe("Check if current Directory exists", func() {
		Context("when directory exists", func() {
			It("Should return correct strings", func() {
				args[0] = "download"
				args[1] = "app_name"
				args[2] = "app/src/node"
				args[3] = "--verbose"
				currentDirectory, _ := os.Getwd()
				rootWD, startingPath := GetDirectoryContext(currentDirectory, args)

				correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

				Expect(correctSuffix).To(BeTrue())
				Expect(startingPath).To(Equal("/app/src/node/"))
			})
		})
		Context("test target directory parsing", func() {
			It("should still return /app/src/node/ for startingPath", func() {
				args[0] = "download"
				args[1] = "app_name"
				args[2] = "/app/src/node/"
				args[3] = "--verbose"
				currentDirectory, _ := os.Getwd()
				rootWD, startingPath := GetDirectoryContext(currentDirectory, args)

				correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

				Expect(correctSuffix).To(BeTrue())
				Expect(startingPath).To(Equal("/app/src/node/"))
			})
		})
		Context("test target directory parsing", func() {
			It("should still return /app/src/node/ for startingPath", func() {
				args[0] = "download"
				args[1] = "app_name"
				args[2] = "app/src/node/"
				args[3] = "--verbose"
				currentDirectory, _ := os.Getwd()
				rootWD, startingPath := GetDirectoryContext(currentDirectory, args)

				correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

				Expect(correctSuffix).To(BeTrue())
				Expect(startingPath).To(Equal("/app/src/node/"))
			})
		})
		Context("test target directory parsing", func() {
			It("should still return /app/src/node/ for startingPath", func() {
				args[0] = "download"
				args[1] = "app_name"
				args[2] = "/app/src/node"
				args[3] = "--verbose"
				currentDirectory, _ := os.Getwd()
				rootWD, startingPath := GetDirectoryContext(currentDirectory, args)

				correctSuffix := strings.HasSuffix(rootWD, "/cf-download/app-download/app/src/node/")

				Expect(correctSuffix).To(BeTrue())
				Expect(startingPath).To(Equal("/app/src/node/"))
			})
		})
	})

})
