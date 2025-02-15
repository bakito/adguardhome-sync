package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("validateSchema", func() {
		DescribeTable("validateSchema config",
			func(configFile string, expectFail bool) {
				err := validateSchema(configFile)
				if expectFail {
					Ω(err).Should(HaveOccurred())
				} else {
					Ω(err).ShouldNot(HaveOccurred())
				}
			},
			Entry(`Should be valid`, "../../testdata/config/config-valid.yaml", false),
			Entry(`Should be valid if file doesn't exist`, "../../testdata/config/foo.bar", false),
			Entry(`Should fail if file is not yaml`, "../../go.mod", true),
		)
	})
})
