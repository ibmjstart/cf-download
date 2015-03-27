package downloader_test

import (
	. "github.com/cf-download/downloader"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Downloader", func() {

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

				err := dir_parser.CheckDownload("/app/node_modules/express/application.js", falseFile, nil)
				Expect(err).To(BeNil())
			})
		})

	})

})
