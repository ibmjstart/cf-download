package main_test

import (
	"errors"
	"os"
	"strings"

	. "github.com/cf-download"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CfDownload", func() {
	args := make([]string, 7)

	Describe("Test Flag functionality", func() {

		Context("Check if overWrite flag works", func() {
			It("Should set the overwritep_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--overwrite"

				_, _, flagVals := ParseFlags(args)
				Expect(*flagVals.MaxRoutinesp_flag).To(Equal(200))
				Expect(*flagVals.OverWritep_flag).To(BeTrue())
				Expect(*flagVals.Instancep_flag).To(Equal(0))
				Expect(*flagVals.Verbosep_flag).To(BeFalse())
				Expect(*flagVals.Omitp_flag).To(Equal(""))
			})
		})

		Context("Check if verbose flag works", func() {
			It("Should set the verbosep_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--verbose"

				_, _, flagVals := ParseFlags(args)
				Expect(*flagVals.MaxRoutinesp_flag).To(Equal(200))
				Expect(*flagVals.OverWritep_flag).To(BeFalse())
				Expect(*flagVals.Instancep_flag).To(Equal(0))
				Expect(*flagVals.Verbosep_flag).To(BeTrue())
				Expect(*flagVals.Omitp_flag).To(Equal(""))
			})
		})

		Context("Check if Routines flag works", func() {
			It("Should set the maxRoutinesp_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--routines"
				args[3] = "555"

				_, _, flagVals := ParseFlags(args)
				Expect(*flagVals.MaxRoutinesp_flag).To(Equal(555))
				Expect(*flagVals.OverWritep_flag).To(BeFalse())
				Expect(*flagVals.Instancep_flag).To(Equal(0))
				Expect(*flagVals.Verbosep_flag).To(BeFalse())
				Expect(*flagVals.Omitp_flag).To(Equal(""))
			})
		})

		Context("Check if instance (i) flag works", func() {
			It("Should set the instancep_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--i"
				args[3] = "3"

				_, _, flagVals := ParseFlags(args)
				Expect(*flagVals.MaxRoutinesp_flag).To(Equal(200))
				Expect(*flagVals.OverWritep_flag).To(BeFalse())
				Expect(*flagVals.Instancep_flag).To(Equal(3))
				Expect(*flagVals.Verbosep_flag).To(BeFalse())
				Expect(*flagVals.Omitp_flag).To(Equal(""))
			})
		})

		Context("Check if omit flag works", func() {
			It("Should set the omitp_flag", func() {
				args[0] = "download"
				args[1] = "app"
				args[2] = "--omit"
				args[3] = "app/node_modules"

				_, _, flagVals := ParseFlags(args)
				Expect(*flagVals.MaxRoutinesp_flag).To(Equal(200))
				Expect(*flagVals.OverWritep_flag).To(BeFalse())
				Expect(*flagVals.Instancep_flag).To(Equal(0))
				Expect(*flagVals.Verbosep_flag).To(BeFalse())
				Expect(*flagVals.Omitp_flag).To(Equal("app/node_modules"))
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
	})

	Describe("Test checkDownload Function", func() {
		Context("when we recieve server error", func() {
			It("Should return server error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "FAILED"

				err := CheckDownload("/app/node_modules/express/application.js", falseFile, nil)
				Expect(err).To(Equal(errors.New("download failed")))
			})
		})

		Context("when we recieve no error", func() {
			It("Should return no error", func() {
				falseFile := make([]string, 3)
				falseFile[0] = "Getting files for app app_name in org org_name / space spacey as user@us.ibm.com"
				falseFile[1] = "OK"

				err := CheckDownload("/app/node_modules/express/application.js", falseFile, nil)
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("Test getDirectoryContext", func() {
		Context("when directory exists", func() {
			It("Should be true", func() {
				currentDirectory, _ := os.Getwd()
				Expect(Exists(currentDirectory)).To(BeTrue())
			})
		})
	})

})
