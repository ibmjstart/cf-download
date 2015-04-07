package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfDownload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfDownload Suite")
}
