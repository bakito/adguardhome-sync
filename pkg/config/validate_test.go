package config

import (
	"github.com/go-faker/faker/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/bakito/adguardhome-sync/pkg/types"
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
		It("validate config with all fields randomly populated", func() {
			cfg := &types.Config{}

			err := faker.FakeData(cfg)
			Ω(err).ShouldNot(HaveOccurred())

			data, err := yaml.Marshal(&cfg)
			Ω(err).ShouldNot(HaveOccurred())

			err = validateYAML(data)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("validate config with empty file", func() {
			var data []byte
			err := validateYAML(data)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
