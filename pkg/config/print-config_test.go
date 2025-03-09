package config

import (
	"fmt"
	"runtime"

	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppConfig", func() {
	var (
		ac  *AppConfig
		env []string
	)
	BeforeEach(func() {
		ac = &AppConfig{
			cfg: &types.Config{
				Origin: types.AdGuardInstance{
					URL: "https://ha.xxxx.net:3000",
				},
			},
			content: `
origin:
  url: https://ha.xxxx.net:3000
`,
		}
		env = []string{"FOO=foo", "BAR=bar"}
	})
	Context("print", func() {
		It("should print config without file", func() {
			out, err := ac.print(env, "v0.0.1", []string{"v0.0.2"})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out).Should(Equal(fmt.Sprintf(expected1, version.Version, version.Build, runtime.GOOS, runtime.GOARCH)))
		})
		It("should print config with file", func() {
			ac.filePath = "config.yaml"
			out, err := ac.print(env, "v0.0.1", []string{"v0.0.2"})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out).Should(Equal(fmt.Sprintf(expected2, version.Version, version.Build, runtime.GOOS, runtime.GOARCH)))
		})
	})
})

const (
	expected1 = "<!-- PLEASE COPY THE FOLLOWING OUTPUT AS IS INTO THE GITHUB ISSUE (Don't forget to mask your usernames, passwords, IPs and other sensitive information when using this in an issue ) -->\n\n### Runtime\n\nAdguardHome-Sync Version: %s\nBuild: %s\nOperatingSystem: %s\nArchitecture: %s\nOriginVersion: v0.0.1\nReplicaVersions:\n- Replica 1: v0.0.2\n\n### AdGuardHome sync aggregated config\n\n```yaml\norigin:\n    url: https://ha.xxxx.net:3000\n    webURL: \"\"\n    insecureSkipVerify: false\n    autoSetup: false\n\n```\n\n### Environment Variables\n\n```ini\nBAR=bar\nFOO=foo\n```\n\n<!-- END OF GITHUB ISSUE CONTENT -->"
	expected2 = "<!-- PLEASE COPY THE FOLLOWING OUTPUT AS IS INTO THE GITHUB ISSUE (Don't forget to mask your usernames, passwords, IPs and other sensitive information when using this in an issue ) -->\n\n### Runtime\n\nAdguardHome-Sync Version: %s\nBuild: %s\nOperatingSystem: %s\nArchitecture: %s\nOriginVersion: v0.0.1\nReplicaVersions:\n- Replica 1: v0.0.2\n\n### AdGuardHome sync aggregated config\n\n```yaml\norigin:\n    url: https://ha.xxxx.net:3000\n    webURL: \"\"\n    insecureSkipVerify: false\n    autoSetup: false\n\n```\n### AdGuardHome sync unmodified config file\n\nConfig file path: config.yaml\n\n```yaml\n\norigin:\n  url: https://ha.xxxx.net:3000\n\n```\n\n### Environment Variables\n\n```ini\nBAR=bar\nFOO=foo\n```\n\n<!-- END OF GITHUB ISSUE CONTENT -->"
)
