package dir_parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDirParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DirParser Suite")
}
