package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/version"
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
	Context("printInternal", func() {
		It("should printInternal config without file", func() {
			out, err := ac.printInternal(env, "v0.0.1", []string{"v0.0.2"})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out).Should(Equal(fmt.Sprintf(expected(1), version.Version, version.Build, runtime.GOOS, runtime.GOARCH)))
		})
		It("should printInternal config with file", func() {
			ac.filePath = "config.yaml"
			out, err := ac.printInternal(env, "v0.0.1", []string{"v0.0.2"})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out).Should(Equal(fmt.Sprintf(expected(2), version.Version, version.Build, runtime.GOOS, runtime.GOARCH)))
		})
	})
})

func expected(id int) string {
	b, err := os.ReadFile(
		filepath.Join("..", "..", "testdata", "config", fmt.Sprintf("print-config_test_expected%d.md", id)),
	)
	Ω(err).ShouldNot(HaveOccurred())
	return normalizeLineEndings(string(b))
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
