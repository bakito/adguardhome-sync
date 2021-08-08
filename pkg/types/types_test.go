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
		url = "https://" + uuid.NewString()
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

	Context("Filters", func() {
		Context("Merge", func() {
			var (
				originFilters  types.Filters
				replicaFilters types.Filters
			)
			BeforeEach(func() {
				originFilters = types.Filters{}
				replicaFilters = types.Filters{}
			})

			It("should add a missing filter", func() {
				originFilters = append(originFilters, types.Filter{URL: url})
				a, u, d := replicaFilters.Merge(originFilters)
				Ω(a).Should(HaveLen(1))
				Ω(u).Should(BeEmpty())
				Ω(d).Should(BeEmpty())

				Ω(a[0].URL).Should(Equal(url))
			})

			It("should remove additional filter", func() {
				replicaFilters = append(replicaFilters, types.Filter{URL: url})
				a, u, d := replicaFilters.Merge(originFilters)
				Ω(a).Should(BeEmpty())
				Ω(u).Should(BeEmpty())
				Ω(d).Should(HaveLen(1))

				Ω(d[0].URL).Should(Equal(url))
			})

			It("should update existing filter when enabled differs", func() {
				enabled := true
				originFilters = append(originFilters, types.Filter{URL: url, Enabled: enabled})
				replicaFilters = append(replicaFilters, types.Filter{URL: url, Enabled: !enabled})
				a, u, d := replicaFilters.Merge(originFilters)
				Ω(a).Should(BeEmpty())
				Ω(u).Should(HaveLen(1))
				Ω(d).Should(BeEmpty())

				Ω(u[0].Enabled).Should(Equal(enabled))
			})

			It("should update existing filter when name differs", func() {
				name1 := uuid.NewString()
				name2 := uuid.NewString()
				originFilters = append(originFilters, types.Filter{URL: url, Name: name1})
				replicaFilters = append(replicaFilters, types.Filter{URL: url, Name: name2})
				a, u, d := replicaFilters.Merge(originFilters)
				Ω(a).Should(BeEmpty())
				Ω(u).Should(HaveLen(1))
				Ω(d).Should(BeEmpty())

				Ω(u[0].Name).Should(Equal(name1))
			})

			It("should have no changes", func() {
				originFilters = append(originFilters, types.Filter{URL: url})
				replicaFilters = append(replicaFilters, types.Filter{URL: url})
				a, u, d := replicaFilters.Merge(originFilters)
				Ω(a).Should(BeEmpty())
				Ω(u).Should(BeEmpty())
				Ω(d).Should(BeEmpty())
			})
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
	Context("QueryLogConfig", func() {
		Context("Equal", func() {
			var (
				a *types.QueryLogConfig
				b *types.QueryLogConfig
			)
			BeforeEach(func() {
				a = &types.QueryLogConfig{}
				b = &types.QueryLogConfig{}
			})
			It("should be equal", func() {
				a.Enabled = true
				a.Interval = 1
				a.AnonymizeClientIP = true
				b.Enabled = true
				b.Interval = 1
				b.AnonymizeClientIP = true
				Ω(a.Equals(b)).Should(BeTrue())
			})
			It("should not be equal when enabled differs", func() {
				a.Enabled = true
				b.Enabled = false
				Ω(a.Equals(b)).ShouldNot(BeTrue())
			})
			It("should not be equal when interval differs", func() {
				a.Interval = 1
				b.Interval = 2
				Ω(a.Equals(b)).ShouldNot(BeTrue())
			})
			It("should not be equal when anonymizeClientIP differs", func() {
				a.AnonymizeClientIP = true
				b.AnonymizeClientIP = false
				Ω(a.Equals(b)).ShouldNot(BeTrue())
			})
		})
	})
	Context("RewriteEntries", func() {
		Context("Merge", func() {
			var (
				originRE  types.RewriteEntries
				replicaRE types.RewriteEntries
				domain    string
			)
			BeforeEach(func() {
				originRE = types.RewriteEntries{}
				replicaRE = types.RewriteEntries{}
				domain = uuid.NewString()
			})

			It("should add a missing rewrite entry", func() {
				originRE = append(originRE, types.RewriteEntry{Domain: domain})
				a, r, d := replicaRE.Merge(&originRE)
				Ω(a).Should(HaveLen(1))
				Ω(r).Should(BeEmpty())
				Ω(d).Should(BeEmpty())

				Ω(a[0].Domain).Should(Equal(domain))
			})

			It("should remove additional ewrite entry", func() {
				replicaRE = append(replicaRE, types.RewriteEntry{Domain: domain})
				a, r, d := replicaRE.Merge(&originRE)
				Ω(a).Should(BeEmpty())
				Ω(r).Should(HaveLen(1))
				Ω(d).Should(BeEmpty())

				Ω(r[0].Domain).Should(Equal(domain))
			})

			It("should have no changes", func() {
				originRE = append(originRE, types.RewriteEntry{Domain: domain})
				replicaRE = append(replicaRE, types.RewriteEntry{Domain: domain})
				a, r, d := replicaRE.Merge(&originRE)
				Ω(a).Should(BeEmpty())
				Ω(r).Should(BeEmpty())
				Ω(d).Should(BeEmpty())
			})

			It("should remove target duplicate", func() {
				originRE = append(originRE, types.RewriteEntry{Domain: domain})
				replicaRE = append(replicaRE, types.RewriteEntry{Domain: domain})
				replicaRE = append(replicaRE, types.RewriteEntry{Domain: domain})
				a, r, d := replicaRE.Merge(&originRE)
				Ω(a).Should(BeEmpty())
				Ω(r).Should(HaveLen(1))
				Ω(d).Should(BeEmpty())
			})

			It("should remove target duplicate", func() {
				originRE = append(originRE, types.RewriteEntry{Domain: domain})
				originRE = append(originRE, types.RewriteEntry{Domain: domain})
				replicaRE = append(replicaRE, types.RewriteEntry{Domain: domain})
				a, r, d := replicaRE.Merge(&originRE)
				Ω(a).Should(BeEmpty())
				Ω(r).Should(BeEmpty())
				Ω(d).Should(HaveLen(1))
			})
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

	Context("Clients", func() {
		Context("Merge", func() {
			var (
				originClients  *types.Clients
				replicaClients types.Clients
				name           string
			)
			BeforeEach(func() {
				originClients = &types.Clients{}
				replicaClients = types.Clients{}
				name = uuid.NewString()
			})

			It("should add a missing client", func() {
				originClients.Clients = append(originClients.Clients, types.Client{Name: name})
				a, u, d := replicaClients.Merge(originClients)
				Ω(a).Should(HaveLen(1))
				Ω(u).Should(BeEmpty())
				Ω(d).Should(BeEmpty())

				Ω(a[0].Name).Should(Equal(name))
			})

			It("should remove additional client", func() {
				replicaClients.Clients = append(replicaClients.Clients, types.Client{Name: name})
				a, u, d := replicaClients.Merge(originClients)
				Ω(a).Should(BeEmpty())
				Ω(u).Should(BeEmpty())
				Ω(d).Should(HaveLen(1))

				Ω(d[0].Name).Should(Equal(name))
			})

			It("should update existing client when name differs", func() {
				disallowed := true
				originClients.Clients = append(originClients.Clients, types.Client{Name: name, Disallowed: disallowed})
				replicaClients.Clients = append(replicaClients.Clients, types.Client{Name: name, Disallowed: !disallowed})
				a, u, d := replicaClients.Merge(originClients)
				Ω(a).Should(BeEmpty())
				Ω(u).Should(HaveLen(1))
				Ω(d).Should(BeEmpty())

				Ω(u[0].Disallowed).Should(Equal(disallowed))
			})
		})
	})
	Context("Services", func() {
		Context("Equals", func() {
			It("should be equal", func() {
				s1 := types.Services([]string{"a", "b"})
				s2 := types.Services([]string{"b", "a"})
				Ω(s1.Equals(s2)).Should(BeTrue())
			})
			It("should not be equal different values", func() {
				s1 := types.Services([]string{"a", "b"})
				s2 := types.Services([]string{"B", "a"})
				Ω(s1.Equals(s2)).ShouldNot(BeTrue())
			})
			It("should not be equal different length", func() {
				s1 := types.Services([]string{"a", "b"})
				s2 := types.Services([]string{"b", "a", "c"})
				Ω(s1.Equals(s2)).ShouldNot(BeTrue())
			})
		})
	})
	Context("DNSConfig", func() {
		Context("Equals", func() {
			It("should be equal", func() {
				dc1 := &types.DNSConfig{Upstreams: []string{"a"}}
				dc2 := &types.DNSConfig{Upstreams: []string{"a"}}
				Ω(dc1.Equals(dc2)).Should(BeTrue())
			})
			It("should not be equal", func() {
				dc1 := &types.DNSConfig{Upstreams: []string{"a"}}
				dc2 := &types.DNSConfig{Upstreams: []string{"b"}}
				Ω(dc1.Equals(dc2)).ShouldNot(BeTrue())
			})
		})
	})
})
