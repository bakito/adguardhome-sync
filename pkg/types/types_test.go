package types_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/google/uuid"
)

var _ = Describe("Types", func() {
	var (
		url     string
		apiPath string
	)
	BeforeEach(func() {
		url = "http://" + uuid.NewString()
		apiPath = "/" + uuid.NewString()
	})

	Context("FilteringStatus", func() {
		It("should correctly parse json", func() {
			b, err := ioutil.ReadFile("../..//testdata/filtering-status.json")
			fs := &types.FilteringStatus{}
			Ω(err).ShouldNot(HaveOccurred())
			err = json.Unmarshal(b, fs)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
	Context("AdGuardInstance", func() {
		It("should build a key with url and api apiPath", func() {
			i := &types.AdGuardInstance{URL: url, APIPath: apiPath}
			Ω(i.Key()).Should(Equal(url + "#" + apiPath))
		})
	})
	Context("RewriteEntry", func() {
		It("should build a key with url and api apiPath", func() {
			domain := uuid.NewString()
			answer := uuid.NewString()
			re := &types.RewriteEntry{Domain: domain, Answer: answer}
			Ω(re.Key()).Should(Equal(domain + "#" + answer))
		})
	})
	Context("UserRules", func() {
		It("should join the rules correctly", func() {
			r1 := uuid.NewString()
			r2 := uuid.NewString()
			ur := types.UserRules([]string{r1, r2})
			Ω(ur.String()).Should(Equal(r1 + "\n" + r2))
		})
	})
	Context("Config", func() {
		var (
			cfg *types.Config
		)
		BeforeEach(func() {
			cfg = &types.Config{}
		})
		Context("UniqueReplicas", func() {
			It("should be empty if noting defined", func() {
				r := cfg.UniqueReplicas()
				Ω(r).Should(BeEmpty())
			})
			It("should be empty if replica url is not set", func() {
				cfg.Replica = types.AdGuardInstance{URL: ""}
				r := cfg.UniqueReplicas()
				Ω(r).Should(BeEmpty())
			})
			It("should be empty if replicas url is not set", func() {
				cfg.Replicas = []types.AdGuardInstance{{URL: ""}}
				r := cfg.UniqueReplicas()
				Ω(r).Should(BeEmpty())
			})
			It("should return only one replica if same url and apiPath", func() {
				cfg.Replica = types.AdGuardInstance{URL: url, APIPath: apiPath}
				cfg.Replicas = []types.AdGuardInstance{{URL: url, APIPath: apiPath}, {URL: url, APIPath: apiPath}}
				r := cfg.UniqueReplicas()
				Ω(r).Should(HaveLen(1))
			})
			It("should return 3 one replicas if urls are different", func() {
				cfg.Replica = types.AdGuardInstance{URL: url, APIPath: apiPath}
				cfg.Replicas = []types.AdGuardInstance{{URL: url + "1", APIPath: apiPath}, {URL: url, APIPath: apiPath + "1"}}
				r := cfg.UniqueReplicas()
				Ω(r).Should(HaveLen(3))
			})
			It("should set default api apiPath if not set", func() {
				cfg.Replica = types.AdGuardInstance{URL: url}
				cfg.Replicas = []types.AdGuardInstance{{URL: url + "1"}}
				r := cfg.UniqueReplicas()
				Ω(r).Should(HaveLen(2))
				Ω(r[0].APIPath).Should(Equal(types.DefaultAPIPath))
				Ω(r[1].APIPath).Should(Equal(types.DefaultAPIPath))
			})
		})
	})
})
